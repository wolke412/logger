# Logger

### A simple Go logger built on top of Go's log/slog package. This logger outputs logs to the terminal and saves them to daily-partitioned log files.


#### Features

- Structured Logging: Utilizes Go's log/slog for structured logging.
- Terminal Output: Logs are printed to the terminal.
- Daily Log Files: Logs are saved to files partitioned by day.


#### Installation
```bash
 > go get github.com/wolke412/logger
```

#### Usage
```go
package main

import (
	"github.com/wolke412/logger"
	"log/slog"
)

func main() {

	// sets saves directory
	logger.SetPaht("logs")

	// Initialize the logger
	go logget.Init()


	// Any logs ran in your app with log or slog 
	// will be printedd into your terminal and into
	// the file;

	// The library will crrate a new file every day at local 00:00

	// Entries in the terminal are colores;
	// Entries in files are raw

	// ...
	slog.Info("Application started", "version", "1.0.0")
	slog.Error("An error occurred", "error", "something went wrong")
}

```
