# Hello, world!

Below is a simple console application, which execute a loop named "HelloLoop" that prints "Hello, world!" every 5 seconds.

```go
type LoopConfig struct {
    IntervalInSec int
}

func main() {
    hostBuilder := hosting.NewDefaultHostBuilder()
    hostBuilder.SetHostName("Sample")
    hostBuilder.ConfigureHostConfiguration(func() *LoopConfig {return &LoopConfig{IntervalInSec: 5}})
    hostBuilder.UseLoopEx("HelloLoop", func(config *LoopConfig, looper hosting.ConfigureLoopContext) {
		looper.SetInterval(time.Duration(config.IntervalInSec) * time.Second)
		looper.RegisterProcessorEx(func(log Logger){log.Info("Hello, world!")})
	})

    host := hostBuilder.Build()
    host.Run()
}
```

