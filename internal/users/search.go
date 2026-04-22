package users

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"strings"
)

// UserSearchResult holds the subset of user data returned by search endpoints.
type UserSearchResult struct {
	UserId   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Email    string `json:"email"`
}

// SearchUsers searches for users whose username matches or is prefixed by
// searchName. If an exact match is found it is the sole result. Otherwise all
// prefix matches are returned. The search is case-insensitive.
//
// Online users are checked first via the in-memory index. Offline users are
// found by scanning the binary index file directly, which avoids loading full
// user records for every candidate.
func SearchUsers(searchName string) []UserSearchResult {
	if searchName == "" {
		return nil
	}

	needle := strings.ToLower(searchName)

	// Collect candidates: username -> userId from the index.
	type candidate struct {
		userId   int
		username string
	}

	var exact *candidate
	var close []candidate

	idx := NewUserIndex()
	if idx.Exists() {
		f, err := os.Open(idx.Filename)
		if err == nil {
			defer f.Close()

			meta := idx.metaData
			for i := uint64(0); i < meta.RecordCount; i++ {
				offset := int64(meta.MetaDataSize) + int64(i*meta.RecordSize)
				if _, err := f.Seek(offset, io.SeekStart); err != nil {
					break
				}

				var recUsername [80]byte
				if n, err := io.ReadFull(f, recUsername[:]); err != nil || n != 80 {
					break
				}

				var userId int64
				if err := binary.Read(f, binary.LittleEndian, &userId); err != nil {
					break
				}

				// consume terminator byte
				term := make([]byte, 1)
				if _, err := f.Read(term); err != nil {
					break
				}

				raw := string(bytes.TrimRight(recUsername[:], "\x00"))
				lower := strings.ToLower(raw)

				if lower == needle {
					c := candidate{userId: int(userId), username: raw}
					exact = &c
					break
				}
				if strings.HasPrefix(lower, needle) && len(close) < 100 {
					close = append(close, candidate{userId: int(userId), username: raw})
				}
			}
		}
	}

	var matches []candidate
	if exact != nil {
		matches = []candidate{*exact}
	} else {
		matches = close
	}

	if len(matches) == 0 {
		return nil
	}

	results := make([]UserSearchResult, 0, len(matches))
	for _, c := range matches {
		result := UserSearchResult{
			UserId:   c.userId,
			Username: c.username,
		}

		// Populate role and email. Check online users first (free), then load
		// from disk only if needed.
		if u := GetByUserId(c.userId); u != nil {
			result.Role = u.Role
			result.Email = u.EmailAddress
		} else {
			// Load just enough from disk to fill role/email without full validation.
			if loaded, err := LoadUser(c.username, true); err == nil {
				result.Role = loaded.Role
				result.Email = loaded.EmailAddress
			}
		}

		results = append(results, result)
	}

	return results
}
