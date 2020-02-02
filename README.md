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
	"github.com/easyCZ/logrotate"
	"log"
	"os"
	"time"
)

func main() {
	logger := log.New(os.Stderr, "logrotate", log.LstdFlags) // Or any other logger conforming to golang's log.Logger
	writer, err := logrotate.New(logger, logrotate.Options{
		Directory:       "path/to/my/logs/directory",
		MaximumFileSize: 1024 * 1024 * 1024,
		MaximumLifetime: time.Hour,
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


