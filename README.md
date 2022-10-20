# process-monitor
process-monitor is a Go language library for observing the life cycle of system processes.

## Usage

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	processwatcher "github.com/merbridge/process-watcher"
)

func main() {
	w := processwatcher.NewProcessWatcher()
	if err := w.Start(); err != nil {
		panic(err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case e := <-w.Events():
			if e.Err != nil {
				panic(e.Err)
			}
			typ := "not-support"
			var obj interface{}
			switch e.GetType() {
			case processwatcher.PROC_EVENT_EXEC:
				typ = "exec"
				obj = e.GetExec()
			case processwatcher.PROC_EVENT_FORK:
				typ = "fork"
				obj = e.GetFork()
			case processwatcher.PROC_EVENT_EXIT:
				typ = "exit"
				obj = e.GetExit()
			}
			fmt.Printf("%s: %+v\n", typ, obj)
		case <-sigs:
			return
		}
	}
}
```

> Nit: Requires a root user to run, you can run as: `go run -exec sudo ./app`
