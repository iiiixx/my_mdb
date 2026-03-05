package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr string

	DBURL string

	OMDbAPIKey        string
	OMDbBaseURL       string
	OMDbTimeout       time.Duration
	OMDbTimeoutForJob time.Duration

	RecGRPCAddr string
	RecTimeout  time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),

		DBURL: getEnv("DB_URL", "postgres://postgres:postgres@localhost:5433/mdb?sslmode=disable"),

		OMDbAPIKey:        getEnv("OMDB_API_KEY", ""),
		OMDbBaseURL:       getEnv("OMDB_BASE_URL", "https://www.omdbapi.com/"),
		OMDbTimeout:       time.Duration(getEnvInt("OMDB_TIMEOUT_SEC", 5)) * time.Second,
		OMDbTimeoutForJob: time.Duration(getEnvInt("OMDB_TIMEOUT_FOR_JOB_SEC", 15)) * time.Second,

		RecGRPCAddr: getEnv("REC_GRPC_ADDR", "localhost:50051"),
		RecTimeout:  time.Duration(getEnvInt("REC_TIMEOUT_SEC", 3)) * time.Second,
	}
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
