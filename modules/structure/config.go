package structure

type Config struct {
	inited bool

	Web struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	Settings struct {
		MaxThreads int `yaml:"max_threads"`
	} `json:"settings"`

	Database struct {
		Login    string `yaml:"login"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
}

func (cfg *Config) Init() {
	cfg.inited = true
}

func (cfg *Config) GetInit() bool {
	return cfg.inited
}
