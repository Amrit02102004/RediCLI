package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type AnalyticsData struct {
	TotalKeys        int64          `json:"totalKeys"`
	PersistentKeys   int64          `json:"persistentKeys"`
	ExpiringKeys     int64          `json:"expiringKeys"`
	MemoryUsedBytes  int64          `json:"memoryUsedBytes"`
	MemoryTotalBytes int64          `json:"memoryTotalBytes"`
	KeyExpirations   map[string]int `json:"keyExpirations"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for simplicity
	},
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

	// WebSocket endpoint for continuous analytics
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Could not open WebSocket connection", http.StatusBadRequest)
			return
		}
		defer conn.Close()

		// Analytics update ticker
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				analyticsData, err := rc.GetAnalytics()
				if err != nil {
					// If error occurs, send error message
					conn.WriteJSON(map[string]string{"error": err.Error()})
					continue
				}

				// Send analytics data through WebSocket
				err = conn.WriteJSON(analyticsData)
				if err != nil {
					// Stop if client disconnects
					return
				}
			}
		}
	})

	// Existing REST endpoint
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		analyticsData, err := rc.GetAnalytics()
		if err != nil {
			http.Error(w, "Failed to retrieve analytics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analyticsData)
	})

	// Serve HTML dashboard with WebSocket support
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
		<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Redis Analytics Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.7.1/dist/chart.min.js"></script>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            background-color: #121212; 
            color: #ffffff; 
            margin: 0;
            padding: 10px;
            font-size: 12px;
        }
        .dashboard { 
            display: grid; 
            grid-template-columns: 1fr 1fr; 
            gap: 10px; 
        }
        .chart-container { 
            background-color: #1e1e1e; 
            border-radius: 5px; 
            box-shadow: 0 2px 4px rgba(0,0,0,0.2); 
            padding: 10px;
        }
        .full-width { grid-column: 1 / -1; }
        h1 {
            font-size: 24px;
            margin: 0 0 20px 0;
            font-weight: bold;
            color: #ffffff;
        }
        h2 { 
            font-size: 14px;
            margin: 5px 0;
            font-weight: bold;
            color: #ffffff;
        }
        canvas { 
            max-height: 150px !important;
            height: 150px !important;
        }
        #keyTypesChart, #keyExpirationChart {
            max-height: 150px !important;
            height: 150px !important;
        }
    </style>
</head>
<body>
    <h1>Redis Analytics Dashboard</h1>
    <div class="dashboard">
        <div class="chart-container">
            <h2>Key Types</h2>
            <canvas id="keyTypesChart"></canvas>
        </div>
        <div class="chart-container">
            <h2>Key Expirations</h2>
            <canvas id="keyExpirationChart"></canvas>
        </div>
        <div class="chart-container full-width">
            <h2>Memory Usage</h2>
            <canvas id="memoryUsageLineChart"></canvas>
        </div>
        <div class="chart-container full-width">
            <h2>Key Count</h2>
            <canvas id="keyUsageLineChart"></canvas>
        </div>
    </div>

    <script>
        var globalCharts = {};
        var memoryHistory = [];
        var keyHistory = [];
        var socket;

        function initWebSocket() {
            socket = new WebSocket('ws://' + window.location.host + '/ws');
            
            socket.onmessage = function(event) {
                try {
                    var data = JSON.parse(event.data);
                    updateCharts(data);
                } catch (error) {
                    console.error('Parsing error:', error);
                }
            };
        }

        function updateCharts(data) {
            Object.values(globalCharts).forEach(chart => chart && chart.destroy());

            var timestamp = new Date().toLocaleTimeString();
            
            memoryHistory.push({
                x: timestamp,
                y: Number((data.memoryUsedBytes / (1024 * 1024)).toFixed(2))
            });
            if (memoryHistory.length > 10) memoryHistory.shift();

            keyHistory.push({
                x: timestamp,
                y: data.totalKeys
            });
            if (keyHistory.length > 10) keyHistory.shift();

            globalCharts.keyTypes = new Chart(document.getElementById('keyTypesChart'), {
                type: 'pie',
                options: { 
                    plugins: { legend: { position: 'bottom' } },
                    animation: false
                },
                data: {
                    labels: ['Persistent', 'Expiring'],
                    datasets: [{
                        data: [data.persistentKeys, data.expiringKeys],
                        backgroundColor: ['#36A2EB', '#FF6384']
                    }]
                }
            });

            globalCharts.keyExpiration = new Chart(document.getElementById('keyExpirationChart'), {
                type: 'bar',
                options: { 
                    plugins: { legend: { display: false } },
                    scales: { 
                        y: { 
                            beginAtZero: true,
                            grid: {
                                color: 'rgba(255, 255, 255, 0.1)'
                            }
                        },
                        x: {
                            grid: {
                                color: 'rgba(255, 255, 255, 0.1)'
                            }
                        }
                    },
                    animation: false
                },
                data: {
                    labels: Object.keys(data.keyExpirations || {}),
                    datasets: [{
                        label: 'Key Expirations',
                        data: Object.values(data.keyExpirations || {}),
                        backgroundColor: '#4BC0C0'
                    }]
                }
            });

            globalCharts.memoryLine = new Chart(document.getElementById('memoryUsageLineChart'), {
                type: 'line',
                options: { 
                    plugins: { legend: { display: true } },
                    scales: { 
                        y: { 
                            beginAtZero: true,
                            grid: {
                                color: 'rgba(255, 255, 255, 0.1)'
                            }
                        },
                        x: {
                            grid: {
                                color: 'rgba(255, 255, 255, 0.1)'
                            }
                        }
                    },
                    animation: false
                },
                data: {
                    labels: memoryHistory.map(d => d.x),
                    datasets: [{
                        label: 'Memory Usage (MB)',
                        data: memoryHistory,
                        borderColor: '#9966FF',
                        fill: false
                    }]
                }
            });

            globalCharts.keyLine = new Chart(document.getElementById('keyUsageLineChart'), {
                type: 'line',
                options: { 
                    plugins: { legend: { display: true } },
                    scales: { 
                        y: { 
                            beginAtZero: true,
                            grid: {
                                color: 'rgba(255, 255, 255, 0.1)'
                            }
                        },
                        x: {
                            grid: {
                                color: 'rgba(255, 255, 255, 0.1)'
                            }
                        }
                    },
                    animation: false
                },
                data: {
                    labels: keyHistory.map(d => d.x),
                    datasets: [{
                        label: 'Key Count',
                        data: keyHistory,
                        borderColor: '#FF6384',
                        fill: false
                    }]
                }
            });
        }

        window.onload = initWebSocket;
    </script>
</body>
</html>
		`))
	})

	go openBrowser(url)
	return http.Serve(listener, nil)
}

// Existing helper functions remain the same (GetAnalytics, getBucket, openBrowser)
func (rc *RedisConnection) GetAnalytics() (*AnalyticsData, error) {
	if rc.client == nil {
		return nil, fmt.Errorf("not connected to Redis")
	}

	// Create a context with timeout to prevent long-running operations
	ctx, cancel := context.WithTimeout(rc.ctx, 10*time.Second)
	defer cancel()

	var analytics AnalyticsData
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []error

	// Key Scanning with Cursor-Based Approach
	wg.Add(1)
	go func() {
		defer wg.Done()
		var cursor uint64
		var keys []string
		var err error

		persistentCount, expiringCount := int64(0), int64(0)
		keyExpirations := make(map[string]int)

		for {
			keys, cursor, err = rc.client.Scan(ctx, cursor, "*", 1000).Result()
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("keys scan error: %v", err))
				mu.Unlock()
				return
			}

			for _, key := range keys {
				ttl, _ := rc.client.TTL(ctx, key).Result()
				if ttl == -1 {
					persistentCount++
				} else if ttl > 0 {
					expiringCount++
					bucket := getBucket(ttl)
					keyExpirations[bucket]++
				}
			}

			if cursor == 0 {
				break
			}
		}

		mu.Lock()
		analytics.PersistentKeys = persistentCount
		analytics.ExpiringKeys = expiringCount
		analytics.KeyExpirations = keyExpirations
		mu.Unlock()
	}()

	// Database Size
	wg.Add(1)
	go func() {
		defer wg.Done()
		size, err := rc.client.DBSize(ctx).Result()
		mu.Lock()
		if err != nil {
			errs = append(errs, fmt.Errorf("DB size error: %v", err))
		} else {
			analytics.TotalKeys = size
		}
		mu.Unlock()
	}()

	// Memory Info
	wg.Add(1)
	go func() {
		defer wg.Done()
		memInfo, err := rc.client.Info(ctx, "memory").Result()
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

// Helper function to categorize TTL
func getBucket(ttl time.Duration) string {
	if ttl <= time.Minute {
		return "< 1 min"
	} else if ttl <= 10*time.Minute {
		return "1-10 min"
	}
	return "> 10 min"
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
