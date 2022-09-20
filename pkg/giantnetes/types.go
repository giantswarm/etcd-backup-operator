package giantnetes

import (
	"crypto/tls"

	"github.com/giantswarm/etcd-backup-operator/v4/pkg/etcd/proxy"
)

type ETCDv2Settings struct {
	DataDir string
}

type ETCDv3Settings struct {
	Endpoints string
	Proxy     *proxy.Proxy
	TLSConfig *tls.Config
}

type ETCDInstance struct {
	Name   string
	ETCDv2 ETCDv2Settings
	ETCDv3 ETCDv3Settings
}

type TLSClientConfig struct {
	CAData  []byte
	KeyData []byte
	CrtData []byte
}

func (s ETCDv2Settings) AreComplete() bool {
	return s.DataDir != ""
}

func (s ETCDv3Settings) AreComplete() bool {
	return s.Endpoints != "" && s.TLSConfig != nil
}
