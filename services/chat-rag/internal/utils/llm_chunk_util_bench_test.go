package utils

import (
	"fmt"
	"math/rand"
	"testing"
)

/*
cpu: Intel(R) Core(TM) i7-10700 CPU @ 2.90GHz
BenchmarkChunkStatsEnd-16                                    235          15366981 ns/op    1048842 B/op
6 allocs/op
BenchmarkChunkStatsEndWithVaryingIntervals-16                624           5708615 ns/op    1048782 B/op
6 allocs/op
BenchmarkChunkStatsEndWithSmallDataset-16                 171805             20375 ns/op       8376 B/op
6 allocs/op
BenchmarkChunkStatsEndWithMediumDataset-16                  5178            659392 ns/op     262332 B/op
6 allocs/op
BenchmarkChunkStatsEndSequential-16                         1364           2594378 ns/op    1048782 B/op
6 allocs/op
*/

// BenchmarkChunkStatsEnd 测试 End 方法的性能
// 场景：使用 128K 容量，通过 OnChunkArrivedWithInterval 放满数据后测试 End 性能
func BenchmarkChunkStatsEnd(b *testing.B) {
	const capacity = 128 * 1024 // 128K

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建新的 ChunkStats 实例
		cs := NewChunkStats(capacity)

		// 使用 OnChunkArrivedWithInterval 放满数据，使用随机间隔（10ms到5000ms）
		// 注意：首次调用不会记录间隔（只是初始化 lastTime）
		// 所以需要调用 capacity + 1 次来填充所有间隔
		for j := 0; j < capacity+1; j++ {
			// 生成10到5000之间的随机间隔
			interval := rand.Float32()*4990 + 10 // 10ms 到 5000ms
			cs.OnChunkArrivedWithInterval(interval)
		}

		// 测试 End 性能
		stats := cs.End()
		if stats == nil {
			b.Fatal("Expected stats to be non-nil")
		}

		// 验证数据完整性
		if stats.Count != capacity {
			b.Fatalf("Expected %d intervals, got %d", capacity, stats.Count)
		}
	}
}

// BenchmarkChunkStatsEndWithVaryingIntervals 测试使用不同间隔值的 End 性能
// 场景：模拟真实的 chunk 到达间隔，使用不同的间隔值
func BenchmarkChunkStatsEndWithVaryingIntervals(b *testing.B) {
	const capacity = 128 * 1024 // 128K

	// 生成模拟的间隔值（模拟真实场景：大部分间隔较小，偶尔有较大间隔）
	intervals := generateVaryingIntervals(capacity)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs := NewChunkStats(capacity)

		// 使用预设的间隔值填充数据
		for j := 0; j < capacity+1; j++ {
			idx := j - 1
			if idx < 0 {
				idx = 0
			}
			cs.OnChunkArrivedWithInterval(intervals[idx])
		}

		// 测试 End 性能
		stats := cs.End()
		if stats == nil {
			b.Fatal("Expected stats to be non-nil")
		}

		if stats.Count != capacity {
			b.Fatalf("Expected %d intervals, got %d", capacity, stats.Count)
		}
	}
}

// BenchmarkChunkStatsEndWithSmallDataset 测试小数据集的 End 性能
func BenchmarkChunkStatsEndWithSmallDataset(b *testing.B) {
	const capacity = 1024 // 1K

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs := NewChunkStats(capacity)

		for j := 0; j < capacity+1; j++ {
			cs.OnChunkArrivedWithInterval(10.0)
		}

		stats := cs.End()
		if stats == nil {
			b.Fatal("Expected stats to be non-nil")
		}
	}
}

// BenchmarkChunkStatsEndWithMediumDataset 测试中等数据集的 End 性能
func BenchmarkChunkStatsEndWithMediumDataset(b *testing.B) {
	const capacity = 32 * 1024 // 32K

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs := NewChunkStats(capacity)

		for j := 0; j < capacity+1; j++ {
			cs.OnChunkArrivedWithInterval(10.0)
		}

		stats := cs.End()
		if stats == nil {
			b.Fatal("Expected stats to be non-nil")
		}
	}
}

// BenchmarkChunkStatsEndSequential 测试 End 在同一实例上多次调用的性能（幂等性测试）
func BenchmarkChunkStatsEndSequential(b *testing.B) {
	const capacity = 128 * 1024 // 128K

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs := NewChunkStats(capacity)

		for j := 0; j < capacity+1; j++ {
			cs.OnChunkArrivedWithInterval(10.0)
		}

		// 第一次调用 End
		stats1 := cs.End()
		if stats1 == nil {
			b.Fatal("Expected stats to be non-nil on first call")
		}

		// 第二次调用 End（应该返回 nil，幂等性）
		stats2 := cs.End()
		if stats2 != nil {
			b.Fatal("Expected nil on second End call (idempotent)")
		}
	}
}

// TestChunkStatsEndCorrectness 测试 End 方法的正确性
func TestChunkStatsEndCorrectness(t *testing.T) {
	const capacity = 128 * 1024 // 128K

	cs := NewChunkStats(capacity)

	// 填充数据
	for j := 0; j < capacity+1; j++ {
		cs.OnChunkArrivedWithInterval(10.0)
	}

	stats := cs.End()

	// 验证结果
	if stats == nil {
		t.Fatal("Expected stats to be non-nil")
	}

	if stats.Count != capacity {
		t.Fatalf("Expected %d intervals, got %d", capacity, stats.Count)
	}

	if stats.Mean != 10.0 {
		t.Fatalf("Expected mean 10.0, got %f", stats.Mean)
	}

	if stats.Min != 10.0 {
		t.Fatalf("Expected min 10.0, got %f", stats.Min)
	}

	if stats.Max != 10.0 {
		t.Fatalf("Expected max 10.0, got %f", stats.Max)
	}

	if stats.P50 != 10.0 {
		t.Fatalf("Expected P50 10.0, got %f", stats.P50)
	}

	if stats.P95 != 10.0 {
		t.Fatalf("Expected P95 10.0, got %f", stats.P95)
	}

	if stats.P99 != 10.0 {
		t.Fatalf("Expected P99 10.0, got %f", stats.P99)
	}

	// 验证幂等性
	stats2 := cs.End()
	if stats2 != nil {
		t.Fatal("Expected nil on second End call (idempotent)")
	}
}

// TestChunkStatsPerformanceSummary 打印性能测试摘要
func TestChunkStatsPerformanceSummary(t *testing.T) {
	const capacity = 128 * 1024 // 128K

	cs := NewChunkStats(capacity)

	// 使用不同的间隔值填充
	intervals := generateVaryingIntervals(capacity)
	for j := 0; j < capacity+1; j++ {
		idx := j - 1
		if idx < 0 {
			idx = 0
		}
		cs.OnChunkArrivedWithInterval(intervals[idx])
	}

	stats := cs.End()

	if stats != nil {
		fmt.Printf("\n========== ChunkStats Performance Summary ==========\n")
		fmt.Printf("Total Intervals: %d\n", stats.Count)
		fmt.Printf("Mean: %.2f ms\n", stats.Mean)
		fmt.Printf("Min: %.2f ms\n", stats.Min)
		fmt.Printf("Max: %.2f ms\n", stats.Max)
		fmt.Printf("StdDev: %.2f ms\n", stats.StdDev)
		fmt.Printf("P50: %.2f ms\n", stats.P50)
		fmt.Printf("P95: %.2f ms\n", stats.P95)
		fmt.Printf("P99: %.2f ms\n", stats.P99)
		fmt.Printf("==================================================\n\n")
	}
}

// generateVaryingIntervals 生成模拟的间隔值
// 模拟真实场景：大部分间隔较小（1-20ms），偶尔有较大间隔（50-200ms）
func generateVaryingIntervals(n int) []float32 {
	intervals := make([]float32, n)

	for i := 0; i < n; i++ {
		// 90% 的间隔在 1-20ms 之间
		// 10% 的间隔在 50-200ms 之间
		if i%10 != 0 {
			intervals[i] = float32(1 + (i % 20))
		} else {
			intervals[i] = float32(50 + (i % 150))
		}
	}

	return intervals
}
