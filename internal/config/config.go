package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env          string `yaml: "env" env-default: "local" env-required: "true"`
	Storage_Path string `yaml: "storage_path" env-required: "true"`
	HTTP_Server  `yaml: "http_server"`
}

type HTTP_Server struct {
	Adress       string        `yaml: "adress" env-default: "0.0.0.0:8080"`
	TimeOut      time.Duration `yaml: "timeout" env-default: "4s"`
	Idle_Timeout time.Duration `yaml: "idle_timeout" env-default: "60s"`
	User         string        `yaml:"user" env-required:"true"`
	Password     string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	//check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	// fmt.Println("cfg.Env=", cfg.Env)
	// fmt.Println("cfg.StoragePath=", cfg.StoragePath)
	// fmt.Println("cfg.Adress=", cfg.Adress)
	// fmt.Println("cfg.TimeOut=", cfg.TimeOut)
	// fmt.Println("cfg.IdleTimeout=", cfg.IdleTimeout)

	// fmt.Printf("cfg.Env=%v, type=%T\n", cfg.Env, cfg.Env)
	// fmt.Printf("cfg.StoragePath=%v, type=%T\n", cfg.StoragePath, cfg.StoragePath)
	// fmt.Printf("cfg.Adress=%v, type=%T\n", cfg.Adress, cfg.Adress)
	// fmt.Printf("cfg.TimeOut=%v, type=%T\n", cfg.TimeOut, cfg.TimeOut)
	// fmt.Printf("cfg.IdleTimeout=%v, type=%T\n", cfg.IdleTimeout, cfg.IdleTimeout)

	return &cfg

}
