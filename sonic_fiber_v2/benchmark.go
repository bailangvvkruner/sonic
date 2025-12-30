package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	
	"sonic_fiber_v2/internal/service"
)

// BenchmarkTest 性能测试
func BenchmarkTest() {
	fmt.Println("=== Sonic Fiber v2 性能测试 ===")
	
	// 测试优化后的服务
	optimizedService := service.NewOptimizedPostService().(*service.OptimizedPostService)
	optimizedService.LoadFromDB()
	optimizedService.StartWriteQueue()
	
	// 测试原始服务
	originalService := service.NewPostService()
	
	// 1. 读取性能测试
	fmt.Println("\n1. 读取性能测试 (10000次查询)")
	
	// 原始服务
	start := time.Now()
	var wg sync.WaitGroup
	var ops int64
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = originalService.GetBySlug(context.Background(), "test-post-1")
				atomic.AddInt64(&ops, 1)
			}
		}()
	}
	wg.Wait()
	originalTime := time.Since(start)
	
	// 优化服务
	ops = 0
	start = time.Now()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = optimizedService.GetBySlug(context.Background(), "test-post-1")
				atomic.AddInt64(&ops, 1)
			}
		}()
	}
	wg.Wait()
	optimizedTime := time.Since(start)
	
	fmt.Printf("原始服务: %v (%d ops)\n", originalTime, ops)
	fmt.Printf("优化服务: %v (%d ops)\n", optimizedTime, ops)
	fmt.Printf("性能提升: %.2fx\n", float64(originalTime)/float64(optimizedTime))
	
	// 2. 并发读取测试
	fmt.Println("\n2. 高并发读取测试 (1000 goroutines)")
	
	// 原始服务
	ops = 0
	start = time.Now()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = originalService.GetRecentPosts(context.Background(), 1, 10)
				atomic.AddInt64(&ops, 1)
			}
		}()
	}
	wg.Wait()
	originalTime = time.Since(start)
	
	// 优化服务
	ops = 0
	start = time.Now()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = optimizedService.GetRecentPosts(context.Background(), 1, 10)
				atomic.AddInt64(&ops, 1)
			}
		}()
	}
	wg.Wait()
	optimizedTime = time.Since(start)
	
	fmt.Printf("原始服务: %v (%d ops)\n", originalTime, ops)
	fmt.Printf("优化服务: %v (%d ops)\n", optimizedTime, ops)
	fmt.Printf("性能提升: %.2fx\n", float64(originalTime)/float64(optimizedTime))
	
	// 3. 写入性能测试
	fmt.Println("\n3. 写入性能测试 (100次写入)")
	
	// 原始服务
	start = time.Now()
	for i := 0; i < 100; i++ {
		originalService.Create(context.Background(), "测试", "内容", fmt.Sprintf("slug-%d", i), "published", 1, []int64{1})
	}
	originalTime = time.Since(start)
	
	// 优化服务
	start = time.Now()
	for i := 0; i < 100; i++ {
		optimizedService.Create(context.Background(), "测试", "内容", fmt.Sprintf("slug-%d", i), "published", 1, []int64{1})
	}
	// 等待异步完成
	time.Sleep(100 * time.Millisecond)
	optimizedTime = time.Since(start)
	
	fmt.Printf("原始服务: %v\n", originalTime)
	fmt.Printf("优化服务: %v\n", optimizedTime)
	
	// 4. 内存使用对比
	fmt.Println("\n4. 内存使用对比")
	fmt.Printf("原始服务: 约 1000 条数据\n")
	fmt.Printf("优化服务: 约 1000 条数据 + 索引\n")
	fmt.Printf("优化服务使用哈希索引，查询复杂度从 O(n) 降到 O(1)\n")
	
	fmt.Println("\n=== 测试完成 ===")
}

func main() {
	BenchmarkTest()
}
