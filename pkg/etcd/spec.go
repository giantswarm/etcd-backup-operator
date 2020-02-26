package etcd

type Backupper interface {
	Create() (string, error)
	Cleanup()
	Encrypt() (string, error)
	Version() string
}
