package windows

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/rivo/tview"
)

type ConnectionConfig struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port string `json:"port"`
}

func getConnectionsFilePath() string {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return filepath.Join(".redicli", "connections.json")
    }
    return filepath.Join(homeDir, ".redicli", "connections.json")
}

func saveConnection(config ConnectionConfig) error {
	// Ensure .redicli directory exists
	filePath := getConnectionsFilePath()
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Read existing connections
	var connections []ConnectionConfig
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &connections); err != nil {
			return err
		}
	}

	// Check for duplicate names
	for _, conn := range connections {
		if conn.Name == config.Name {
			return fmt.Errorf("connection with name '%s' already exists", config.Name)
		}
	}

	// Append new connection
	connections = append(connections, config)

	// Write back to file
	updatedData, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, updatedData, 0644)
}

func GetConnections() ([]ConnectionConfig, error) {
	filePath := getConnectionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ConnectionConfig{}, nil
		}
		return nil, err
	}

	var connections []ConnectionConfig
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, err
	}

	return connections, nil
}

func RefreshData(logDisplay *tview.TextView, kvDisplay *tview.TextView, redis *utils.RedisConnection) {
	if redis.IsConnected() {
		_, err := redis.GetAllKeys()
		if err != nil {
			logDisplay.Write([]byte(fmt.Sprintf("Error fetching keys: %v\n", err)))
			return
		}
	}
}

func ConnectionForm(app *tview.Application, logDisplay *tview.TextView, redis *utils.RedisConnection, kvDisplay *tview.TextView) tview.Primitive {
	// Create the form
	form := tview.NewForm()

	var name, host, port string
	
	form.AddInputField("Connection Name*", "", 18, nil, func(text string) {
		name = text
	})
	
	form.AddInputField("Host/URL*    ", "", 18, nil, func(text string) {
		host = text
	})
	
	form.AddInputField("Port   ", "", 18, nil, func(text string) {
		port = text
	})

	// Create Flex layout
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add the form items directly into Flex
	flex.AddItem(form, 0, 1, false)

	// Add other components (buttons, text views)
	form.AddButton("Save & Connect", func() {
		logDisplay.SetText("")
		
		// Validate inputs
		if name == "" {
			logDisplay.Write([]byte("[red]Error: Connection Name is required[white]\n"))
			return
		}
		
		// Default to localhost if no input
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "6379"
		}
		
		// Create connection config
		config := ConnectionConfig{
			Name: name,
			Host: host,
			Port: port,
		}
		
		// Save connection
		err := saveConnection(config)
		if err != nil {
			logDisplay.Write([]byte(fmt.Sprintf("[red]Error saving connection: %v[white]\n", err)))
			return
		}
		
		// Attempt to connect
		err = redis.Connect(host, port)
		if err != nil {
			logDisplay.Write([]byte(fmt.Sprintf("[red]Connection failed: %v[white]\n", err)))
			return
		}
		
		logDisplay.Write([]byte(fmt.Sprintf("[green]Connection '%s' saved and connected successfully[white]\n", name)))
	})

	// Set up the overall layout for the application
	flex.SetBorder(true).SetTitle(" Add Redis Connection ")
	
	return flex
}