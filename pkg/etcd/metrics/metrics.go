package metrics

type BackupAttemptResult struct {
	Successful                bool
	BackupSizeMeasurement     int64
	CreationTimeMeasurement   int64
	EncryptionTimeMeasurement int64
	UploadTimeMeasurement     int64
	Filename                  string
}

func NewSuccessfulBackupAttemptResult(backupSize int64, creationTime int64, encryptionTime int64, uploadTime int64, filename string) *BackupAttemptResult {
	return &BackupAttemptResult{
		Successful:                true,
		BackupSizeMeasurement:     backupSize,
		CreationTimeMeasurement:   creationTime,
		EncryptionTimeMeasurement: encryptionTime,
		UploadTimeMeasurement:     uploadTime,
		Filename:                  filename,
	}
}

func NewSkippedBackupAttemptResult(creationTime int64) *BackupAttemptResult {
	return &BackupAttemptResult{
		Successful:                false,
		BackupSizeMeasurement:     -1,
		CreationTimeMeasurement:   creationTime,
		EncryptionTimeMeasurement: -1,
		UploadTimeMeasurement:     -1,
		Filename:                  "",
	}
}

func NewFailedBackupAttemptResult() *BackupAttemptResult {
	return &BackupAttemptResult{
		Successful:                false,
		BackupSizeMeasurement:     -1,
		CreationTimeMeasurement:   -1,
		EncryptionTimeMeasurement: -1,
		UploadTimeMeasurement:     -1,
		Filename:                  "",
	}
}
