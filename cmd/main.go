package cmd

import (
    "fmt"
    "github.com/rivo/tview"
    "github.com/Amrit02102004/RediCLI/windows"
)

func Func() {
    app := tview.NewApplication()
    
    // Create the main flex container
    flex := tview.NewFlex()
    
    // Create left side form
    form := tview.NewForm()
    
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
        logs.Write([]byte("Showing Logs...\n"))
    })
    
    form.AddButton("Quit", func() {
        redis.Close()
        app.Stop()
    })
    
    // Set form styling
    form.SetBorder(true).SetTitle(" Redis Connection ")
    logs.SetBorder(true).SetTitle(" Logs ")
    
    // Create the layout
    flex.AddItem(form, 0, 1, true).
        AddItem(logs, 0, 1, false)
    
    // Run the application
    if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
        panic(err)
    }
}