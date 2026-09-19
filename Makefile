.PHONY: test test-all test-unit test-offline test-coverage branch-coverage test-allure test-allure-unit test-allure-report allure-open test-unit-container test-integration test-e2e mocks clean openapi-check openapi-generate graphql-mock rest-mock

UNIT_PACKAGES := ./internal/service ./internal/infrastructure/auth ./internal/infrastructure/logger ./internal/delivery/middleware ./internal/worker
TEST_PACKAGES ?= ./...
GOBCO_VERSION ?= v1.3.4
OAPI_CODEGEN_VERSION ?= v2.4.1

openapi-check:
	npx @redocly/cli@latest lint openapi/openapi.yaml

openapi-generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION) --config openapi/oapi-codegen.yaml -o internal/delivery/openapi/generated.go openapi/openapi.yaml

graphql-mock:
	npm --prefix graphql/mock install
	npm --prefix graphql/mock start

rest-mock:
	npm --prefix openapi/mock install
	npm --prefix openapi/mock start

mocks:
	mockery --all --dir=./internal --output=./internal/testing/mocks --outpkg=mocks

# Fast local checks: no Docker, database, Redis, MinIO or running API required.
test: test-unit

# Full test pipeline in the same order as CI. A failed stage stops the following stages.
test-all:
	$(MAKE) test-unit-container
	$(MAKE) test-integration
	$(MAKE) test-e2e

test-unit:
	go test -v -shuffle=on $(UNIT_PACKAGES)

test-offline:
	GOPROXY=off GOSUMDB=off go test -v -shuffle=on $(UNIT_PACKAGES)

test-unit-container:
	./scripts/run-compose-tests.sh unit

test-integration:
	./scripts/run-compose-tests.sh integration

test-e2e:
	./scripts/run-compose-tests.sh e2e

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

test-allure: TEST_PACKAGES := ./...
test-allure-unit: TEST_PACKAGES := $(UNIT_PACKAGES)
test-allure test-allure-unit: test-allure-report

test-allure-report:
	rm -rf allure-results allure-report test-results.json
	set +e; go test -shuffle=on -json $(TEST_PACKAGES) > test-results.json; status=$$?; set -e; \
	go run ./scripts/testreport -input test-results.json -output allure-results; \
	if command -v allure >/dev/null 2>&1; then allure generate allure-results --clean -o allure-report; \
	else echo "allure CLI is not installed; allure-results was generated"; fi; \
	echo "Allure report: $$(pwd)/allure-report"; \
	echo "Open it with: make allure-open"; \
	exit $$status

allure-open:
	@test -d allure-report || (echo "allure-report not found; run make test-allure first"; exit 1)
	@command -v allure >/dev/null 2>&1 || (echo "allure CLI is not installed"; exit 1)
	allure open allure-report

clean:
	rm -rf coverage.out coverage.html allure-results allure-report test-results.json
