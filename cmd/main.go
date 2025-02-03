package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Amrit02102004/RediCLI/windows"
	"github.com/rivo/tview"
)

func setupLogging() *os.File {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create log file with timestamp
	logFileName := fmt.Sprintf("redis_monitor_%s.log", time.Now().Format("20060102_150405"))
	logFilePath := filepath.Join("logs", logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}

	// Set up multi-writer to write logs to both file and stdout
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Printf("📝 Logging to: %s\n", logFilePath)
	return logFile
}

func Main() {
	// Setup logging
	logFile := setupLogging()
	defer logFile.Close()

	log.Println("Starting Redis Monitor Application")

	// Create application
	app := tview.NewApplication()
	windowManager := windows.NewWindowManager(app)

	if windowManager == nil {
		log.Fatal("Failed to create window manager")
		return
	}

	// Create UI components
	log.Println("Creating UI components")
	logView := windowManager.CreateLogView()
	extraView := windowManager.CreateExtraView()
	connectionForm := windowManager.CreateConnectionForm()

	// Set custom error handler
	// windowManager.SetErrorHandler(func(err error) {
	// 	log.Printf("Application error: %v", err)
	// 	app.QueueUpdateDraw(func() {
	// 		currentText := logView.GetText(true)
	// 		errorMsg := fmt.Sprintf("%s\n❌ Error: %v", currentText, err)
	// 		logView.SetText(errorMsg)
	// 	})
	// })

	// Create layout
	log.Println("Setting up application layout")
	leftColumn := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(logView, 0, 1, false).
		AddItem(connectionForm, 7, 1, true)

	mainLayout := tview.NewFlex().
		AddItem(leftColumn, 0, 2, true).
		AddItem(extraView, 0, 1, false)

	// Add application-level keyboard handling
	// app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	switch event.Key() {
	// 	case tcell.KeyCtrlC:
	// 		log.Println("Received Ctrl+C, shutting down...")
	// 		app.Stop()
	// 		return nil
	// 	}
	// 	return event
	// })

	// Run application
	log.Println("Starting application UI")
	if err := app.SetRoot(mainLayout, true).EnableMouse(true).Run(); err != nil {
		log.Printf("Application crashed: %v", err)
		panic(err)
	}

	log.Println("Application shutting down gracefully")
}