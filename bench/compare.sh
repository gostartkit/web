#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BASELINE_FILE="${1:-$ROOT_DIR/bench/baseline.txt}"
CURRENT_FILE="${CURRENT_FILE:-$ROOT_DIR/bench/current.txt}"
COUNT="${COUNT:-5}"
BENCH_EXPR="${BENCH_EXPR:-Benchmark(ServeHTTP|TreeGetValue|TryParse|TryInt|TryUint|TryBool|Post(JSON|Bytes)|DoReqWithClient(Struct|RawBody)|Ctx|ParamsVal|ParseMediaType|AcceptMediaType)}"
SHOW_MISSING="${SHOW_MISSING:-0}"

cd "$ROOT_DIR"

go test -run '^$' -bench "$BENCH_EXPR" -benchmem -count "$COUNT" ./... | tee "$CURRENT_FILE"

printf '\nComparison vs %s\n\n' "$BASELINE_FILE"

awk -v show_missing="$SHOW_MISSING" '
function pct(curr, base) {
	if (base == 0) {
		return "n/a"
	}
	return sprintf("%+.2f%%", ((curr - base) / base) * 100)
}

function parse_metric(prefix, line,   name, rest, arr) {
	name = $1
	rest = line
	sub(/^[^[:space:]]+[[:space:]]+/, "", rest)
	gsub(/[[:space:]]+/, " ", rest)
	split(rest, arr, " ")
	stats[prefix, name, "ns_sum"] += arr[2] + 0
	stats[prefix, name, "b_sum"] += arr[4] + 0
	stats[prefix, name, "allocs_sum"] += arr[6] + 0
	stats[prefix, name, "count"]++
}

function avg(prefix, name, metric) {
	if (stats[prefix, name, "count"] == 0) {
		return 0
	}
	return stats[prefix, name, metric "_sum"] / stats[prefix, name, "count"]
}

FNR == NR {
	if ($1 ~ /^Benchmark/) {
		parse_metric("base", $0)
		seen[$1] = 1
		if (!base_seen[$1]) {
			base_order[++base_count] = $1
			base_seen[$1] = 1
		}
	}
	next
}

{
	if ($1 ~ /^Benchmark/) {
		parse_metric("curr", $0)
		curr_seen[$1] = 1
	}
}

END {
	for (i = 1; i <= base_count; i++) {
		name = base_order[i]
		base_ns[name] = avg("base", name, "ns")
		base_b[name] = avg("base", name, "b")
		base_allocs[name] = avg("base", name, "allocs")
	}
	for (name in curr_seen) {
		curr_ns[name] = avg("curr", name, "ns")
		curr_b[name] = avg("curr", name, "b")
		curr_allocs[name] = avg("curr", name, "allocs")
	}

	printf "%-36s %12s %12s %12s\n", "Benchmark", "ns/op", "B/op", "allocs/op"
	for (i = 1; i <= base_count; i++) {
		name = base_order[i]
		if (!(name in curr_seen)) {
			if (show_missing == 1) {
				printf "%-36s %12s %12s %12s\n", name, "missing", "missing", "missing"
			}
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
