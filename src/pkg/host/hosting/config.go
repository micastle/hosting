package hosting

type Configuration interface {
	Get() interface{}
}

type DefaultConfiguration struct {
	Value interface{}
}

func NewDefaultConfiguration(config interface{}) *DefaultConfiguration {
	return &DefaultConfiguration{
		Value: config,
	}
}

func (c *DefaultConfiguration) Get() interface{} {
	return c.Value
}
