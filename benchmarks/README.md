# estelled Benchmark

Performance measurement tool for estelled using [vegeta](https://github.com/tsenart/vegeta).

## Prerequisites

```bash
go install github.com/tsenart/vegeta@latest
```

## Quick Start

```bash
cd benchmarks
chmod +x run.sh gen_targets.sh
./run.sh -s /path/to/test-image.jpg
```

## Scenarios

| # | Scenario | Endpoint | Concurrency | metric |
|:-:|:---|:---|:---:|:---|
| 01 | Cache Hit | `/get` | 1 | Baseline Latency |
| 02 | Cache Hit | `/get` | 10 | Scaling w/ concurrency |
| 03 | Cache Hit | `/get` | 50 | Latency under load |
| 06 | Queue Saturation | `/queue` | 50 | 503 Rate under load |
| 07 | Queue Saturation | `/queue` | 100 | Queue Saturation Limit |

> **Note:** Queue Saturation scenarios use **Cache Miss** requests (unique sizes) to force thumbnail generation and fill the worker queue.

## Options

| Option | Default | Description |
|:---|:---|:---|
| `-s SOURCE` | **(Required)** | Absolute path to source image |
| `-u URL` | `http://localhost:1186` | estelled Base URL |
| `-k KEY` | (none) | Authentication key (`ESTELLE_SECRET`) |
| `-d DURATION` | `10s` | Duration per scenario |
| `-n COUNT` | `5000` | Number of targets for cache miss |

## Results

Results are saved in `results/<timestamp>/`:

```
results/20260215-000000/
├── 01_hit_c1.bin       # vegeta binary results
├── 01_hit_c1.txt       # text report
├── ...
└── 07_queue_c100.txt
```

Generate HTML plot:

```bash
vegeta plot results/20260215-000000/*.bin > results/20260215-000000/plot.html
```
