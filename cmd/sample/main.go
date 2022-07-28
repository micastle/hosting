package main

func main() {
	host := ConfigureHost().Build()

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	host.Run()
}
