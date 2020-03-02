package metrics

type instanceBackupMetrics struct {
	AttemptsCount   int32
	ETCDVersion     string
	InstanceName    string
	LatestAttemptTS int64
	BackupFileSize  int64
	CreationTime    int64
	EncryptionTime  int64
	FailuresCount   int32
	SuccessesCount  int32
	LatestSuccessTS int64
	UploadTime      int64
}
