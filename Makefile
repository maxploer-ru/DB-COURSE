.PHONY: test test-coverage test-allure mocks clean

mocks:
	mockery --all --dir=./internal --output=./internal/testing/mocks --outpkg=mocks

test:
	go test -v -shuffle=on ./...

test-coverage:
	go test -shuffle=on -coverprofile=coverage.out -covermode=atomic ./internal/...
	awk -i inplace '!/(testing|delivery|logger|config|cache|mappers)/' coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-allure:
	mkdir allure-results
	go test -shuffle=on -v ./... | go-junit-report > allure-results/report.xml
	allure generate allure-results --clean -o allure-report
	allure open allure-report


clean:
	rm -rf coverage.out allure-results allure-report report.xml