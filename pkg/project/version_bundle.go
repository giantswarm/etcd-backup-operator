package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "etcd-backup-operator",
				Description: "This operator is installed on the control plane and takes care of creating backups of the control plane and tenant cluster's ETCDs. It uploads the backups to Amazon S3",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "etcd-backup-operator",
		Version:    BundleVersion(),
	}
}
