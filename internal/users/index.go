package users

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	ErrIndexMissing        = errors.New(`index does not exist`)
	ErrUserFilesOldFormat  = errors.New(`user files are in old format of username.yaml`)
	ErrIndexVersionInvalid = errors.New(`version out of date.`)
	ErrSearchNameTooLong   = errors.New(`search name provided is too long`)
	ErrNotFound            = errors.New("user not found")
)

const (
	IndexVersion           = 1
	IndexLineTerminatorV1  = byte(10) // "\n"
	IndexRecordSizeV1      = 89
	FixedHeaderTotalLength = 100 // 99 bytes header content + 1 byte newline
)

// IndexMetaData holds header info that helps in reading the file.
type IndexMetaData struct {
	MetaDataSize uint64 // size of the metadata header (in bytes)
	IndexVersion uint64
	RecordCount  uint64
	RecordSize   uint64
}

// IndexUserRecord represents one fixed-width record.
type IndexUserRecord struct {
	UserID   int64
	Username [80]byte
}

// UserIndex is the central struct that holds the index filename and methods
// to work with the index.
type UserIndex struct {
	mu            sync.RWMutex
	metaData      IndexMetaData
	highestUserId int
	Filename      string

	records    []IndexUserRecord
	byUsername map[string]int64
	byUserId   map[int64]string
}

var userIndex *UserIndex

// InitUserIndex creates and initializes the singleton UserIndex. Called once at startup.
func InitUserIndex() *UserIndex {
	filename := util.FilePath(string(configs.GetFilePathsConfig().DataFiles), `/`, `users`, `/`, `users.idx`)
	userIndex = &UserIndex{Filename: filename}
	if userIndex.Exists() {
		userIndex.metaData = userIndex.getMetaDataFromFile()
		userIndex.loadRecords()
	}
	return userIndex
}

// GetUserIndex returns the singleton UserIndex, initializing it if needed.
func GetUserIndex() *UserIndex {
	if userIndex == nil {
		return InitUserIndex()
	}
	return userIndex
}

func (idx *UserIndex) Exists() bool {
	_, err := os.Stat(idx.Filename)
	return err == nil
}

func (idx *UserIndex) Delete() {
	if idx.Exists() {
		os.Remove(idx.Filename)
	}
}

// loadRecords bulk-reads all records from disk into memory and builds lookup maps.
func (idx *UserIndex) loadRecords() {
	idx.byUsername = make(map[string]int64, idx.metaData.RecordCount)
	idx.byUserId = make(map[int64]string, idx.metaData.RecordCount)
	idx.highestUserId = 0

	if idx.metaData.RecordCount == 0 {
		idx.records = nil
		return
	}

	f, err := os.Open(idx.Filename)
	if err != nil {
		return
	}
	defer f.Close()

	dataSize := idx.metaData.RecordCount * idx.metaData.RecordSize
	buf := make([]byte, dataSize)
	if _, err := f.Seek(int64(idx.metaData.MetaDataSize), io.SeekStart); err != nil {
		return
	}
	if _, err := io.ReadFull(f, buf); err != nil {
		return
	}

	idx.records = make([]IndexUserRecord, idx.metaData.RecordCount)

	for i := uint64(0); i < idx.metaData.RecordCount; i++ {
		offset := i * idx.metaData.RecordSize
		rec := &idx.records[i]
		copy(rec.Username[:], buf[offset:offset+80])
		rec.UserID = int64(binary.LittleEndian.Uint64(buf[offset+80 : offset+88]))

		username := string(bytes.TrimRight(rec.Username[:], "\x00"))
		idx.byUsername[username] = rec.UserID
		idx.byUserId[rec.UserID] = username

		if int(rec.UserID) > idx.highestUserId {
			idx.highestUserId = int(rec.UserID)
		}
	}
}

// Create initializes a new empty index file with a header.
func (idx *UserIndex) Create() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.Delete()

	f, err := os.Create(idx.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	idx.metaData = IndexMetaData{
		MetaDataSize: FixedHeaderTotalLength,
		IndexVersion: IndexVersion,
		RecordCount:  0,
		RecordSize:   IndexRecordSizeV1,
	}
	idx.highestUserId = 0
	idx.records = nil
	idx.byUsername = make(map[string]int64)
	idx.byUserId = make(map[int64]string)

	headerBytes, err := idx.metaData.Format()
	if err != nil {
		return err
	}
	if _, err := f.Write(headerBytes); err != nil {
		return err
	}

	return nil
}

// Rebuild recreates the index from all offline user records.
// It calls Create() internally so it is self-contained.
func (idx *UserIndex) Rebuild() error {
	if err := idx.Create(); err != nil {
		return fmt.Errorf("index create failed: %w", err)
	}

	var firstErr error
	SearchOfflineUsers(func(u *UserRecord) bool {
		if err := idx.AddUser(u.UserId, u.Username); err != nil {
			mudlog.Error("UserIndex.Rebuild", "error", err.Error(), "userId", u.UserId, "username", u.Username)
			if firstErr == nil {
				firstErr = err
			}
		}
		return true
	})

	return firstErr
}

func (idx *UserIndex) GetMetaData() IndexMetaData {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.metaData
}

func (idx *UserIndex) GetHighestUserId() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.highestUserId
}

// ForEachRecord iterates over all index records, calling fn for each.
// Returning false from fn stops iteration.
func (idx *UserIndex) ForEachRecord(fn func(rec IndexUserRecord) bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	for _, rec := range idx.records {
		if !fn(rec) {
			return
		}
	}
}

// FindByUsername searches the index for a username and returns its userId.
func (idx *UserIndex) FindByUsername(username string) (int, bool) {
	if len(username) > 80 {
		return 0, false
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	userId, ok := idx.byUsername[strings.ToLower(username)]
	if !ok {
		return 0, false
	}
	return int(userId), true
}

// FindByUserId searches for a user record matching the provided userId.
func (idx *UserIndex) FindByUserId(userId int) (string, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	username, ok := idx.byUserId[int64(userId)]
	if !ok {
		return "", false
	}
	return username, true
}

func (idx *UserIndex) getMetaDataFromFile() IndexMetaData {
	f, err := os.Open(idx.Filename)
	if err != nil {
		return IndexMetaData{}
	}
	defer f.Close()

	header := make([]byte, FixedHeaderTotalLength)
	if _, err := io.ReadFull(f, header); err != nil {
		return IndexMetaData{}
	}

	var meta IndexMetaData
	meta.MetaDataSize = uint64(len(header))
	headerContent := strings.TrimSpace(string(header[:FixedHeaderTotalLength-1]))
	n, _ := fmt.Sscanf(headerContent, "VERSION=%d,RECORDCOUNT=%d,RECORDSIZE=%d", &meta.IndexVersion, &meta.RecordCount, &meta.RecordSize)
	if n != 3 {
		return IndexMetaData{}
	}

	return meta
}

// AddUser appends a new record to the index file and updates the header.
func (idx *UserIndex) AddUser(userId int, username string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	username = strings.ToLower(username)

	newRecord := IndexUserRecord{
		UserID: int64(userId),
	}
	copy(newRecord.Username[:], username)

	f, err := os.OpenFile(idx.Filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("error seeking to file end: %w", err)
	}

	var recBuf [IndexRecordSizeV1]byte
	copy(recBuf[:80], newRecord.Username[:])
	binary.LittleEndian.PutUint64(recBuf[80:88], uint64(newRecord.UserID))
	recBuf[88] = IndexLineTerminatorV1
	if _, err := f.Write(recBuf[:]); err != nil {
		return fmt.Errorf("error writing record: %w", err)
	}

	if userId > idx.highestUserId {
		idx.highestUserId = userId
	}

	idx.metaData.RecordCount++

	newHeaderBytes, err := idx.metaData.Format()
	if err != nil {
		return fmt.Errorf("error formatting header: %w", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("error seeking to beginning: %w", err)
	}
	if _, err := f.Write(newHeaderBytes); err != nil {
		return fmt.Errorf("error writing updated header: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("error syncing file: %w", err)
	}

	idx.records = append(idx.records, newRecord)
	idx.byUsername[username] = int64(userId)
	idx.byUserId[int64(userId)] = username

	return nil
}

// RemoveByUsername removes the first record matching the username and rewrites the index.
func (idx *UserIndex) RemoveByUsername(username string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	username = strings.ToLower(username)

	if _, ok := idx.byUsername[username]; !ok {
		return ErrNotFound
	}

	newRecords := make([]IndexUserRecord, 0, len(idx.records)-1)
	removed := false

	for _, rec := range idx.records {
		if !removed {
			recUser := string(bytes.TrimRight(rec.Username[:], "\x00"))
			if recUser == username {
				removed = true
				delete(idx.byUsername, username)
				delete(idx.byUserId, rec.UserID)
				continue
			}
		}
		newRecords = append(newRecords, rec)
	}

	idx.records = newRecords
	idx.metaData.RecordCount = uint64(len(newRecords))

	idx.highestUserId = 0
	for _, rec := range idx.records {
		if int(rec.UserID) > idx.highestUserId {
			idx.highestUserId = int(rec.UserID)
		}
	}

	return idx.writeCompleteIndex(newRecords)
}

// Format formats the metadata header as a fixed-width string.
// The header (without newline) is exactly 99 bytes.
func (m IndexMetaData) Format() ([]byte, error) {
	headerContent := fmt.Sprintf("VERSION=%d,RECORDCOUNT=%d,RECORDSIZE=%d", m.IndexVersion, m.RecordCount, m.RecordSize)
	if len(headerContent) > FixedHeaderTotalLength-1 {
		return nil, fmt.Errorf("header content too long: %d bytes", len(headerContent))
	}
	padded := headerContent + strings.Repeat(" ", FixedHeaderTotalLength-1-len(headerContent))
	return []byte(padded + string(IndexLineTerminatorV1)), nil
}

// writeCompleteIndex writes metadata and all records atomically via temp file + rename.
func (idx *UserIndex) writeCompleteIndex(records []IndexUserRecord) error {
	tmpFile := idx.Filename + ".tmp"
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}

	writeErr := func() error {
		headerBytes, err := idx.metaData.Format()
		if err != nil {
			return err
		}

		buf := make([]byte, 0, len(headerBytes)+len(records)*IndexRecordSizeV1)
		buf = append(buf, headerBytes...)

		var recBuf [IndexRecordSizeV1]byte
		for _, rec := range records {
			copy(recBuf[:80], rec.Username[:])
			binary.LittleEndian.PutUint64(recBuf[80:88], uint64(rec.UserID))
			recBuf[88] = IndexLineTerminatorV1
			buf = append(buf, recBuf[:]...)
		}

		if _, err := f.Write(buf); err != nil {
			return err
		}

		return f.Sync()
	}()

	if closeErr := f.Close(); writeErr == nil {
		writeErr = closeErr
	}

	if writeErr != nil {
		os.Remove(tmpFile)
		return writeErr
	}

	return os.Rename(tmpFile, idx.Filename)
}
