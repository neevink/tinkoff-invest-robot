package robot

type Config struct {
	TinkoffApiEndpoint string `yaml:"tinkoff_api_endpoint"`
	AccessToken        string `yaml:"access_token"`
}

func NewConfig() *Config {
	return &Config{
		TinkoffApiEndpoint: "",
		AccessToken:        "",
	}
}
