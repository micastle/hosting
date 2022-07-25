# Console Application

Below is a simple console application, which execute a loop named "HelloLoop" that prints "Hello, world!" every 5 seconds.

```go
func main() {
    hostBuilder := hosting.NewDefaultHostBuilder().SetHostName("Sample")
    hostBuilder.UseLoopEx("HelloLoop", func(looper hosting.ConfigureLoopContext) {
		looper.SetInterval(time.Duration(5) * time.Second)
		looper.RegisterProcessorEx(func(){fmt.Println("Hello, world!")})
	})

    host := hostBuilder.Build()
    host.Run()
}
```

