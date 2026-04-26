#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BASELINE_FILE="${1:-$ROOT_DIR/bench/baseline.txt}"
COUNT="${COUNT:-5}"
BENCH_EXPR="${BENCH_EXPR:-Benchmark(ServeHTTP|TreeGetValue|TryParse|TryInt|TryUint|TryBool|Post(JSON|Bytes)|DoReqWithClient(Struct|RawBody)|Ctx|ParamsVal|ParseMediaType|AcceptMediaType)}"

cd "$ROOT_DIR"

TMP_FILE="${BASELINE_FILE}.tmp"
trap 'rm -f "$TMP_FILE"' EXIT INT TERM

go test -run '^$' -bench "$BENCH_EXPR" -benchmem -count "$COUNT" ./... > "$TMP_FILE"
mv "$TMP_FILE" "$BASELINE_FILE"

printf 'Updated baseline: %s\n' "$BASELINE_FILE"
