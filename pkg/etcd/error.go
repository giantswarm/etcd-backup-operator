package etcd

import "github.com/giantswarm/microerror"

var emptyEndpointHealthError = &microerror.Error{
	Kind: "emptyEndpointHealthError",
}
