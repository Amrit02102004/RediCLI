package windows

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
)

func (w *WindowManager) CreateExtraView() *tview.TextView {
	w.extraView = tview.NewTextView()
	w.extraView.SetWrap(true)
	w.extraView.SetBorder(true)
	w.extraView.SetTitle(" Redis Info ")
	w.extraView.Box.SetTitleAlign(tview.AlignLeft)

	w.extraView.SetText("Waiting for Redis connection...")

	return w.extraView
}

func (w *WindowManager) updateExtraInfo() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-w.stopChan:
				return
			case <-ticker.C:
				if w.rdb == nil {
					continue
				}

				info, err := w.rdb.Info(w.ctx).Result()
				if err != nil {
					w.logger.Printf("❌ Error fetching Redis info: %v", err)
					w.appendLog(fmt.Sprintf("Error fetching Redis info: %v", err))
					continue
				}

				// Parse and log Redis info
				infoMap := parseRedisInfo(info)
				w.logger.Printf("💾 Redis Memory Usage: %s", getMapValue(infoMap, "used_memory_human"))
				w.logger.Printf("👥 Connected Clients: %s", getMapValue(infoMap, "connected_clients"))
				
				// Update UI
				w.updateExtraViewText(infoMap)
				
				// Force write logs to disk
				w.logFile.Sync()
			}
		}
	}()
}

func parseRedisInfo(info string) map[string]string {
	infoMap := make(map[string]string)
	for _, line := range strings.Split(info, "\n") {
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				infoMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}
	return infoMap
}

func (w *WindowManager) updateExtraViewText(infoMap map[string]string) {
	w.app.QueueUpdateDraw(func() {
		extraInfo := fmt.Sprintf(
			"Redis Version: %s\n"+
				"Connected Clients: %s\n"+
				"Used Memory: %s\n"+
				"Total Keys: %s\n"+
				"Uptime: %s seconds\n"+
				"Commands Processed: %s\n"+
				"Last Updated: %s",
			getMapValue(infoMap, "redis_version"),
			getMapValue(infoMap, "connected_clients"),
			getMapValue(infoMap, "used_memory_human"),
			getMapValue(infoMap, "keys"),
			getMapValue(infoMap, "uptime_in_seconds"),
			getMapValue(infoMap, "total_commands_processed"),
			time.Now().Format("15:04:05"),
		)
		w.extraView.SetText(extraInfo)
	})
}

func getMapValue(m map[string]string, key string) string {
	if value, exists := m[key]; exists {
		return value
	}
	return "N/A"
}