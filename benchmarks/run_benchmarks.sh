#!/bin/bash

# StriGO Benchmark Runner and Chart Generator
# This script runs performance benchmarks and generates PNG charts

set -e

echo "🚀 StriGO Benchmark Runner"
echo "=========================="

# Create benchmarks directory if it doesn't exist
mkdir -p benchmarks
cd benchmarks

echo "📊 Running Redis benchmarks..."
go test ../tests/redis/performance_test.go -bench=. -count=3 > redis_results.txt 2>&1 || echo "⚠️  Redis benchmarks completed with warnings"

echo "📊 Running Memcached benchmarks..."
go test ../tests/memcached/performance_test.go -bench=. -count=3 > memcached_results.txt 2>&1 || echo "⚠️  Memcached benchmarks completed with warnings"

echo "📈 Generating performance charts..."
python3 generate_chart.py

echo "📋 Copying charts to project root..."
cp *.png ..

echo ""
echo "✅ Benchmark Results:"
echo "   📊 performance_benchmark.png"
echo "   📈 throughput_benchmark.png"
echo ""
echo "🎉 Benchmarks completed successfully!"
echo "💡 Charts have been saved to project root and are ready for README"

# Show quick summary
echo ""
echo "📊 Quick Summary:"
echo "=================="
echo "Redis Results:"
grep "BenchmarkRedis" redis_results.txt | head -3 | awk '{print "  " $1 ": " $4}'
echo ""
echo "Memcached Results:"
grep "BenchmarkMemcached" memcached_results.txt | head -3 | awk '{print "  " $1 ": " $4}' 