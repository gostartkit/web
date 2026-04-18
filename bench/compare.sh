#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BASELINE_FILE="${1:-$ROOT_DIR/bench/baseline.txt}"
CURRENT_FILE="$ROOT_DIR/bench/current.txt"
BENCH_EXPR='Benchmark(ServeHTTP|TreeGetValue|TryParse|TryInt|TryUint|TryBool|Post(JSON|Bytes)|DoReqWithClient(Struct|Bytes)|CtxWriteBinaryReader)'

cd "$ROOT_DIR"

go test -run '^$' -bench "$BENCH_EXPR" -benchmem ./... | tee "$CURRENT_FILE"

printf '\nComparison vs %s\n\n' "$BASELINE_FILE"

awk '
function pct(curr, base) {
	if (base == 0) {
		return "n/a"
	}
	return sprintf("%+.2f%%", ((curr - base) / base) * 100)
}

function parse_metric(line,   name, rest, arr) {
	name = $1
	rest = line
	sub(/^[^[:space:]]+[[:space:]]+/, "", rest)
	gsub(/[[:space:]]+/, " ", rest)
	split(rest, arr, " ")
	bench[name, "ns"] = arr[2] + 0
	bench[name, "b"] = arr[4] + 0
	bench[name, "allocs"] = arr[6] + 0
}

FNR == NR {
	if ($1 ~ /^Benchmark/) {
		parse_metric($0)
		seen[$1] = 1
		base_order[++base_count] = $1
		base_ns[$1] = bench[$1, "ns"]
		base_b[$1] = bench[$1, "b"]
		base_allocs[$1] = bench[$1, "allocs"]
	}
	next
}

{
	if ($1 ~ /^Benchmark/) {
		parse_metric($0)
		curr_seen[$1] = 1
		curr_ns[$1] = bench[$1, "ns"]
		curr_b[$1] = bench[$1, "b"]
		curr_allocs[$1] = bench[$1, "allocs"]
	}
}

END {
	printf "%-36s %12s %12s %12s\n", "Benchmark", "ns/op", "B/op", "allocs/op"
	for (i = 1; i <= base_count; i++) {
		name = base_order[i]
		if (!(name in curr_seen)) {
			printf "%-36s %12s %12s %12s\n", name, "missing", "missing", "missing"
			continue
		}
		printf "%-36s %12s %12s %12s\n",
			name,
			pct(curr_ns[name], base_ns[name]),
			pct(curr_b[name], base_b[name]),
			pct(curr_allocs[name], base_allocs[name])
	}
	for (name in curr_seen) {
		if (!(name in seen)) {
			printf "%-36s %12s %12s %12s\n", name, "new", "new", "new"
		}
	}
}
' "$BASELINE_FILE" "$CURRENT_FILE"
