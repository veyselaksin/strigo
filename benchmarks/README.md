# StriGO Benchmark Charts

This directory contains tools and scripts for generating performance benchmark charts for StriGO.

## 📊 Generated Charts

- **`performance_benchmark.png`** - Operation latency comparison (lower is better)
- **`throughput_benchmark.png`** - Operations per second comparison (higher is better)

## 🚀 Quick Generation

### Automated (Recommended)

```bash
# Run all benchmarks and generate charts
./benchmarks/run_benchmarks.sh
```

### Manual Steps

```bash
# 1. Run benchmarks
cd benchmarks
go test ../tests/redis/performance_test.go -bench=. -count=3 > redis_results.txt
go test ../tests/memcached/performance_test.go -bench=. -count=3 > memcached_results.txt

# 2. Generate charts (requires Python + matplotlib)
python3 generate_chart.py

# 3. Copy to project root
cp *.png ..
```

## 📋 Prerequisites

- **Go 1.22.3+** for running benchmarks
- **Python 3** with matplotlib and numpy
- **Redis** running on localhost:6379
- **Memcached** running on localhost:11211

### Install Python Dependencies

```bash
pip3 install matplotlib numpy
```

## 📁 Files

- **`run_benchmarks.sh`** - Automated benchmark runner
- **`generate_chart.py`** - Python script for chart generation
- **`redis_results.txt`** - Raw Redis benchmark results
- **`memcached_results.txt`** - Raw Memcached benchmark results
- **`*.png`** - Generated chart files

## 🎨 Chart Customization

Edit `generate_chart.py` to customize:

- Colors and styling
- Chart dimensions
- Data formatting
- Additional metrics

## 📊 Understanding Results

### Performance Chart (Latency)

- **Lower values are better**
- Shows time per operation in microseconds
- Useful for understanding individual operation cost

### Throughput Chart (Operations/Second)

- **Higher values are better**
- Shows operations per second capability
- Useful for understanding maximum capacity

## 🔄 Updating Charts

When adding new benchmark tests:

1. Update `generate_chart.py` to parse new test patterns
2. Run `./run_benchmarks.sh` to regenerate charts
3. Charts automatically copied to project root for README

## 💡 Tips

- Run benchmarks multiple times (`-count=3`) for consistent results
- Ensure services are warmed up before benchmarking
- Close other resource-intensive applications during benchmarking
- Results may vary based on system load and hardware
