package tiles

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func Watch() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal("tiles: watcher:", err)
    }
    defer watcher.Close()

    if err := watcher.Add("./activities"); err != nil {
        log.Fatal("tiles: watch activities:", err)
    }

    timer := time.NewTimer(0)
    <-timer.C // drain initial tick

    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            ext := strings.ToLower(filepath.Ext(event.Name))
            if (event.Has(fsnotify.Create) || event.Has(fsnotify.Write)) &&
                (ext == ".gpx" || ext == ".fit") {
                log.Printf("tiles: %s changed, rebuild in 3s...", filepath.Base(event.Name))
                timer.Reset(3 * time.Second)
            }
        case <-timer.C:
            if err := Generate(); err != nil {
                log.Printf("tiles: rebuild failed: %v", err)
            } else {
                log.Println("tiles: rebuild complete")
            }
        case err := <-watcher.Errors:
            log.Println("tiles: watcher error:", err)
        }
    }
}