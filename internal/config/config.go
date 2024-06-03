package config

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App  `yaml:"app"`
		HTTP `yaml:"http"`
		GRPC `yaml:"grpc"`
		Log  `yaml:"logger"`
		PG   `yaml:"postgres"`
	}

	App struct {
		Env           string `yaml:"env"            env-default:"local"`
		Name          string `yaml:"name"           env-default:"dav-go"`
		Version       string `yaml:"version"        env-required:"true"      env:"APP_VERSION" `
		CalDAVPrefix  string `yaml:"caldav_prefix"  env-default:"calendars"`
		CardDAVPrefix string `yaml:"carddav_prefix" env-default:"contacts"`
	}

	HTTP struct {
		IP         string        `yaml:"ip"           env-default:"0.0.0.0"`
		Port       string        `yaml:"port"         env-default:"8082"`
		Timeout    time.Duration `yaml:"timeout"      env-default:"4s"`
		IdleTimout time.Duration `yaml:"idle_timeout" env-default:"60s"`
		User       string        `yaml:"user"         env-required:"true"`
		Password   string        `yaml:"password"     env-required:"true"    env:"HTTP_SERVER_PASSWORD"`
		CORS       struct {
			AllowedMethods     []string `yaml:"allowed_methods"`
			AllowedOrigins     []string `yaml:"allowed_origins"`
			AllowCredentials   bool     `yaml:"allow_credentials"`
			AllowedHeaders     []string `yaml:"allowed_headers"`
			OptionsPassthrough bool     `yaml:"options_passthrough"`
			ExposedHeaders     []string `yaml:"exposed_headers"`
			Debug              bool     `yaml:"debug"`
		} `yaml:"cors"`
	}

	GRPC struct {
		IP   string `yaml:"ip"   env-default:"0.0.0.0"`
		Port string `yaml:"port" env-default:"30000"`
	}

	Log struct {
		Level string `yaml:"log_level" env-required:"true" env:"LOG_LEVEL"`
	}

	PG struct {
		PoolMax int    `yaml:"pool_max" env-default:"2"`
		URL     string `                env-required:"true" env:"PG_URL"`
	}
)

const (
	EnvConfigPathName  = "CONFIG-PATH"
	FlagConfigPathName = "config"
)

var (
	configPath string
	instance   *Config
	once       sync.Once
)

// GetConfig returns app configs.
func GetConfig() *Config {
	once.Do(func() {
		flag.StringVar(
			&configPath,
			FlagConfigPathName,
			"./configs/config.yml",
			"this is app config file",
		)
		flag.Parse()

		log.Print("config init")

		if configPath == "" {
			configPath = os.Getenv(EnvConfigPathName)
		}

		if configPath == "" {
			log.Fatal("config path is required")
		}

		instance = &Config{}

		if err := cleanenv.ReadConfig(configPath, instance); err != nil {
			helpText := "Dav-Go - CalDAV+CardDAV Service"
			help, _ := cleanenv.GetDescription(instance, &helpText)
			log.Print(help)
			log.Fatal(err)
		}
	})
	return instance
}
