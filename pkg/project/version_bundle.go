package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "etcd-backup-operator",
				Description: "Operator that performs backups of the ETCD database for both control planes and tenant clusters.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "etcd-backup-operator",
		Version:    BundleVersion(),
	}
}
