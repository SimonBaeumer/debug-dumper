# Debug Watcher

Debug watcher lists all Pods metrics (in this case for Pods with `app.kubernetes.io/component=central`).

```go
// Create and start the debug dumper
ctx, cancelFn := context.WithCancel(context.Background())
dumper := monitoring.Dumper{
    DebugInfoGetter:  debughandlers.DebugInfoGetterFileSystem,
    MetricsClientset: metricsClientset,
    Clientset:        clientset,
    MemoryThreshold:  MemoryThreshold,
    Interval:         5 * time.Minute,
}
go dumper.Watch(ctx)
```
