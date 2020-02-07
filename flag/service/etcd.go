package service

type ETCDv2Settings struct {
	DataDir string
}

type ETCDv3Settings struct {
	Endpoints string
	CaCert    string
	Key       string
	Cert      string
}
