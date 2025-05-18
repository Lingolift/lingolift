package job

import (
	"io"
	"log"
	"net/http"
	"time"
)

func HealthCheck() {
	interval := 20

	for {
		time.Sleep(time.Duration(interval) * time.Second)

		client := &http.Client{Timeout: 10 * time.Second}

		start := time.Now()
		duration := time.Since(start)
		timestamp := time.Now().Format("2006-01-02 15:04:05")

		resp, err := client.Get("https://lingolift.onrender.com/health")
		if err != nil {
			log.Printf("[%s] ❌ 请求失败: %v (耗时: %v)", timestamp, err, duration)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		// 输出检查结果
		log.Printf("[%s] ✅ 状态码: %d | 耗时: %v | 响应: %s",
			timestamp, resp.StatusCode, duration, string(body))
	}
}
