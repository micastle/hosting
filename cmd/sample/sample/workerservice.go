package main

import (
	"context"
	"fmt"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/hosting"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
)

type WorkerService interface {
	Run()
	Stop(ctx context.Context) error

	Start()
	Print()
}

type DefaultWorkerService struct {
	logger logger.Logger
	config *Configuration

	dependentComponent Component

	context      hosting.ServiceContext
	scopeFactory dep.ScopeFactory
	workerFunc   dep.FreeStyleScopeActionMethod
}

func NewWorkerService(scopeFactory dep.ScopeFactory) *DefaultWorkerService {
	return &DefaultWorkerService{
		scopeFactory: scopeFactory,
	}
}

func (ir *DefaultWorkerService) Initialize(context hosting.ServiceContext, workerFunc dep.FreeStyleScopeActionMethod) error {
	ir.context = context
	ir.logger = context.GetLogger()
	ir.logger.Infow("initializing WorkerService")

	ir.config = dep.GetConfig[Configuration](context)
	ir.dependentComponent = dep.GetComponent[Component](context)

	ir.workerFunc = workerFunc

	return nil
}

func (ir *DefaultWorkerService) Start() {
	ir.logger.Infow("starting WorkerService")

	go ir.Run()
}

func (ir *DefaultWorkerService) Run() {
	ir.logger.Infow("WorkerService running")
	for {
		iteration := dep.GetComponent[LoopIteration](ir.context)
		fmt.Printf("LoopIteration scope instance created: %p\n", iteration)
		ir.executeIteration(iteration)

		time.Sleep(time.Duration(5) * time.Second)
	}
}

type AnonScopeFunc dep.FreeStyleScopeActionMethod

func (ir *DefaultWorkerService) executeIteration(iter LoopIteration) {
	scope := dep.CreateScopeFrom[LoopIteration](ir.scopeFactory, iter)
	defer func() { scope.Dispose() }()
	dep.Using[LoopIteration](iter, ir.workerFunc)
}

func (ir *DefaultWorkerService) Stop(ctx context.Context) error {
	ir.logger.Infow("shutting down WorkerService")

	if ir.logger == nil {
		panic("WorkerService was shut down already")
	}

	return nil
}

func (ir *DefaultWorkerService) Print() {
	ir.dependentComponent.Print()
}
