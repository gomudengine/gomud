package configs

const (
	BackupScheduleDisabled = ``
	BackupScheduleNever    = `never`
	BackupScheduleNightly  = `nightly`
	BackupScheduleWeekly   = `weekly`
	BackupScheduleMonthly  = `monthly`
)

type Backup struct {
	Schedule ConfigString `yaml:"Schedule"` // nightly, weekly, monthly, or empty to disable
	S3       BackupS3     `yaml:"S3"`
}

type BackupS3 struct {
	Enabled   ConfigBool   `yaml:"Enabled"`
	Bucket    ConfigString `yaml:"Bucket"`
	Region    ConfigString `yaml:"Region"`
	Prefix    ConfigString `yaml:"Prefix"` // key prefix / folder within the bucket
	AccessKey ConfigSecret `yaml:"AccessKey" env:"BACKUP_S3_ACCESS_KEY"`
	SecretKey ConfigSecret `yaml:"SecretKey" env:"BACKUP_S3_SECRET_KEY"`
}

func (b *Backup) Validate() {
	switch string(b.Schedule) {
	case BackupScheduleNightly, BackupScheduleWeekly, BackupScheduleMonthly, BackupScheduleNever, BackupScheduleDisabled:
	default:
		b.Schedule = ``
	}

	if b.S3.Region == `` {
		b.S3.Region = `us-east-1`
	}
}

func GetBackupConfig() Backup {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Backup
}
