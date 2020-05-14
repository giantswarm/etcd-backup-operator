package giantnetes

import "github.com/giantswarm/microerror"

// executionFailedError should never be matched against and therefore there is
// no matcher implement. For further information see:
//
//     https://github.com/giantswarm/fmt/blob/master/go/errors.md#matching-errors
//
var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

var failedBackupError = &microerror.Error{
	Kind: "failedBackupError",
}

// IsFailedBackup asserts failedBackupError.
func IsFailedBackup(err error) bool {
	return microerror.Cause(err) == failedBackupError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var unableToGetTenantClustersError = &microerror.Error{
	Kind: "unableToGetTenantClustersError",
}

// IsUnableToGetTenantClusters asserts unableToGetTenantClustersError.
func IsUnableToGetTenantClusters(err error) bool {
	return microerror.Cause(err) == unableToGetTenantClustersError
}
