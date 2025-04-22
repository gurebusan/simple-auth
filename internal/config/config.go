package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var configPathEnv = os.Getenv("CONFIG_PATH")

type Config struct {
	Env        string     `yaml:"env" env-default:"local"`
	StorageDSN string     `yaml:"storage_dsn" env-required:"true" env:"STORAGE_DSN"`
	Token      Token      `yaml:"token"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Email      Email      `yaml:"email"`
}

type Token struct {
	Secret     string        `yaml:"secret" env-required:"true" env:"JWT_SECRET"`
	AccessTTL  time.Duration `yaml:"access_ttl" env-default:"5m"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env-default:"72h"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:":8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"10s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Email struct {
	From     string `yaml:"from" env-required:"true"`
	SMTPHost string `yaml:"smtp_host" env-required:"true"`
	SMTPPort int    `yaml:"smtp_port" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is not set")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist" + path)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("cannot read config" + err.Error())

	}
	return &cfg
}

func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", configPathEnv, "path to config file")
	flag.Parse()
	return res
}
