package structure

type Config struct {
	inited bool

	Web struct {
		Port             int  `yaml:"port"`
		WebSocketAllowed bool `yaml:"websocket_support"`
	} `yaml:"server"`

	Settings struct {
		MaxThreads int `yaml:"max_threads"`
		MaxData    int `yaml:"max_data"`

		TasksCount   int  `yaml:"worker_tasks"`
		CheckUpdates bool `yaml:"check_updates"`
	} `json:"settings"`

	Distribute struct {
		StartFrom int `yaml:"from"`
		GiveMax   int `yaml:"give_at_one_call"`
	} `yaml:"distribute"`

	Garbage struct {
		Enabled  bool `yaml:"enabled"`
		Interval int  `yaml:"interval"`
	} `yaml:"garbage"`

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
