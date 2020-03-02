package metrics

type InstanceBackupMetrics struct {
	InstanceName    string
	AttemptsCount   int32
	LatestAttemptTS int64
	BackupFileSize  int64
	CreationTime    int64
	EncryptionTime  int64
	FailuresCount   int32
	SuccessesCount  int32
	LatestSuccessTS int64
	UploadTime      int64
}
