package main

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	XMPP     XMPPConfig     `yaml:"xmpp"`
	RSS      RSSConfig      `yaml:"rss"`
	Database DatabaseConfig `yaml:"database"`
}

type XMPPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	JID      string `yaml:"jid"`
	Password string `yaml:"password"`
}

type RSSConfig struct {
	PollInterval int    `yaml:"poll_interval"`
	UserAgent    string `yaml:"user_agent"`
}

type DatabaseConfig struct {
	Type     string `yaml:"type"`     // sqlite or postgres
	Path     string `yaml:"path"`     // for sqlite
	Host     string `yaml:"host"`     // for postgres
	Port     int    `yaml:"port"`     // for postgres
	User     string `yaml:"user"`     // for postgres
	Password string `yaml:"password"` // for postgres
	DBName   string `yaml:"dbname"`   // for postgres
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	applyEnvOverrides(&cfg)

	if cfg.RSS.PollInterval == 0 {
		cfg.RSS.PollInterval = 3600
	}

	if cfg.RSS.UserAgent == "" {
		cfg.RSS.UserAgent = "RSSBot/1.0"
	}

	if cfg.Database.Type == "" {
		cfg.Database.Type = "sqlite"
	}

	if cfg.Database.Type == "sqlite" && cfg.Database.Path == "" {
		cfg.Database.Path = "./rss.db"
	}

	if cfg.Database.Type == "postgres" {
		if cfg.Database.Host == "" {
			cfg.Database.Host = "localhost"
		}
		if cfg.Database.Port == 0 {
			cfg.Database.Port = 5432
		}
		if cfg.Database.DBName == "" {
			cfg.Database.DBName = "rssbot"
		}
	}

	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("XMPP_HOST"); v != "" {
		cfg.XMPP.Host = v
	}
	if v := os.Getenv("XMPP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.XMPP.Port = port
		}
	}
	if v := os.Getenv("XMPP_JID"); v != "" {
		cfg.XMPP.JID = v
	}
	if v := os.Getenv("XMPP_PASSWORD"); v != "" {
		cfg.XMPP.Password = v
	}

	if v := os.Getenv("RSS_POLL_INTERVAL"); v != "" {
		if interval, err := strconv.Atoi(v); err == nil {
			cfg.RSS.PollInterval = interval
		}
	}
	if v := os.Getenv("RSS_USER_AGENT"); v != "" {
		cfg.RSS.UserAgent = v
	}

	if v := os.Getenv("DB_TYPE"); v != "" {
		cfg.Database.Type = v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
}
