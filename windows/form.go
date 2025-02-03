package windows

import (
	"fmt"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/rivo/tview"
)	

func RefreshData( logDisplay *tview.TextView, kvDisplay *tview.TextView, redis *utils.RedisConnection) {
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

func ConnectionForm( app *tview.Application, logDisplay *tview.TextView, redis *utils.RedisConnection, kvDisplay *tview.TextView) *tview.Form  {
    form := tview.NewForm()

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
        if(host == "" && port == "") {
            host = "localhost"
            port = "6379"
        }
        logDisplay.Write([]byte(fmt.Sprintf("Connecting to %s:%s ...\n", host, port)))
        
        err := redis.Connect(host, port)
        if err != nil {
            logDisplay.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
            return
        }
        
        logDisplay.Write([]byte("Connected!\n"))
        RefreshData(logDisplay,kvDisplay,redis)  // Initial data load
    })
    
    form.AddButton("Refresh", func() {
        RefreshData(logDisplay,kvDisplay,redis)
    })
    
    form.AddButton("Quit", func() {
        redis.Close()
        app.Stop()
    })
    
    // Set form styling
    form.SetBorder(true).SetTitle(" Redis Connection ")
		
		return form
}