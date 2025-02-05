# RediCLI

![Image](https://github.com/user-attachments/assets/533962e4-c3cb-4190-96ff-8c472c3b6060)

A powerful Terminal User Interface (TUI) for Redis, built with Go. RediCLI provides an intuitive, interactive way to manage your Redis databases with features like command suggestions, data import/export, and connection management.

## Features

- 🚀 Interactive terminal interface
- 💡 Command auto-completion and suggestions
- 📊 Built-in analytics dashboard
- 📁 Import/Export data from CSV and XLSX files
- 🔑 Advanced key management with TTL support
- 🔄 Connection management for multiple Redis instances
- 🎨 Color-coded interface for better readability

## Installation

### From Releases

Download the latest binary for your platform from the [releases page](https://github.com/yourusername/RediCLI/releases):

- Linux: `redicli-linux`
- macOS (Intel): `redicli-mac`
- macOS (Apple Silicon): `redicli-mac-arm64`
- Windows: `redicli.exe`

Make the binary executable (Unix-based systems):
```bash
chmod +x redicli-*
```

### From Source

```bash
# Clone the repository
git clone https://github.com/Amrit02102004/RediCLI.git

# Navigate to the project directory
cd RediCLI

# Build the project
go build

# Run RediCLI
./redicli
```

### Using Snap (Ubuntu/Linux)

```bash
snap install redicli
```

## Available Commands

### Basic Commands

- `get <key>` - Retrieve the value of a key
- `set <key> <value>` - Set the string value of a key
- `del <key>` - Delete a key
- `keys <pattern>` - Find all keys matching a pattern
- `ttl <key>` - Get the time to live for a key
- `expire <key> <seconds>` - Set a key's time to live in seconds

### Advanced Commands

- `key filter set` - Open form to set a key with TTL in milliseconds
- `key filter update` - Open form to update a key with KEEPTTL option
- `flushall` - Delete all keys (use with caution)
- `see analytics` - Open analytics dashboard in browser

### Data Management

- `import` - Import data from CSV/XLSX file
- `export` - Export data to CSV file
- `import ./path/to/file.csv` - Direct import from file path
- `export ./path/to/file.csv` - Direct export to file path

### Connection Management

- `add connection` - Add and connect to a new Redis instance
- `view all connections` - List all saved Redis connections
- `connect <name>` - Connect to a saved Redis connection
- `del connection <name>` - Delete a specific saved connection
- `del all connections` - Delete all saved connections

### Interface Commands

- `clear all` - Clear console and logs screen
- `clear logs` - Clear logs screen
- `clear display` - Clear display screen
- `help` - Display help information
- `quit` - Exit RediCLI

## Keyboard Shortcuts

- `Tab` - Cycle through command suggestions
- `↑/↓` - Navigate command history
- `Enter` - Execute command

## Development

### Prerequisites

- Go 1.23 or higher
- Redis server

### Dependencies

```go
require (
    github.com/gdamore/tcell/v2 v2.7.1
    github.com/gorilla/websocket v1.5.3
    github.com/lithammer/fuzzysearch v1.1.8
    github.com/redis/go-redis/v9 v9.7.0
    github.com/rivo/tview v0.0.0-20241227133733-17b7edb88c57
    github.com/xuri/excelize/v2 v2.9.0
)
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
