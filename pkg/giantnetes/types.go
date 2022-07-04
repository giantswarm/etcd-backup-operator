package giantnetes

import (
	"crypto/tls"
)

type ETCDv2Settings struct {
	DataDir string
}

type ETCDv3Settings struct {
	Endpoints string
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
