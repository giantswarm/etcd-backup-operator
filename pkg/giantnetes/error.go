package giantnetes

import "github.com/giantswarm/microerror"

var unableToGetTenantClustersError = microerror.New("unable to get any tenant cluster from either azure, aws or kvm")

func IsUnableToGetTenantClustersError(err error) bool {
	return microerror.Cause(err) == unableToGetTenantClustersError
}

var failedBackupError = microerror.New("backup failed")

func IsFailedBackupError(err error) bool {
	return microerror.Cause(err) == failedBackupError
}
