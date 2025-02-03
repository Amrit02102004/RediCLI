package main

import (
    "fmt"
    "github.com/rivo/tview"
	"github.com/redis/go-redis/v9"
    "context"
    "time"
)

func main() {
    app := tview.NewApplication()
    
    // Create a form for port input
    form := tview.NewForm()
    textView := tview.NewTextView().
        SetDynamicColors(true).
        SetRegions(true).
        SetScrollable(true).
        SetChangedFunc(func() {
            app.Draw()
        })

    // Create a flex container for layout
    flex := tview.NewFlex().
        SetDirection(tview.FlexRow).
        AddItem(form, 3, 1, true).
        AddItem(textView, 0, 1, false)

    // Add port input field
    var port string
    form.AddInputField("Redis Port:", "", 20, nil, func(text string) {
        port = text
    }).
    AddButton("Connect", func() {
        // Create Redis client
        rdb := redis.NewClient(&redis.Options{
            Addr:     fmt.Sprintf("localhost:%s", port),
            Password: "", // no password set
            DB:       0,  // use default DB
        })

        // Test connection
        ctx := context.Background()
        _, err := rdb.Ping(ctx).Result()
        if err != nil {
            textView.SetText(fmt.Sprintf("Error connecting to Redis: %v", err))
            return
        }

        textView.SetText("Connected to Redis! Monitoring logs...\n")
        
        // Start monitoring in a goroutine
        go func() {
            pubsub := rdb.Subscribe(ctx, "__keyspace@0__:*")
            defer pubsub.Close()

            // Monitor all commands
            _, err := rdb.Do(ctx, "MONITOR").Result()
            if err != nil {
                textView.SetText(fmt.Sprintf("Error starting monitor: %v", err))
                return
            }

            for {
                msg, err := pubsub.ReceiveMessage(ctx)
                if err != nil {
                    textView.SetText(fmt.Sprintf("%s\nError receiving message: %v", textView.GetText(true), err))
                    continue
                }

                // Append new log with timestamp
                timestamp := time.Now().Format("15:04:05")
                newLog := fmt.Sprintf("[%s] %s: %s\n", timestamp, msg.Channel, msg.Payload)
                
                app.QueueUpdateDraw(func() {
                    currentText := textView.GetText(true)
                    textView.SetText(currentText + newLog)
                })
            }
        }()
    })

    if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
        panic(err)
    }
}