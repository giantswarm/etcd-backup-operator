package giantnetes

type TLSClientConfig struct {
	CAData  []byte
	KeyData []byte
	CrtData []byte
	CAFile  string
	CrtFile string
	KeyFile string
}

type ETCDv2Settings struct {
	DataDir string
}

type ETCDv3Settings struct {
	Endpoints string
	CaCert    string
	Key       string
	Cert      string
}

type ETCDInstance struct {
	Name   string
	ETCDv2 ETCDv2Settings
	ETCDv3 ETCDv3Settings
}

func (s ETCDv2Settings) AreComplete() bool {
	return s.DataDir != ""
}

func (s ETCDv3Settings) AreComplete() bool {
	return s.Endpoints != "" && s.Cert != "" && s.CaCert != "" && s.Key != ""
}
