package cmd

import (
    "fmt"
    "strings"
    "github.com/rivo/tview"
    "github.com/Amrit02102004/RediCLI/windows"
     "github.com/gdamore/tcell/v2"
)

func Func() {
    app := tview.NewApplication()
    
    // Create the main flex container
    flex := tview.NewFlex()
    
    // Create left side form for connection
    form := tview.NewForm()
    
    // Create center command input and key-value display
    cmdFlex := tview.NewFlex().SetDirection(tview.FlexRow)
    
    // Create key-value display
    kvDisplay := tview.NewTextView().
        SetDynamicColors(true).
        SetChangedFunc(func() {
            app.Draw()
        })
    kvDisplay.SetBorder(true).SetTitle(" Redis Data ")
    
    // Create command input field
    cmdInput := tview.NewInputField().
        SetLabel("> ").
        SetFieldWidth(0)
    
    // Create right side text view for logs
    logs := tview.NewTextView().
        SetDynamicColors(true).
        SetChangedFunc(func() {
            app.Draw()
        })
    
    // Add form fields
    var host, port string
    form.AddInputField("Host", "localhost", 20, nil, func(text string) {
        host = text
    })
    form.AddInputField("Port", "6379", 20, nil, func(text string) {
        port = text
    })
    
    // Create Redis connection instance
    redis := windows.NewRedisConnection()
    
    // Function to refresh key-value display
    refreshData := func() {
        if redis.IsConnected() {
            keys, err := redis.GetAllKeys()
            if err != nil {
                logs.Write([]byte(fmt.Sprintf("Error fetching keys: %v\n", err)))
                return
            }
            
            kvDisplay.Clear()
            for _, key := range keys {
                value, err := redis.GetValue(key)
                if err != nil {
                    logs.Write([]byte(fmt.Sprintf("Error fetching value for %s: %v\n", key, err)))
                    continue
                }
                
                ttl, err := redis.GetTTL(key)
                if err != nil {
                    logs.Write([]byte(fmt.Sprintf("Error fetching TTL for %s: %v\n", key, err)))
                    continue
                }
                
                kvDisplay.Write([]byte(fmt.Sprintf("[yellow]Key:[white] %s\n", key)))
                kvDisplay.Write([]byte(fmt.Sprintf("[yellow]Value:[white] %s\n", value)))
                kvDisplay.Write([]byte(fmt.Sprintf("[yellow]TTL:[white] %v\n", ttl)))
                kvDisplay.Write([]byte("------------------------\n"))
            }
        }
    }
    
    // Handle command input
    cmdInput.SetDoneFunc(func(key tcell.Key) {
        if key != tcell.KeyEnter {
            return
        }
        
        cmd := strings.TrimSpace(cmdInput.GetText())
        if cmd == "" {
            return
        }
        
        logs.Write([]byte(fmt.Sprintf("> %s\n", cmd)))
        
        result, err := redis.ExecuteCommand(cmd)
        if err != nil {
            logs.Write([]byte(fmt.Sprintf("[red]Error:[white] %v\n", err)))
        } else {
            logs.Write([]byte(fmt.Sprintf("[green]Result:[white] %v\n", result)))
        }
        
        cmdInput.SetText("")
        refreshData()  // Refresh display after command execution
    })
    
    // Add buttons
    form.AddButton("Connect", func() {
        logs.SetText("")
        logs.Write([]byte(fmt.Sprintf("Connecting to %s:%s ...\n", host, port)))
        
        err := redis.Connect(host, port)
        if err != nil {
            logs.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
            return
        }
        
        logs.Write([]byte("Connected!\n"))
        refreshData()  // Initial data load
    })
    
    form.AddButton("Refresh", func() {
        refreshData()
    })
    
    form.AddButton("Quit", func() {
        redis.Close()
        app.Stop()
    })
    
    // Set form styling
    form.SetBorder(true).SetTitle(" Redis Connection ")
    logs.SetBorder(true).SetTitle(" Logs ")
    
    // Add command input and key-value display to center flex
    cmdFlex.AddItem(kvDisplay, 0, 3, false).
        AddItem(cmdInput, 3, 1, true)
    
    // Create the layout with three panels
    flex.AddItem(form, 30, 1, true).
        AddItem(cmdFlex, 0, 2, false).
        AddItem(logs, 30, 1, false)
    
    // Run the application
    if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
        panic(err)
    }
}