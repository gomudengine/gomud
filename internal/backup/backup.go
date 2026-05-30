package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	stopChan    chan struct{}
	reschedChan chan struct{}
	stopOnce    sync.Once
)

func BackupFilename() string {
	return fmt.Sprintf("world-backup-%s.tar.gz", time.Now().UTC().Format("2006-01-02-150405"))
}

func broadcast(msg string) {
	connections.Broadcast([]byte(msg))
}

// saveAllData persists users, rooms, and plugin state to disk.
// The caller must hold the MUD lock.
func saveAllData() {
	users.SaveAllUsers()
	rooms.SaveAllRooms()
	plugins.Save()
}

// CreateWorldBackup creates a tar.gz archive of the world data directory.
// The caller must hold the MUD lock during this call.
func CreateWorldBackup() ([]byte, string, error) {
	dataFiles := configs.GetFilePathsConfig().DataFiles.String()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	baseDir := filepath.Clean(dataFiles)

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})

	if err != nil {
		tw.Close()
		gz.Close()
		return nil, "", fmt.Errorf("creating tar archive: %w", err)
	}

	if err := tw.Close(); err != nil {
		gz.Close()
		return nil, "", fmt.Errorf("closing tar writer: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, "", fmt.Errorf("closing gzip writer: %w", err)
	}

	filename := BackupFilename()
	return buf.Bytes(), filename, nil
}

// RunBackup performs a full backup: broadcasts a warning to players, saves all
// data, locks the MUD, creates the archive, unlocks, and broadcasts completion.
// Returns the archive bytes, filename, and any error.
func RunBackup() ([]byte, string, error) {
	broadcast("\r\n*** World backup starting — the server will pause briefly. ***\r\n")

	util.LockMud()

	broadcast("Saving world data...\r\n")
	saveAllData()

	broadcast("Creating backup archive...\r\n")
	data, filename, err := CreateWorldBackup()

	util.UnlockMud()

	if err != nil {
		broadcast("*** Backup failed. ***\r\n")
		return nil, "", err
	}

	broadcast("*** World backup complete. ***\r\n")
	return data, filename, nil
}

func uploadToS3(data []byte, filename string, cfg configs.BackupS3) error {
	accessKey := string(cfg.AccessKey)
	secretKey := string(cfg.SecretKey)

	if accessKey == "" || secretKey == "" {
		return fmt.Errorf("S3 access key and secret key must be configured")
	}

	client := s3.New(s3.Options{
		Region:      string(cfg.Region),
		Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	})

	key := filename
	if prefix := strings.TrimRight(string(cfg.Prefix), "/"); prefix != "" {
		key = prefix + "/" + filename
	}

	_, err := client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(string(cfg.Bucket)),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/gzip"),
	})
	if err != nil {
		return fmt.Errorf("S3 upload: %w", err)
	}

	return nil
}

func runScheduledBackup() {
	mudlog.Info("Backup", "action", "scheduled backup starting")

	data, filename, err := RunBackup()
	if err != nil {
		mudlog.Error("Backup", "action", "scheduled backup failed", "error", err)
		return
	}

	mudlog.Info("Backup", "action", "world archive created", "filename", filename, "size", len(data))

	cfg := configs.GetBackupConfig()
	if bool(cfg.S3.Enabled) {
		if err := uploadToS3(data, filename, cfg.S3); err != nil {
			mudlog.Error("Backup", "action", "S3 upload failed", "error", err)
			return
		}
		mudlog.Info("Backup", "action", "S3 upload complete", "filename", filename)
	}
}

func nextScheduleTime(schedule string) time.Duration {
	now := time.Now()

	switch schedule {
	case configs.BackupScheduleNightly:
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
		return time.Until(next)
	case configs.BackupScheduleWeekly:
		daysUntilSunday := (7 - int(now.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		next := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSunday, 3, 0, 0, 0, now.Location())
		return time.Until(next)
	case configs.BackupScheduleMonthly:
		next := time.Date(now.Year(), now.Month()+1, 1, 3, 0, 0, 0, now.Location())
		return time.Until(next)
	default:
		return 0
	}
}

// Reschedule signals the backup scheduler to re-read the config and
// recalculate the next backup time immediately.
func Reschedule() {
	select {
	case reschedChan <- struct{}{}:
	default:
	}
}

// StartScheduler starts the background goroutine that runs scheduled backups.
// It re-reads the config each cycle so schedule/destination changes take effect
// without a restart. Call Reschedule() after changing backup config to
// apply immediately.
func StartScheduler() {
	stopChan = make(chan struct{})
	reschedChan = make(chan struct{}, 1)

	configs.OnChanged(func(key string) {
		if strings.HasPrefix(key, "Backup.") {
			Reschedule()
		}
	})

	go func() {
		for {
			cfg := configs.GetBackupConfig()
			schedule := string(cfg.Schedule)

			wait := nextScheduleTime(schedule)
			if wait <= 0 {
				// Backups disabled; wait for a reschedule signal or check hourly.
				select {
				case <-time.After(time.Hour):
					continue
				case <-reschedChan:
					continue
				case <-stopChan:
					return
				}
			}

			mudlog.Info("Backup", "schedule", schedule, "next", time.Now().Add(wait).Format(time.RFC3339))

			select {
			case <-time.After(wait):
				runScheduledBackup()
			case <-reschedChan:
				continue
			case <-stopChan:
				return
			}
		}
	}()
}

// StopScheduler signals the backup scheduler goroutine to exit.
func StopScheduler() {
	stopOnce.Do(func() {
		if stopChan != nil {
			close(stopChan)
		}
	})
}
