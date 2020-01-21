package storage

import "os"

func IsEnvVariableDefined(varname string) bool {
	value := os.Getenv(varname)
	return len(value) > 0
}
