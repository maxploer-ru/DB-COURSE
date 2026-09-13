.PHONY: test test-coverage test-allure mocks clean

mocks:
	mockery --all --dir=./internal --output=./internal/testing/mocks --outpkg=mocks

test:
	go test -v -shuffle=on ./...

test-coverage:
	go test -shuffle=on -coverprofile=coverage.out -covermode=atomic ./internal/...
	awk -i inplace '!/(testing|delivery|logger|config|cache|mappers)/' coverage.out
	go tool cover -func=coverage.out
	@echo "Для просмотра покрытия по строкам в браузере выполните: go tool cover -html=coverage.out"

test-allure:
	@echo "Очистка старых отчетов..."
	rm -rf allure-results report.xml
	mkdir -p allure-results
	@echo "Запуск тестов и генерация XML..."
	-go test -v -shuffle=on ./internal/... 2>&1 | go-junit-report -set-exit-code > allure-results/report.xml
	@echo "Генерация HTML отчета Allure..."
	allure generate allure-results --clean -o allure-report
	@echo "Открытие отчета..."
	allure open allure-report

clean:
	rm -rf coverage.out allure-results allure-report report.xml