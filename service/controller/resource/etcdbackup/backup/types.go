package backup

type BackupInterface interface {
	Create() (string, error)
	Cleanup()
	Encrypt() (string, error)
	Version() string
}
