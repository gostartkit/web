#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TMP_SNAPSHOT="${TMP_SNAPSHOT:-$ROOT_DIR/bench/.snapshot.tmp}"

cd "$ROOT_DIR"

trap 'rm -f "$TMP_SNAPSHOT"' EXIT INT TERM

sh "$ROOT_DIR/bench/snapshot.sh" > "$TMP_SNAPSHOT"

update_file() {
	file=$1
	awk -v snapshot_file="$TMP_SNAPSHOT" '
	BEGIN {
		while ((getline line < snapshot_file) > 0) {
			snapshot[++snapshot_n] = line
		}
		close(snapshot_file)
		in_block = 0
	}

	/<!-- BENCHMARK_SNAPSHOT:BEGIN -->/ {
		print
		for (i = 1; i <= snapshot_n; i++) {
			print snapshot[i]
		}
		in_block = 1
		next
	}

	/<!-- BENCHMARK_SNAPSHOT:END -->/ {
		in_block = 0
		print
		next
	}

	!in_block {
		print
	}
	' "$file" > "$file.tmp"

	mv "$file.tmp" "$file"
}

update_file "$ROOT_DIR/README.md"
update_file "$ROOT_DIR/README_CN.md"

printf 'Updated benchmark snapshot blocks in README.md and README_CN.md\n'
