package types

type Service struct {
	Name     string   `yaml:"service"`
	Commands []string `yaml:"commands"`
}

type Project struct {
	Name     string    `yaml:"project"`
	Services []Service `yaml:"services"`
}

type Config struct {
	Projects []Project `yaml:"projects"`
}
