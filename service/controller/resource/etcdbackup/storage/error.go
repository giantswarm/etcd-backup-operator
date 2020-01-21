package storage

import (
	"fmt"
	"github.com/giantswarm/microerror"
)

func newMissingRequiredEnvVariableError(varname string) *microerror.Error {
	return &microerror.Error{
		Desc: fmt.Sprintf("Required EnvironmentVariable %s is not defined", varname),
		Kind: "missingRequiredEnvVariableError",
	}
}
