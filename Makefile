.PHONY: test test-unit test-offline test-coverage branch-coverage test-allure test-allure-unit mocks clean

UNIT_PACKAGES := ./internal/service ./internal/infrastructure/auth ./internal/infrastructure/logger ./internal/delivery/middleware
GOBCO_VERSION ?= v1.3.4

mocks:
	mockery --all --dir=./internal --output=./internal/testing/mocks --outpkg=mocks

test:
	go test -v -shuffle=on ./...

test-unit:
	go test -v -shuffle=on $(UNIT_PACKAGES)

test-offline:
	GOPROXY=off GOSUMDB=off go test -v -shuffle=on $(UNIT_PACKAGES)

test-coverage:
	go test -shuffle=on -coverprofile=coverage.out -covermode=atomic -coverpkg=./internal/... $(UNIT_PACKAGES)
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	$(MAKE) branch-coverage

branch-coverage:
	@for package in internal/service internal/infrastructure/auth internal/infrastructure/logger; do \
		echo "== branch coverage: ./$$package =="; \
		(cd "$$package" && go run github.com/rillig/gobco@$(GOBCO_VERSION)); \
	done

test-allure:
	rm -rf allure-results allure-report test-results.json
	set +e; go test -shuffle=on -json ./... > test-results.json; status=$$?; set -e; \
	go run ./scripts/testreport -input test-results.json -output allure-results; \
	if command -v allure >/dev/null 2>&1; then allure generate allure-results --clean -o allure-report; \
	else echo "allure CLI is not installed; allure-results was generated"; fi; \
	exit $$status

test-allure-unit:
	rm -rf allure-results allure-report test-results.json
	set +e; go test -shuffle=on -json $(UNIT_PACKAGES) > test-results.json; status=$$?; set -e; \
	go run ./scripts/testreport -input test-results.json -output allure-results; \
	if command -v allure >/dev/null 2>&1; then allure generate allure-results --clean -o allure-report; \
	else echo "allure CLI is not installed; allure-results was generated"; fi; \
	exit $$status


clean:
	rm -rf coverage.out coverage.html allure-results allure-report test-results.json
