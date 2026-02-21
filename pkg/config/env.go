package config

import "strings"

type Env string

const (
	EnvDev     Env = "dev"
	EnvStaging Env = "staging"
	EnvProd    Env = "prod"
)

const EnvKey = "APP_ENV"

func CurrentEnv() Env {
	v := strings.ToLower(String(EnvKey, string(EnvDev)))
	switch Env(v) {
	case EnvDev, EnvStaging, EnvProd:
		return Env(v)
	default:
		return EnvDev
	}
}

func IsProd() bool {
	return CurrentEnv() == EnvProd
}
