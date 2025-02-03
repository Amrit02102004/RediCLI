package windows

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rivo/tview"
)

type WindowManager struct {
	app       *tview.Application
	logView   *tview.TextView
	extraView *tview.TextView
	form      *tview.Form
	rdb       *redis.Client
	ctx       context.Context
	onError   func(error)
	stopChan  chan struct{}
	logger    *log.Logger
	logFile   *os.File
}

func NewWindowManager(app *tview.Application) *WindowManager {
	// Create log file
	logFilePath := filepath.Join(".", fmt.Sprintf("redis_%s.log", time.Now().Format("20060102_150405")))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error creating log file: %v\n", err)
		return nil
	}

	return &WindowManager{
		app:      app,
		ctx:      context.Background(),
		stopChan: make(chan struct{}),
		logger:   log.New(logFile, "", log.LstdFlags|log.LUTC),
		logFile:  logFile,
		onError: func(err error) {
			fmt.Printf("Error: %v\n", err)
		},
	}
}

func (w *WindowManager) CreateConnectionForm() *tview.Form {
	w.form = tview.NewForm()
	w.form.AddInputField("Host", "localhost", 20, nil, nil)
	w.form.AddInputField("Port", "6379", 10, nil, nil)
	w.form.AddInputField("Password", "", 20, nil, nil)
	w.form.AddButton("Connect", w.handleConnect)
	w.form.AddButton("Quit", func() {
		w.cleanup()
		w.app.Stop()
	})
	
	w.form.SetBorder(true)
	w.form.SetTitle(" Redis Connection ")
	w.form.Box.SetTitleAlign(tview.AlignLeft)

	return w.form
}

func (w *WindowManager) cleanup() {
	close(w.stopChan)
	if w.rdb != nil {
		w.rdb.Close()
		w.rdb = nil
	}
	if w.logFile != nil {
		w.logFile.Sync()
		w.logFile.Close()
	}
}

func (w *WindowManager) handleConnect() {
	w.cleanup()
	w.stopChan = make(chan struct{})

	host := w.form.GetFormItem(0).(*tview.InputField).GetText()
	port := w.form.GetFormItem(1).(*tview.InputField).GetText()
	password := w.form.GetFormItem(2).(*tview.InputField).GetText()

	redisAddr := fmt.Sprintf("http://%s:%s", host, port)
	options := &redis.Options{
		Addr:        redisAddr,
		DB:          0,
		MaxRetries:  3,
		DialTimeout: 5 * time.Second,
	}

	if password != "" {
		options.Password = password
	}

	w.rdb = redis.NewClient(options)
	
	// Log and test connection
	w.logger.Printf("Attempting to connect to Redis at %s", redisAddr)
	
	if _, err := w.rdb.Ping(w.ctx).Result(); err != nil {
		errMsg := fmt.Sprintf("Connection error: %v", err)
		w.logger.Printf("❌ %s", errMsg)
		w.appendLog(errMsg)
		w.rdb.Close()
		w.rdb = nil
		return
	}

	w.logger.Printf("✅ Connected to Redis at %s", redisAddr)
	w.appendLog(fmt.Sprintf("Connected to Redis at %s", redisAddr))
	
	// Configure keyspace notifications
	if _, err := w.rdb.ConfigSet(w.ctx, "notify-keyspace-events", "KEA").Result(); err != nil {
		w.logger.Printf("⚠️ Warning: Could not configure keyspace events: %v", err)
	}

	w.startMonitoring()
	w.updateExtraInfo()
}

func (w *WindowManager) appendLog(message string) {
	if w.logView == nil {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s", timestamp, message)
	
	// Write to file
	if w.logger != nil {
		w.logger.Println(message)
		w.logFile.Sync() // Force write to disk
	}

	// Update UI
	w.app.QueueUpdateDraw(func() {
		currentText := w.logView.GetText(true)
		w.logView.SetText(currentText + logMessage + "\n")
	})
}

func (w *WindowManager) startMonitoring() {
	go func() {
		pubsub := w.rdb.PSubscribe(w.ctx, "__keyspace@0__:*", "__keyevent@0__:*")
		defer pubsub.Close()

		w.logger.Println("Started Redis monitoring")

		for {
			select {
			case <-w.stopChan:
				w.logger.Println("Stopping Redis monitoring")
				return
			default:
				msg, err := pubsub.ReceiveMessage(w.ctx)
				if err != nil {
					if !strings.Contains(err.Error(), "closed") {
						errMsg := fmt.Sprintf("Monitoring error: %v", err)
						w.logger.Printf("❌ %s", errMsg)
						w.appendLog(errMsg)
					}
					return
				}

				key := strings.TrimPrefix(msg.Channel, "__keyspace@0__:")
				if strings.HasPrefix(msg.Channel, "__keyevent@0__:") {
					key = strings.TrimPrefix(msg.Channel, "__keyevent@0__:")
				}

				logMsg := fmt.Sprintf("Key: %s, Operation: %s", key, msg.Payload)
				w.logger.Printf("🔑 %s", logMsg)
				w.appendLog(logMsg)
			}
		}
	}()
}