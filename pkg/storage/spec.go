package storage

type Uploader interface {
	Upload(string) (int64, error)
}
