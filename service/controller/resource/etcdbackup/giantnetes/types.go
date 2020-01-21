package giantnetes

type TLSClientConfig struct {
	CAData  []byte
	KeyData []byte
	CrtData []byte
	CAFile  string
	CrtFile string
	KeyFile string
}
