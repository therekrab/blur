package errorhandling

import "sync"

type errorWatcher struct {
    mu sync.Mutex
    hadError bool
}

var once sync.Once
var instance *errorWatcher

func getWatcher() *errorWatcher {
    once.Do(func() {
        instance = &errorWatcher{}
    })
    return instance
}

func hadError() bool {
    watcher := getWatcher()
    watcher.mu.Lock()
    defer watcher.mu.Unlock()
    return watcher.hadError
}

func fatalError() {
    watcher := getWatcher()
    watcher.mu.Lock()
    defer watcher.mu.Unlock()
    watcher.hadError = true
}
