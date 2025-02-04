package utils

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

type AnalyticsData struct {
	TotalKeys        int64
	PersistentKeys   int64
	ExpiringKeys     int64
	MemoryUsedBytes  int64
	MemoryTotalBytes int64
	KeyExpirations   map[string]int
}

func (rc *RedisConnection) GetAnalytics() (*AnalyticsData, error) {
	if rc.client == nil {
		return nil, fmt.Errorf("not connected to Redis")
	}

	var wg sync.WaitGroup
	var analytics AnalyticsData
	var errs []error
	var mu sync.Mutex

	wg.Add(4)
	go func() {
		defer wg.Done()
		size, err := rc.client.DBSize(rc.ctx).Result()
		mu.Lock()
		if err != nil {
			errs = append(errs, fmt.Errorf("DB size error: %v", err))
		} else {
			analytics.TotalKeys = size
		}
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		keys, err := rc.client.Keys(rc.ctx, "*").Result()
		mu.Lock()
		if err != nil {
			errs = append(errs, fmt.Errorf("keys scan error: %v", err))
		} else {
			persistentCount, expiringCount := int64(0), int64(0)
			keyExpirations := make(map[string]int)
			for _, key := range keys {
				ttl, _ := rc.client.TTL(rc.ctx, key).Result()
				if ttl == -1 {
					persistentCount++
				} else if ttl > 0 {
					expiringCount++
					bucket := "< 1 min"
					if ttl.Minutes() > 1 {
						bucket = "1-10 min"
					}
					if ttl.Minutes() > 10 {
						bucket = "> 10 min"
					}
					keyExpirations[bucket]++
				}
			}
			analytics.PersistentKeys = persistentCount
			analytics.ExpiringKeys = expiringCount
			analytics.KeyExpirations = keyExpirations
		}
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		memInfo, err := rc.client.Info(rc.ctx, "memory").Result()
		mu.Lock()
		if err != nil {
			errs = append(errs, fmt.Errorf("memory info error: %v", err))
		} else {
			lines := strings.Split(memInfo, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "used_memory:") {
					fmt.Sscanf(line, "used_memory:%d", &analytics.MemoryUsedBytes)
				}
				if strings.HasPrefix(line, "total_system_memory:") {
					fmt.Sscanf(line, "total_system_memory:%d", &analytics.MemoryTotalBytes)
				}
			}
		}
		mu.Unlock()
	}()

	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("multiple errors: %v", errs)
	}

	return &analytics, nil
}
func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // Linux
		cmd = "xdg-open"
		args = []string{url}
	}

	exec.Command(cmd, args...).Start()
}

func (rc *RedisConnection) ServeAnalytics() error {
	startPort := 8080
	maxAttempts := 100
	var listener net.Listener
	var err error

	for port := startPort; port < startPort+maxAttempts; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "address already in use") {
			continue
		}
		return fmt.Errorf("failed to listen: %v", err)
	}
	if listener == nil {
		return fmt.Errorf("no available port found after %d attempts", maxAttempts)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", port)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		analyticsHTML := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>Redis Analytics Dashboard</title>
		</head>
		<body class="bg-gray-100 p-6">
			<h1 class="text-2xl font-bold text-center mb-6">Redis Analytics Dashboard</h1>
		</body>
		</html>
		`)

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(analyticsHTML))
	})

	go openBrowser(url)
	return http.Serve(listener, nil)
}
