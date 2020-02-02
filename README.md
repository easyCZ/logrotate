[![Build Status](https://cloud.drone.io/api/badges/easyCZ/logrotate/status.svg)](https://cloud.drone.io/easyCZ/logrotate)

# logrotate
Concurrency safe logging writer for Golang for application-side log rotation.

## Motivation
In general, application-side log rotation should be avoided. Log rotation is best done through [logrotate](https://linux.die.net/man/8/logrotate) and other tools. 
However, some applications are constrained to only application-side log rotation and benefit from this package.

## Usage
```
go get github.com/easyCZ/logrotate
```

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"
    
    "github.com/easyCZ/logrotate"
)

func main() {
	logger := log.New(os.Stderr, "logrotate", log.LstdFlags) // Or any other logger conforming to golang's log.Logger
	writer, err := logrotate.New(logger, logrotate.Options{
        // Where should the writer be outputting files?
        // If the directory does not exist, it will be created.
        // Required.
		Directory:       "path/to/my/logs/directory",
        // What is the maximum size of each file?
        // Optional. Use 0 for unlimited.
		MaximumFileSize: 1024 * 1024 * 1024,
        // How often should a new file be created, based on time?
        // Optional. Use 0 to disable time based log rotation.
		MaximumLifetime: time.Hour,
        // How would you like to name your files?
        // Invoked each time a new file is being created.
        FileNameFunc:    logrotate.DefaultFilenameFunc,
	})
	if err != nil {
		// handle err
	}
    
    // ...
    
    // Ensure all messages are flushed to files before exiting 
    if err := writer.Close(); err != nil {
        // handle err
    }   
}
```

### Usage with Golang's log.Logger
```go
package main

import (
	"log"
	"os"

	"github.com/easyCZ/logrotate"
)

func main() {
	// Write logrotate's own log lines to stderr
	stderrLogger := log.New(os.Stderr, "logrotate", log.LstdFlags)

	writer, err := logrotate.New(stderrLogger, logrotate.Options{
		Directory: "/path/to/my/logs",
		// see other options above
	})

	// Your application logger, using logrotate as the backing Writer implementation.
	logger := log.New(writer, "application", log.LstdFlags)

    // Ensure all messages are flushed to files before exiting 
    if err := writer.Close(); err != nil {
        // handle err
    }  
}
```

### Usage with logrus
```go
package main

import (
	"os"
	log "github.com/sirupsen/logrus"
	logrotate "github.com/easyCZ/logrotate"
)

func init() {
	writer, err := logrotate.New(logger, logrotate.Options{
		Directory: "path/to/my/logs/directory",
	})
	if err != nil {
		// handle err
	}
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(writer)
}
```

