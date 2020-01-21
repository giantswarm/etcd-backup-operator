package resource

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
