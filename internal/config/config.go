package config

type Config struct {
	AppName  string         `env:"APP_NAME"`
	AppEnv   string         `env:"APP_ENV"`
	Port     int            `env:"PORT"`
	Postgres PostgresConfig `env:"POSTGRES"`
	Logger   LoggerConfig   `env:"LOGGER"`
	Kafka    KafkaConfig    `env:"KAFKA"`
	GRPC     GRPCConfig     `env:"GRPC"`
}

type PostgresConfig struct {
	PostgresHost     string `env:"POSTGRES_HOST"`
	PostgresPort     int    `env:"POSTGRES_PORT"`
	PostgresDB       string `env:"POSTGRES_DB"`
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
}

type LoggerConfig struct {
	LogLevel  string `env:"LOG_LEVEL"`
	LogFormat string `env:"LOG_FORMAT"`
}

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS"`
	Topic   string   `env:"KAFKA_TOPIC"`
}

type GRPCConfig struct {
	LedgerPort string `env:"GRPC_LEDGER_PORT"`
}
