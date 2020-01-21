package storage

type Interface interface {
	Upload(string) (int64, error)
}
