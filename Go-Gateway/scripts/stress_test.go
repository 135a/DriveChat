package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 本程序用于对网关进行端到端的压力测试，并量化延迟分布。
// 使用方法：go run stress_test.go -url=http://localhost:8080/api/v1/user -c=100 -n=10000

func main() {
	targetURL := "http://localhost:8080/api/v1/user" // 默认压测目标
	concurrency := 50                               // 默认并发数
	totalRequests := 1000                           // 默认总请求数

	fmt.Printf("[压测启动] 目标: %s, 并发: %d, 总数: %d\n", targetURL, concurrency, totalRequests)

	var wg sync.WaitGroup
	start := time.Now()
	
	results := make(chan time.Duration, totalRequests)
	
	// 启动 Worker 协程池。
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}
			for j := 0; j < totalRequests/concurrency; j++ {
				reqStart := time.Now()
				resp, err := client.Get(targetURL)
				if err == nil {
					resp.Body.Close()
					results <- time.Since(reqStart)
				}
			}
		}()
	}

	wg.Wait()
	close(results)
	
	duration := time.Since(start)
	
	// 统计数据
	var totalDuration time.Duration
	count := 0
	for r := range results {
		totalDuration += r
		count++
	}

	if count > 0 {
		fmt.Printf("\n[压测报告]\n")
		fmt.Printf("总耗时: %v\n", duration)
		fmt.Printf("完成请求: %d\n", count)
		fmt.Printf("平均延迟: %v\n", totalDuration/time.Duration(count))
		fmt.Printf("吞吐量 (QPS): %.2f\n", float64(count)/duration.Seconds())
	}
}
