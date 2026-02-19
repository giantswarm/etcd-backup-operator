package giantnetes

import (
	"crypto/tls"

	"github.com/giantswarm/etcd-backup-operator/v5/pkg/etcd/proxy"
)

type ETCDv3Settings struct {
	Endpoints string
	Proxy     *proxy.Proxy
	TLSConfig *tls.Config
}

type ETCDInstance struct {
	Name   string
	ETCDv3 ETCDv3Settings
}

type TLSClientConfig struct {
	CAData  []byte
	KeyData []byte
	CrtData []byte
}

func (s ETCDv3Settings) AreComplete() bool {
	return s.Endpoints != "" && s.TLSConfig != nil
}
