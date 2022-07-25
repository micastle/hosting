package hosting

type LoadConfigurationMethod func(configFilePath string) interface{}

type ConfigurationBuilder interface {
	SetConfigurationFilePath(configFilePath string)
	SetConfigurationLoader(configLoader LoadConfigurationMethod)
}

type DefaultConfigurationBuilder struct {
	configFilePath string
	configLoader   LoadConfigurationMethod
}

func NewDefaultConfigurationBuilder() *DefaultConfigurationBuilder {
	return &DefaultConfigurationBuilder{
		configFilePath: "",
	}
}

func (cb *DefaultConfigurationBuilder) SetConfigurationFilePath(configFilePath string) {
	cb.configFilePath = configFilePath
}

func (cb *DefaultConfigurationBuilder) SetConfigurationLoader(configLoader LoadConfigurationMethod) {
	cb.configLoader = configLoader
}
