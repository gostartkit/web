#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
COUNT="${COUNT:-1}"
BENCH_EXPR="${BENCH_EXPR:-Benchmark(ServeHTTPStaticJSON|ServeHTTPPathParamJSON|ServeHTTPStaticJSONRawMessage|TryParseJSONBodyFast|PostBytes|DoReqWithClientRawBody|ServeHTTPBinary|ServeHTTPAvro|TreeGetValueParamPooled|TryParseIntSlice|TryParseStringSlice)}"
TMP_FILE="${TMP_FILE:-$ROOT_DIR/bench/snapshot.txt}"

cd "$ROOT_DIR"

trap 'rm -f "$TMP_FILE"' EXIT INT TERM

go test -run '^$' -bench "$BENCH_EXPR" -benchmem -count "$COUNT" ./... > "$TMP_FILE"

awk '
function parse_metric(line,   name, rest, arr) {
	name = $1
	sub(/-[0-9]+$/, "", name)
	rest = line
	sub(/^[^[:space:]]+[[:space:]]+/, "", rest)
	gsub(/[[:space:]]+/, " ", rest)
	split(rest, arr, " ")
	stats[name, "ns_sum"] += arr[2] + 0
	stats[name, "b_sum"] += arr[4] + 0
	stats[name, "allocs_sum"] += arr[6] + 0
	stats[name, "count"]++
	if (!(name in seen)) {
		order[++n] = name
		seen[name] = 1
	}
}

function avg(name, metric) {
	return stats[name, metric "_sum"] / stats[name, "count"]
}

$1 ~ /^Benchmark/ {
	parse_metric($0)
}

END {
	print "| Benchmark | Result | Memory |"
	print "|---|---:|---:|"
	for (i = 1; i <= n; i++) {
		name = order[i]
		printf "| `%s` | `%.1f ns/op` | `%.0f B/op`, `%.0f alloc/op` |\n",
			name, avg(name, "ns"), avg(name, "b"), avg(name, "allocs")
	}
}
' "$TMP_FILE"
