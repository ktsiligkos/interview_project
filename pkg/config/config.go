package config

import (
	"os"
	"strings"
)

// Config represents the runtime configuration values for the service.
type Config struct {
	HTTPAddr     string
	MySQLDSN     string
	JWTSecret    string
	KafkaBrokers []string
	KafkaTopic   string
}

// Load reads configuration from the environment, applying sane defaults.
func Load() (Config, error) {
	// configuration file exists
	// read the contents

	// no file present
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	connString := os.Getenv("MYSQL_DSN")
	if connString == "" {
		connString = "root:password@tcp(127.0.0.1:3306)/xm_companies"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret1234"
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	var brokers []string
	if kafkaBrokers == "" {
		brokers = []string{"localhost:9094"}
	} else {
		for _, b := range strings.Split(kafkaBrokers, ",") {
			if trimmed := strings.TrimSpace(b); trimmed != "" {
				brokers = append(brokers, trimmed)
			}
		}
		if len(brokers) == 0 {
			brokers = []string{"localhost:9094"}
		}
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "company-events"
	}

	return Config{
		HTTPAddr:     addr,
		MySQLDSN:     connString,
		JWTSecret:    secret,
		KafkaBrokers: brokers,
		KafkaTopic:   kafkaTopic,
	}, nil
}
