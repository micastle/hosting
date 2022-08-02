package main

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/hosting"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type MyFuncProc hosting.FunctionProcessor

func ConfigureComponents(context hosting.BuilderContext, components dep.ComponentCollection) {
	// configure component configurations (other than host & app configuration)
	//components.AddConfiguration(&context.GetHostConfiguration().(*Configuration).Log)

	// configure dependent components
	dep.RegisterSingleton[Component](components, NewComponent)
	dep.RegisterTransient[MyProc](components, NewMyProcessor)

	dep.RegisterTransient[LoopIteration](components, NewLoopIteration)
	dep.RegisterTransient[MyHandler](components, NewHandler)

	dep.RegisterScoped[ScopedComp, LoopIteration](components, NewScopedComp)
	dep.RegisterScoped[HandleComp, any](components, NewHandleComp)

	// func processor
	hosting.RegisterFuncProcessor[MyFuncProc](components, func(context dep.Context, scopeCtxt hosting.ScopeContext, component Component) {
		context.GetLogger().Infof("MyFuncProcessor called, context of looper %s: %s-%s", scopeCtxt.GetLoopRunContext().LooperName(), context.Type(), context.Name())
		component.Print()
	})

	dep.RegisterComponent[Downloader](components,
		func(props dep.Properties) string { return dep.GetProp[string](props, "type") },
		func(comp dep.CompImplCollection[string]) {
			comp.AddImpl("url", NewUrlDownloader)
			comp.AddImpl("blob", NewBlobDownloader)
		},
	)
}

type LoopIteration interface {
	dep.Scopable
	IterationId() string
}
type DefaultLoopIteration struct {
	context dep.Context
}

func NewLoopIteration(context dep.Context) *DefaultLoopIteration {
	return &DefaultLoopIteration{
		context: context,
	}
}
func (li *DefaultLoopIteration) Context() dep.Context {
	return li.context
}
func (li *DefaultLoopIteration) IterationId() string {
	return "123"
}

type MyHandler interface {
	dep.Scopable
	GetName() string
}
type DefaultHandler struct {
	context dep.Context
}

func NewHandler(context dep.Context) *DefaultHandler {
	return &DefaultHandler{context: context}
}
func (dh *DefaultHandler) Context() dep.Context {
	return dh.context
}
func (dh *DefaultHandler) GetName() string {
	return "Default"
}

type HandleComp interface {
	Print()
}
type DefaultHandleComp struct {
	logger logger.Logger
}

func NewHandleComp(context dep.Context, logger logger.Logger, comp ScopedComp) *DefaultHandleComp {

	comp.Print()

	return &DefaultHandleComp{
		logger: logger,
	}
}
func (hc *DefaultHandleComp) Print() {
	hc.logger.Infof("handler component: %p", hc)
}

type Downloader interface {
	Download()
}

type UrlDownloader interface {
	Downloader

	Url() string
}
type DefaultUrlDownloader struct {
	props dep.Properties
}

func NewUrlDownloader() UrlDownloader {
	return &DefaultUrlDownloader{
		props: nil,
	}
}
func (ud *DefaultUrlDownloader) Download() {
	//fmt.Printf("download for type: %v\n", ud.props.Get("type"))
}
func (ud *DefaultUrlDownloader) Url() string {
	return "url"
}

type BlobDownloader interface {
	Downloader

	Blob() string
}
type DefaultBlobDownloader struct {
	props dep.Properties
}

func NewBlobDownloader(props dep.Properties) BlobDownloader {
	return &DefaultBlobDownloader{
		props: props,
	}
}
func (bd *DefaultBlobDownloader) Download() {
	fmt.Printf("download for type: %v\n", bd.props.Get("type"))
}
func (bd *DefaultBlobDownloader) Blob() string {
	return "blob"
}

type Component interface {
	Print()
}

type DefaultComponent struct {
	context dep.Context
	logger  logger.Logger
	config  *Configuration

	downloader Downloader
}

func NewComponent(context dep.Context, config *Configuration, host hosting.Host) *DefaultComponent {
	comp := &DefaultComponent{
		context:    context,
		logger:     context.GetLogger(),
		config:     config,
		downloader: dep.CreateComponent[Downloader](context, dep.Props(dep.Pair("type", "url"))),
	}

	comp.logger.Debug("created component: ", types.Of(comp).FullName())

	return comp
}

func (c *DefaultComponent) Print() {
	c.logger.Info("component print a log line")

	c.downloader.Download()
}

type ScopedComp interface {
	Print()
}
type DefaultScopedComp struct {
	context dep.Context
	logger  logger.Logger
}

func NewScopedComp(context dep.Context, logger logger.Logger) *DefaultScopedComp {
	return &DefaultScopedComp{
		context: context,
		logger:  logger,
	}
}

func (c *DefaultScopedComp) Print() {
	c.logger.Infof("scoped component: %p", c)
}
