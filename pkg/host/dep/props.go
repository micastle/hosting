package dep

import (
	"fmt"
	"sort"
	"strings"
)

type Properties interface {
	Count() int
	Keys() []string
	Get(key string) any
	Has(key string) bool

	Set(key string, value any)
	Update(Properties)

	String() string
}

func GetProp[T any](props Properties, key string) T {
	return props.Get(key).(T)
}
func SetProp[T any](props Properties, key string, value T) {
	props.Set(key, value)
}

type DefaultProperties struct {
	props map[string]any
}

func NewProperties() *DefaultProperties {
	return &DefaultProperties{
		props: make(map[string]any),
	}
}
func NewPropertiesFrom(props Properties) *DefaultProperties {
	p := NewProperties()
	p.Update(props)
	return p
}
func Props(pairs ...*PropertyPair) *DefaultProperties {
	props := NewProperties()
	for _, pair := range pairs {
		props.Set(pair.Key, pair.Value)
	}
	return props
}

func (p *DefaultProperties) Update(props Properties) {
	if props == nil {
		return
	}
	for _, key := range props.Keys() {
		p.Set(key, props.Get(key))
	}
}
func (p *DefaultProperties) String() string {
	props := make([]string, 0, len(p.props))
	for key, val := range p.props {
		prop := fmt.Sprintf("%s=%v", key, val)
		props = append(props, prop)
	}
	sort.Strings(props)
	return "{" + strings.Join(props, ",") + "}"
}
func (p *DefaultProperties) Count() int {
	return len(p.props)
}
func (p *DefaultProperties) Keys() []string {
	keys := make([]string, 0, len(p.props))
	for key, _ := range p.props {
		keys = append(keys, key)
	}
	return keys
}
func (p *DefaultProperties) Has(key string) bool {
	_, exist := p.props[key]
	return exist
}
func (p *DefaultProperties) Get(key string) any {
	val, exist := p.props[key]
	if !exist {
		panic(fmt.Errorf("property \"%s\" not exist", key))
	}
	return val
}
func (p *DefaultProperties) Set(key string, val any) {
	p.props[key] = val
}

type PropertyPair struct {
	Key   string
	Value any
}

func Pair(key string, val any) *PropertyPair {
	return &PropertyPair{
		Key:   key,
		Value: val,
	}
}
