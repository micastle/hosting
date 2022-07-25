package main

import (
	"context"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/hosting"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
)

type WebAPIService interface {
	Run()
	Stop(ctx context.Context) error

	Start()
	Print()
}

type DefaultWebAPIService struct {
	logger logger.Logger
	config *Configuration

	dependentComponent Component
}

func NewWebAPIService() *DefaultWebAPIService {
	return &DefaultWebAPIService{}
}

func (ir *DefaultWebAPIService) Initialize(context hosting.ServiceContext) error {
	ir.logger = context.GetLogger()
	ir.logger.Infow("initializing WebAPIService")

	ir.config = dep.GetConfig[Configuration](context)
	ir.dependentComponent = dep.GetComponent[Component](context)

	return nil
}

func (ir *DefaultWebAPIService) Start() {
	ir.logger.Infow("starting WebAPIService")

	go ir.Run()
}
func (ir *DefaultWebAPIService) Run() {
	ir.logger.Infow("WebAPIService running")
	ir.Print()
}
func (ir *DefaultWebAPIService) Stop(ctx context.Context) error {
	ir.logger.Infow("shutting down WebAPIService")

	if ir.logger == nil {
		panic("WebAPIService was shut down already")
	}

	return nil
}

func (ir *DefaultWebAPIService) Print() {
	ir.dependentComponent.Print()
}
