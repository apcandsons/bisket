package internal

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port       string     `yaml:"port"`
	AdminPort  string     `yaml:"admin_port"`
	Preview    bool       `yaml:"preview"`
	RunCommand []string   `yaml:"run"`
	RepoConfig RepoConfig `yaml:"repository"`
}

type RepoConfig struct {
	Github GithubConfig `yaml:"github"`
}

type GithubConfig struct {
	RepoUrl string `yaml:"repo_url"`
	ApiKey  string `yaml:"api_key"`
}

func (cfg *Config) Init() {
	cfg.Port = "8080"
	cfg.AdminPort = "18080"
	cfg.Preview = false
	cfg.RunCommand = []string{"go run dist/@$(uname -m)/server -p $BISKET_PORT"}
	cfg.RepoConfig = RepoConfig{
		Github: GithubConfig{
			RepoUrl: "https://github.com/apcandsons/echo-app",
			ApiKey:  "",
		},
	}
}

func (cfg *Config) WriteToFile(filename string) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("Failed to marshal config to YAML: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write config to file: %v", err)
	}
}

func (cfg *Config) ReadFromFile(s string) error {
	data, err := os.ReadFile(s)
	if err != nil {
		log.Fatalf("Failed to read config from file: %v", err)
		return err
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Fatalf("Failed to unmarshal config from YAML: %v", err)
		return err
	}
	return nil
}
