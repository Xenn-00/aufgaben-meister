package config

import (
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type AppConfig struct {
	APP struct {
		Name  string `mapstructure:"NAME"`
		Port  string `mapstructure:"PORT"`
		State string `mapstructure:"STATE"`
	}

	DATABASE struct {
		Postgres struct {
			DSN string `mapstructure:"DSN"`
		}
		Redis struct {
			Addr     string `mapstructure:"ADDR"`
			Password string `mapstructure:"PASSWORD"`
		}
	}

	APP_SECRET struct {
		Paseto struct {
			HexKey string `mapstructure:"HEX_KEY"`
		}
	}

	MAILTRAP struct {
		Sandbox struct {
			SandboxHost   string `mapstructure:"SANDBOX_HOST"`
			SandboxAPI    string `mapstructure:"SANDBOX_API"`
			SandboxURL    string `mapstructure:"SANDBOX_URL"`
			SandboxDomain string `mapstructure:"SANDBOX_DOMAIN"`
		}
		API struct {
			APIToken         string `mapstructure:"API_TOKEN"`
			APIHost          string `mapstructure:"API_HOST"`
			MailtrapTokenAPI string `mapstructure:"MAILTRAP_TOKEN_API"`
			MailtrapURL      string `mapstructure:"MAILTRAP_URL"`
			MailtrapDomain   string `mapstructure:"MAILTRAP_DOMAIN"`
		}
	}
}

func LoadConfig() *AppConfig {
	viper.SetConfigName("application")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Fehler beim Lesen der Konfigurationsdatei")
		return nil
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		log.Error().Err(err).Msg("Fehler beim Entpacken der Konfiguration")
		return nil
	}

	if config.APP.Port == "" {
		config.APP.Port = "8080"
	}

	if config.DATABASE.Postgres.DSN == "" {
		log.Error().Msg("Datenbank-DSN ist nicht konfiguriert")
		return nil
	}

	if config.APP_SECRET.Paseto.HexKey == "" {
		config.APP_SECRET.Paseto.HexKey = utils.GenerateSymmetricKey()
	}

	log.Info().Msg("Konfiguration geladen...")
	return &config
}
