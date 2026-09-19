package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type testEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

type testResult struct {
	Package string
	Name    string
	Status  string
	Start   time.Time
	Stop    time.Time
	Output  strings.Builder
}

type junitReport struct {
	XMLName xml.Name     `xml:"testsuites"`
	Suites  []junitSuite `xml:"testsuite"`
}

type junitSuite struct {
	XMLName  xml.Name    `xml:"testsuite"`
	Name     string      `xml:"name,attr"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Skipped  int         `xml:"skipped,attr"`
	Time     string      `xml:"time,attr"`
	Cases    []junitCase `xml:"testcase"`
}

type junitCase struct {
	XMLName   xml.Name    `xml:"testcase"`
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Time      string      `xml:"time,attr"`
	Failure   *junitIssue `xml:"failure,omitempty"`
	Skipped   *junitIssue `xml:"skipped,omitempty"`
	SystemOut string      `xml:"system-out,omitempty"`
}

type junitIssue struct {
	Message string `xml:"message,attr,omitempty"`
	Body    string `xml:",chardata"`
}

type allureResult struct {
	UUID       string        `json:"uuid"`
	HistoryID  string        `json:"historyId"`
	Name       string        `json:"name"`
	FullName   string        `json:"fullName"`
	Status     string        `json:"status"`
	Stage      string        `json:"stage"`
	Start      int64         `json:"start"`
	Stop       int64         `json:"stop"`
	StatusInfo allureStatus  `json:"statusDetails"`
	Labels     []allureLabel `json:"labels"`
}

type allureStatus struct {
	Message string `json:"message,omitempty"`
	Trace   string `json:"trace,omitempty"`
}

type allureLabel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func main() {
	input := flag.String("input", "test-results.json", "go test -json output")
	output := flag.String("output", "allure-results", "report output directory")
	flag.Parse()

	results, err := readResults(*input)
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll(*output, 0o755); err != nil {
		fail(err)
	}
	if err := writeJUnit(*output, results); err != nil {
		fail(err)
	}
	if err := writeAllure(*output, results); err != nil {
		fail(err)
	}
}

func readResults(path string) ([]*testResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open test events: %w", err)
	}
	defer file.Close()

	byID := make(map[string]*testResult)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var event testEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if event.Test == "" {
			continue
		}

		id := event.Package + "\x00" + event.Test
		result := byID[id]
		if result == nil {
			result = &testResult{Package: event.Package, Name: event.Test, Status: "broken"}
			byID[id] = result
		}
		if !event.Time.IsZero() && result.Start.IsZero() {
			result.Start = event.Time
		}
		switch event.Action {
		case "run":
			result.Status = "broken"
		case "pass":
			result.Status = "passed"
		case "fail":
			result.Status = "failed"
		case "skip":
			result.Status = "skipped"
		}
		if event.Output != "" {
			result.Output.WriteString(event.Output)
		}
		if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			result.Stop = event.Time
			if result.Stop.IsZero() && !result.Start.IsZero() {
				result.Stop = result.Start.Add(time.Duration(event.Elapsed * float64(time.Second)))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read test events: %w", err)
	}

	results := make([]*testResult, 0, len(byID))
	for _, result := range byID {
		if result.Start.IsZero() {
			result.Start = time.Now()
		}
		if result.Stop.IsZero() {
			result.Stop = result.Start
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Package == results[j].Package {
			return results[i].Name < results[j].Name
		}
		return results[i].Package < results[j].Package
	})
	return results, nil
}

func writeJUnit(directory string, results []*testResult) error {
	byPackage := make(map[string][]*testResult)
	for _, result := range results {
		byPackage[result.Package] = append(byPackage[result.Package], result)
	}

	packages := make([]string, 0, len(byPackage))
	for packageName := range byPackage {
		packages = append(packages, packageName)
	}
	sort.Strings(packages)

	report := junitReport{}
	for _, packageName := range packages {
		resultsForPackage := byPackage[packageName]
		suite := junitSuite{Name: packageName, Tests: len(resultsForPackage)}
		var suiteSeconds float64
		for _, result := range resultsForPackage {
			elapsed := result.Stop.Sub(result.Start).Seconds()
			if elapsed < 0 {
				elapsed = 0
			}
			caseResult := junitCase{
				Classname: packageName,
				Name:      result.Name,
				Time:      fmt.Sprintf("%.3f", elapsed),
				SystemOut: result.Output.String(),
			}
			switch result.Status {
			case "failed", "broken":
				suite.Failures++
				caseResult.Failure = &junitIssue{Message: "test failed", Body: result.Output.String()}
			case "skipped":
				suite.Skipped++
				caseResult.Skipped = &junitIssue{Message: "test skipped"}
			}
			suiteSeconds += elapsed
			suite.Cases = append(suite.Cases, caseResult)
		}
		suite.Time = fmt.Sprintf("%.3f", suiteSeconds)
		report.Suites = append(report.Suites, suite)
	}

	data, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JUnit report: %w", err)
	}
	data = append([]byte(xml.Header), data...)
	return os.WriteFile(filepath.Join(directory, "junit.xml"), data, 0o644)
}

func writeAllure(directory string, results []*testResult) error {
	for _, result := range results {
		fullName := result.Package + "/" + result.Name
		hash := sha256.Sum256([]byte(fullName))
		id := hex.EncodeToString(hash[:])
		allure := allureResult{
			UUID:      id,
			HistoryID: id,
			Name:      result.Name,
			FullName:  fullName,
			Status:    result.Status,
			Stage:     "finished",
			Start:     result.Start.UnixMilli(),
			Stop:      result.Stop.UnixMilli(),
			Labels:    []allureLabel{{Name: "suite", Value: result.Package}},
		}
		if result.Status == "failed" || result.Status == "broken" {
			allure.StatusInfo = allureStatus{Message: "test failed", Trace: result.Output.String()}
		}
		data, err := json.MarshalIndent(allure, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal Allure result: %w", err)
		}
		if err := os.WriteFile(filepath.Join(directory, id+"-result.json"), data, 0o644); err != nil {
			return fmt.Errorf("write Allure result: %w", err)
		}
	}
	return nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
