package config

type RunEnv string

const (
	Local   RunEnv = "local"
	Dev     RunEnv = "dev"
	Test    RunEnv = "test"
	Prod    RunEnv = "prod"
	Unknown RunEnv = ""
)

func GetRunEnv() RunEnv {
	v := RunEnv(getEnvVar("RUN_ENV"))
	switch v {
	case Local, Dev, Test, Prod:
		return v
	default:
		panic("invalid RUN_ENV")
		return Unknown
	}
}
