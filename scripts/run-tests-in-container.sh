#!/usr/bin/env bash

set -euo pipefail

suite="${1:?test suite is required: unit, integration or e2e}"

case "$suite" in
unit)
	packages=(
		./internal/service
		./internal/infrastructure/auth
		./internal/infrastructure/logger
		./internal/delivery/middleware
		./internal/worker
	)
	;;
integration)
	packages=(./tests/integration ./internal/infrastructure/db/postgres/repository)
	;;
e2e)
	packages=(./tests/e2e)
	;;
*)
	echo "unknown test suite: $suite" >&2
	exit 2
	;;
esac

mkdir -p /artifacts

set +e
go test -p 1 -shuffle=on -json "${packages[@]}" | tee /artifacts/test-results.json
test_status=${PIPESTATUS[0]}
set -e

go run ./scripts/testreport \
	-input /artifacts/test-results.json \
	-output /artifacts/allure-results

exit "$test_status"
