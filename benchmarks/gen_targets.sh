#!/usr/bin/env bash
#
# Generate vegeta target files for estelled benchmarking.
#
# Usage:
#   ./gen_targets.sh -s /path/to/image.jpg [-u http://localhost:1186] [-k secret] [-n 5000]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ---------- Defaults ----------
BASE_URL="http://localhost:1186"
KEY=""

SOURCE=""

# ---------- Parse Arguments ----------
usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Options:
  -s SOURCE   Absolute path to source image (required)
  -u URL      Base URL of estelled (default: http://localhost:1186)
  -k KEY      Authentication key (corresponding to ESTELLE_SECRET)

  -h          Show this help
EOF
    exit 1
}

while getopts "s:u:k:n:h" opt; do
    case $opt in
        s) SOURCE="$OPTARG" ;;
        u) BASE_URL="$OPTARG" ;;
        k) KEY="$OPTARG" ;;
        n) COUNT="$OPTARG" ;;
        h) usage ;;
        *) usage ;;
    esac
done

if [[ -z "$SOURCE" ]]; then
    echo "Error: -s SOURCE is required" >&2
    usage
fi

KEY_PARAM=""
if [[ -n "$KEY" ]]; then
    KEY_PARAM="&key=${KEY}"
fi

# --- Cache Hit: single target (vegeta repeats it) ---
echo "GET ${BASE_URL}/get?source=${SOURCE}&size=85x85${KEY_PARAM}" \
    > "${SCRIPT_DIR}/targets_hit.txt"
echo "[OK] targets_hit.txt  (1 target for cache-hit)"

# --- Cache Miss /queue: unique size, offset to avoid collision with /get ---
{
    for ((i = 0; i < COUNT; i++)); do
        s=$((i + COUNT + 50))
        echo "GET ${BASE_URL}/queue?source=${SOURCE}&size=${s}x${s}${KEY_PARAM}"
    done
} > "${SCRIPT_DIR}/targets_miss_queue.txt"
echo "[OK] targets_miss_queue.txt  (${COUNT} targets for cache-miss /queue)"

