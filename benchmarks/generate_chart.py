#!/usr/bin/env python3

import matplotlib.pyplot as plt
import numpy as np
import re

def parse_benchmark_results(filename):
    """Parse Go benchmark results from file"""
    with open(filename, 'r') as f:
        content = f.read()
    
    # Extract benchmark data
    pattern = r'Benchmark(\w+)-\d+\s+(\d+)\s+(\d+)\s+ns/op'
    matches = re.findall(pattern, content)
    
    benchmarks = {}
    for match in matches:
        name, iterations, ns_per_op = match
        if name not in benchmarks:
            benchmarks[name] = []
        benchmarks[name].append(int(ns_per_op))
    
    # Calculate averages
    avg_benchmarks = {}
    for name, values in benchmarks.items():
        avg_benchmarks[name] = np.mean(values)
    
    return avg_benchmarks

def create_performance_chart():
    """Create performance comparison chart"""
    # Parse results
    redis_results = parse_benchmark_results('redis_results.txt')
    memcached_results = parse_benchmark_results('memcached_results.txt')
    
    # Prepare data
    operations = ['Consume', 'Get', 'Reset', 'MixedOperations']
    redis_times = []
    memcached_times = []
    
    for op in operations:
        redis_key = f'Redis{op}'
        memcached_key = f'Memcached{op}'
        
        redis_times.append(redis_results.get(redis_key, 0) / 1000)  # Convert to microseconds
        memcached_times.append(memcached_results.get(memcached_key, 0) / 1000)
    
    # Create chart
    x = np.arange(len(operations))
    width = 0.35
    
    fig, ax = plt.subplots(figsize=(12, 8))
    
    bars1 = ax.bar(x - width/2, redis_times, width, label='Redis', color='#e74c3c', alpha=0.8)
    bars2 = ax.bar(x + width/2, memcached_times, width, label='Memcached', color='#3498db', alpha=0.8)
    
    # Styling
    ax.set_xlabel('Operations', fontsize=12, fontweight='bold')
    ax.set_ylabel('Time (microseconds)', fontsize=12, fontweight='bold')
    ax.set_title('StriGO Performance Benchmarks\n(Lower is Better)', fontsize=16, fontweight='bold', pad=20)
    ax.set_xticks(x)
    ax.set_xticklabels(operations)
    ax.legend(fontsize=11)
    ax.grid(True, alpha=0.3, axis='y')
    
    # Add value labels on bars
    def add_value_labels(bars):
        for bar in bars:
            height = bar.get_height()
            if height > 0:
                ax.annotate(f'{height:.1f}μs',
                           xy=(bar.get_x() + bar.get_width() / 2, height),
                           xytext=(0, 3),
                           textcoords="offset points",
                           ha='center', va='bottom',
                           fontsize=9, fontweight='bold')
    
    add_value_labels(bars1)
    add_value_labels(bars2)
    
    plt.tight_layout()
    plt.savefig('performance_benchmark.png', dpi=300, bbox_inches='tight')
    print("✅ Performance chart saved as 'performance_benchmark.png'")

def create_operations_per_second_chart():
    """Create operations per second chart"""
    redis_results = parse_benchmark_results('redis_results.txt')
    memcached_results = parse_benchmark_results('memcached_results.txt')
    
    # Convert to operations per second
    operations = ['Consume', 'Get', 'Reset']
    redis_ops = []
    memcached_ops = []
    
    for op in operations:
        redis_key = f'Redis{op}'
        memcached_key = f'Memcached{op}'
        
        redis_ns = redis_results.get(redis_key, 0)
        memcached_ns = memcached_results.get(memcached_key, 0)
        
        redis_ops.append(1_000_000_000 / redis_ns if redis_ns > 0 else 0)  # ops/second
        memcached_ops.append(1_000_000_000 / memcached_ns if memcached_ns > 0 else 0)
    
    # Create chart
    x = np.arange(len(operations))
    width = 0.35
    
    fig, ax = plt.subplots(figsize=(12, 8))
    
    bars1 = ax.bar(x - width/2, redis_ops, width, label='Redis', color='#e74c3c', alpha=0.8)
    bars2 = ax.bar(x + width/2, memcached_ops, width, label='Memcached', color='#3498db', alpha=0.8)
    
    # Styling
    ax.set_xlabel('Operations', fontsize=12, fontweight='bold')
    ax.set_ylabel('Operations per Second', fontsize=12, fontweight='bold')
    ax.set_title('StriGO Throughput Benchmarks\n(Higher is Better)', fontsize=16, fontweight='bold', pad=20)
    ax.set_xticks(x)
    ax.set_xticklabels(operations)
    ax.legend(fontsize=11)
    ax.grid(True, alpha=0.3, axis='y')
    
    # Format y-axis to show K/M suffixes
    ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda x, p: f'{x/1000:.1f}K' if x >= 1000 else f'{x:.0f}'))
    
    # Add value labels on bars
    def add_ops_labels(bars):
        for bar in bars:
            height = bar.get_height()
            if height > 0:
                label = f'{height/1000:.1f}K' if height >= 1000 else f'{height:.0f}'
                ax.annotate(label,
                           xy=(bar.get_x() + bar.get_width() / 2, height),
                           xytext=(0, 3),
                           textcoords="offset points",
                           ha='center', va='bottom',
                           fontsize=9, fontweight='bold')
    
    add_ops_labels(bars1)
    add_ops_labels(bars2)
    
    plt.tight_layout()
    plt.savefig('throughput_benchmark.png', dpi=300, bbox_inches='tight')
    print("✅ Throughput chart saved as 'throughput_benchmark.png'")

if __name__ == "__main__":
    print("🚀 Generating benchmark charts...")
    create_performance_chart()
    create_operations_per_second_chart()
    print("🎉 All charts generated successfully!") 