#!/usr/bin/env bash

set -euo pipefail

suite="${1:?test suite is required: unit, integration or e2e}"
case "$suite" in
unit) runner=unit-runner ;;
integration) runner=integration-runner ;;
e2e) runner=e2e-runner ;;
*)
	echo "unknown test suite: $suite" >&2
	exit 2
	;;
esac

run_id="${CI_PIPELINE_ID:-${GITHUB_RUN_ID:-${BUILD_NUMBER:-$$}}}"
project="${COMPOSE_PROJECT_NAME:-zvideo-${suite}-${run_id}-$$}"
artifact_dir="${TEST_ARTIFACTS_DIR:-$PWD/test-artifacts/$project}"
history_dir="${ALLURE_HISTORY_DIR:-$PWD/.allure-history/$suite}"
mkdir -p "$artifact_dir"
artifact_dir="$(cd "$artifact_dir" && pwd)"
history_dir="$(mkdir -p "$history_dir" && cd "$history_dir" && pwd)"

compose=(docker compose --project-name "$project" --file docker-compose.test.yml)

cleanup() {
	status=$?
	set +e
	"${compose[@]}" --profile "$suite" down --volumes --remove-orphans >/dev/null 2>&1
	if [ -d "$artifact_dir/allure-results" ] && [ -n "$(find "$history_dir" -mindepth 1 -maxdepth 1 -print -quit)" ]; then
		mkdir -p "$artifact_dir/allure-results/history"
		cp -a "$history_dir"/. "$artifact_dir/allure-results/history/"
	fi
	if command -v allure >/dev/null 2>&1 && [ -d "$artifact_dir/allure-results" ]; then
		allure generate "$artifact_dir/allure-results" --clean -o "$artifact_dir/allure-report" >/dev/null
		if [ -d "$artifact_dir/allure-report/history" ]; then
			cp -a "$artifact_dir/allure-report/history"/. "$history_dir"/
		fi
	fi
	echo "Test artifacts: $artifact_dir"
	exit "$status"
}
trap cleanup EXIT INT TERM

if "${compose[@]}" --profile "$suite" up --build --abort-on-container-exit --exit-code-from "$runner" "$runner"; then
	test_status=0
else
	test_status=$?
fi

if "${compose[@]}" cp "$runner:/artifacts/." "$artifact_dir/"; then
	copy_status=0
else
	copy_status=$?
fi

if [ "$test_status" -eq 0 ] && [ "$copy_status" -ne 0 ]; then
	test_status=$copy_status
fi

exit "$test_status"
