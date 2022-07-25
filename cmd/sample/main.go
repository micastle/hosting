package main

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
)

func main() {
	host := ConfigureHost().Build()

	provider := host.GetComponentProvider()
	lf := dep.GetComponent[logger.LoggerFactory](provider)
	logger := lf.GetDefaultLogger()
	logger.Debugw("logging system ready")

	logger = host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	host.Run()
}
