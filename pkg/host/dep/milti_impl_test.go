package dep

import (
	"fmt"
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func TestComponent_multi_implementations(t *testing.T) {
	cm, ctxt := prepareComponentManager(true)
	components, provider := createCollection(ctxt, cm)

	RegisterComponent[Downloader](
		components,
		func(props Properties) interface{} { return props.Get("type") },
		func(comp CompImplCollection) {
			comp.AddImpl("url", NewUrlDownloader)
			comp.AddImpl("blob", NewBlobDownloader)
		},
	)

	for _, Type := range []string{"url", "blob"} {
		downloader := CreateComponent[Downloader](provider, Props(Pair("type", Type)))
		downloader.Download()

		if downloader.GetType() != Type {
			t.Errorf("expected - %s, actual - %s", Type, downloader.GetType())
		}
	}
}

func TestComponent_multi_implementations_evaluator_negative(t *testing.T) {
	cm, ctxt := prepareComponentManager(true)
	components, provider := createCollection(ctxt, cm)

	RegisterComponent[Downloader](
		components,
		func(props Properties) interface{} { return nil },
		func(comp CompImplCollection) {
			comp.AddImpl("url", NewUrlDownloader)
			comp.AddImpl("blob", NewBlobDownloader)
		},
	)

	for _, Type := range []string{"url", "blob"} {
		defer test.AssertPanicContent(t, "evaluated component implementation key should never be nil", "panic content is not expected")

		downloader := CreateComponent[Downloader](provider, Props(Pair("type", Type)))
		downloader.Download()
	}
}

func TestComponent_multi_implementations_key_not_exist(t *testing.T) {
	cm, ctxt := prepareComponentManager(true)
	components, provider := createCollection(ctxt, cm)

	RegisterComponent[Downloader](
		components,
		func(props Properties) interface{} { return props.Get("type") },
		func(comp CompImplCollection) {
			comp.AddImpl("url", NewUrlDownloader)
			comp.AddImpl("blob", NewBlobDownloader)
		},
	)

	defer test.AssertPanicContent(t, "implementation not exist for key not_exist", "panic content is not expected")

	downloader := CreateComponent[Downloader](provider, Props(Pair("type", "not_exist")))
	downloader.Download()
}

func TestComponent_multi_impl_singleton(t *testing.T) {
	cm, ctxt := prepareComponentManager(true)
	components, provider := createCollection(ctxt, cm)

	RegisterComponent[Downloader](
		components,
		func(props Properties) interface{} { return props.Get("type") },
		func(comp CompImplCollection) {
			compType := comp.GetComponentType()
			fmt.Printf("multi-impl componnent type: %s\n", compType.FullName())
			comp.AddSingletonImpl("url", NewUrlDownloader)
			comp.AddImpl("blob", NewBlobDownloader)
		},
	)

	for _, Type := range []string{"url", "blob"} {
		downloader := CreateComponent[Downloader](provider, Props(Pair("type", Type)))
		downloader.Download()

		if downloader.GetType() != Type {
			t.Errorf("expected - %s, actual - %s", Type, downloader.GetType())
		}

		if Type == "url" {
			// check singleton
			singleton := CreateComponent[Downloader](provider, Props(Pair("type", "url")))
			if singleton != downloader {
				t.Error("singleton implementation should not have multiple instances!")
			}
		}
	}
}

type Downloader interface {
	GetType() string
	Download()
}

type UrlDownloader interface {
	Downloader

	Url() string
}
type DefaultUrlDownloader struct {
	props Properties
}

func NewUrlDownloader() UrlDownloader {
	return &DefaultUrlDownloader{
		props: nil,
	}
}
func (ud *DefaultUrlDownloader) GetType() string {
	return "url"
}
func (ud *DefaultUrlDownloader) Download() {
	fmt.Printf("download for type: %v\n", ud.GetType())
}
func (ud *DefaultUrlDownloader) Url() string {
	return "url"
}

type BlobDownloader interface {
	Downloader

	Blob() string
}
type DefaultBlobDownloader struct {
	props Properties
}

func NewBlobDownloader(props Properties) BlobDownloader {
	return &DefaultBlobDownloader{
		props: props,
	}
}
func (bd *DefaultBlobDownloader) GetType() string {
	return bd.props.Get("type").(string)
}
func (bd *DefaultBlobDownloader) Download() {
	fmt.Printf("download for type: %v\n", bd.GetType())
}
func (bd *DefaultBlobDownloader) Blob() string {
	return "blob"
}
