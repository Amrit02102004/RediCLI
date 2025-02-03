// Create left side form for connection details
package windows

import (
	"fmt"

	"github.com/rivo/tview"
)

func refreshData(redis *RedisConnection, logDisplay *tview.TextView, kvDisplay *tview.TextView) {
        if redis.IsConnected() {
            keys, err := redis.GetAllKeys()
            if err != nil {
                logDisplay.Write([]byte(fmt.Sprintf("Error fetching keys: %v\n", err)))
                return
            }
            
            kvDisplay.Clear()
            for _, key := range keys {
                value, err := redis.GetValue(key)
                if err != nil {
                    logDisplay.Write([]byte(fmt.Sprintf("Error fetching value for %s: %v\n", key, err)))
                    continue
                }
                
                ttl, err := redis.GetTTL(key)
                if err != nil {
                    logDisplay.Write([]byte(fmt.Sprintf("Error fetching TTL for %s: %v\n", key, err)))
                    continue
                }
                
                kvDisplay.Write([]byte(fmt.Sprintf("[yellow]Key:[white] %s\n", key)))
                kvDisplay.Write([]byte(fmt.Sprintf("[yellow]Value:[white] %s\n", value)))
                kvDisplay.Write([]byte(fmt.Sprintf("[yellow]TTL:[white] %v\n", ttl)))
                kvDisplay.Write([]byte("------------------------\n"))
            }
        }
    }

func ConnectionForm(app *tview.Application  ) (*tview.Form) {	

    form := tview.NewForm()
		logDisplay := Win2(app) 
    redis := NewRedisConnection() // Create Redis connection instance 
		_, kvDisplay, _ := Win3(app) 

		

		// Add form fields
    var host, port string
    form.AddInputField("Host", "localhost", 20, nil, func(text string) {
        host = text
    })
    form.AddInputField("Port", "6379", 20, nil, func(text string) {
        port = text
    })
    
		// Add buttons
    form.AddButton("Connect", func() {
        logDisplay.SetText("")
        logDisplay.Write([]byte(fmt.Sprintf("Connecting to %s:%s ...\n", host, port)))
        err := redis.Connect(host, port)
        if err != nil {
            logDisplay.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
            return
        }
        
        logDisplay.Write([]byte("Connected!\n"))
        refreshData(redis, logDisplay ,kvDisplay)  // Initial data load
    })
    
    form.AddButton("Refresh", func() {
        refreshData(redis, logDisplay ,kvDisplay)  // Initial data load
    })
    
    form.AddButton("Quit", func() {
        redis.Close()
        app.Stop()
    })
    
    // Set form styling
    form.SetBorder(true).SetTitle(" Redis Connection ")
		
		return form ;
}