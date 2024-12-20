package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	ServiceName            string
	Environment            string
	Version                string
	HTTPPort               string
	HTTPScheme             string
	PostgresHost           string
	PostgresPort           int
	PostgresUser           string
	PostgresPassword       string
	PostgresDatabase       string
	PostgresMaxConnections int32
	RedisURL               string
	RedisHost              string
	RedisPort              string
	RedisPassword          string
	SecretKey              string
	AuthServiceHost        string
	AuthGRPCPort           string
}

func Load() Config {
	// Load environment variables from the .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: No .env file found")
	}

	config := Config{}

	config.ServiceName = cast.ToString(os.Getenv("SERVICE_NAME"))
	config.Environment = cast.ToString(os.Getenv("ENVIRONMENT"))
	config.Version = cast.ToString(os.Getenv("VERSION"))

	config.HTTPPort = cast.ToString(os.Getenv("HTTP_PORT"))
	config.HTTPScheme = cast.ToString(os.Getenv("HTTP_SCHEME"))

	// PostgreSQL Config
	config.PostgresHost = cast.ToString(os.Getenv("POSTGRES_HOST"))
	config.PostgresPort = cast.ToInt(os.Getenv("POSTGRES_PORT"))
	config.PostgresUser = cast.ToString(os.Getenv("POSTGRES_USER"))
	config.PostgresPassword = cast.ToString(os.Getenv("POSTGRES_PASSWORD"))
	config.PostgresDatabase = cast.ToString(os.Getenv("POSTGRES_DATABASE"))
	config.PostgresMaxConnections = cast.ToInt32(cast.ToInt(os.Getenv("POSTGRES_MAX_CONNECTIONS")))

	// Redis Config
	config.RedisURL = cast.ToString(os.Getenv("REDIS_URL"))
	config.RedisHost = cast.ToString(os.Getenv("REDIS_HOST"))
	config.RedisPort = cast.ToString(os.Getenv("REDIS_PORT"))
	config.RedisPassword = cast.ToString(os.Getenv("REDIS_PASSWORD"))

	// Secret Key
	config.SecretKey = cast.ToString(os.Getenv("SECRET_KEY"))

	// Auth Service Config
	config.AuthServiceHost = cast.ToString(os.Getenv("AUTH_SERVICE_HOST"))
	config.AuthGRPCPort = cast.ToString(os.Getenv("AUTH_GRPC_PORT"))

	return config
}
