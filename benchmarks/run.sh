#!/usr/bin/env bash
#
# Run estelled HTTP benchmark using vegeta.
# Measures latency and req/sec for cache hit and queue saturation scenarios.
#
# Usage:
#   ./run.sh -s /path/to/image.jpg [-u http://localhost:1186] [-k secret] [-d 10s] [-n 5000]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ---------- Defaults ----------
BASE_URL="http://localhost:1186"
KEY=""
DURATION="10s"
COUNT=5000
SOURCE=""

# ---------- Parse Arguments ----------
usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Options:
  -s SOURCE     Absolute path to source image (required)
  -u URL        Base URL of estelled (default: http://localhost:1186)
  -k KEY        Authentication key (corresponding to ESTELLE_SECRET)
  -d DURATION   Duration for each scenario (default: 10s)
  -n COUNT      Number of targets for cache miss (default: 5000)
  -h            Show this help
EOF
    exit 1
}

while getopts "s:u:k:d:n:h" opt; do
    case $opt in
        s) SOURCE="$OPTARG" ;;
        u) BASE_URL="$OPTARG" ;;
        k) KEY="$OPTARG" ;;
        d) DURATION="$OPTARG" ;;
        n) COUNT="$OPTARG" ;;
        h) usage ;;
        *) usage ;;
    esac
done

if [[ -z "$SOURCE" ]]; then
    echo "Error: -s SOURCE is required" >&2
    usage
fi

# ---------- Preflight ----------

if ! command -v vegeta &>/dev/null; then
    echo "Error: vegeta が見つかりません。" >&2
    echo "  go install github.com/tsenart/vegeta@latest" >&2
    exit 1
fi

# ---------- Generate Targets ----------

echo ""
echo "=== Start Generating Targets ==="
bash "${SCRIPT_DIR}/gen_targets.sh" -s "$SOURCE" -u "$BASE_URL" -k "$KEY" -n "$COUNT"

# ---------- Results Directory ----------

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
RESULTS_DIR="${SCRIPT_DIR}/results/${TIMESTAMP}"
mkdir -p "$RESULTS_DIR"
echo ""
echo "Results Directory: ${RESULTS_DIR}"

# ---------- Warmup ----------

echo ""
echo "=== Warmup ==="
KEY_PARAM=""
if [[ -n "$KEY" ]]; then
    KEY_PARAM="&key=${KEY}"
fi

WARMUP_URL="${BASE_URL}/get?source=${SOURCE}&size=85x85${KEY_PARAM}"
if curl -sf --max-time 30 "$WARMUP_URL" > /dev/null 2>&1; then
    echo "Cache primed (85x85)"
else
    echo "Error: Warmup failed. Please check if estelled is running at ${BASE_URL}" >&2
    exit 1
fi

# ---------- Helper ----------

run_scenario() {
    local name="$1"
    local description="$2"
    local targets_file="$3"
    local max_workers="$4"

    echo ""
    echo "--- ${description} ---"

    local bin_file="${RESULTS_DIR}/${name}.bin"
    local txt_file="${RESULTS_DIR}/${name}.txt"

    vegeta attack \
        -name="$name" \
        -targets="$targets_file" \
        -rate=0 \
        -max-workers="$max_workers" \
        -duration="$DURATION" \
        -timeout=30s \
        > "$bin_file"

    vegeta report < "$bin_file" | tee "$txt_file"

    echo "  -> ${name}.txt"
}

# ---------- Target File Paths ----------

HIT_TARGETS="${SCRIPT_DIR}/targets_hit.txt"


# ---------- Scenarios ----------

echo ""
echo "========================================"
echo "  estelled HTTP Benchmark"
echo "  Duration: ${DURATION}  |  Source: ${SOURCE}"
echo "========================================"

# 1. Cache Hit
run_scenario "01_hit_c1"  "Cache Hit - 1 Client (c=1)"    "$HIT_TARGETS"  1
run_scenario "02_hit_c10" "Cache Hit - 10 Clients (c=10)" "$HIT_TARGETS" 10
run_scenario "03_hit_c50" "Cache Hit - 50 Clients (c=50)" "$HIT_TARGETS" 50

# 2. Queue Saturation /queue
QUEUE_TARGETS="${SCRIPT_DIR}/targets_miss_queue.txt"
run_scenario "06_queue_c50"  "Queue /queue - 50 Clients (c=50)"   "$QUEUE_TARGETS"  50
run_scenario "07_queue_c100" "Queue /queue - 100 Clients (c=100)" "$QUEUE_TARGETS" 100




# ---------- Summary ----------

echo ""
echo "========================================"
echo "  All scenarios completed"
echo "  Results: ${RESULTS_DIR}"
echo "========================================"
echo ""
echo "Generate HTML Plot: vegeta plot ${RESULTS_DIR}/*.bin > ${RESULTS_DIR}/plot.html"
