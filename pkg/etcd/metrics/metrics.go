package metrics

type InstanceBackupMetrics struct {
	InstanceName   string
	Attempts       int32
	AttemptTS      int64
	BackupSize     int64
	CreationTime   int64
	EncryptionTime int64
	Failures       int32
	Successes      int32
	SuccessTS      int64
	UploadTime     int64
}
