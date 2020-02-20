package giantnetes

import "github.com/giantswarm/microerror"

var failedBackupError = microerror.New("backup failed")

func IsFailedBackupError(err error) bool {
	return microerror.Cause(err) == failedBackupError
}

var unableToGetTenantClustersError = microerror.New("unable to get any tenant cluster")

func IsUnableToGetTenantClustersError(err error) bool {
	return microerror.Cause(err) == unableToGetTenantClustersError
}
