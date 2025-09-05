package types

type Service struct {
	Commands []string `yaml:"commands"`
}

type Project struct {
	Services map[string]*Service `yaml:"services"`
}

type Config struct {
	Projects map[string]*Project `yaml:"projects"`
}
