package users

import (
	"bytes"
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

	type candidate struct {
		userId   int
		username string
	}

	var exact *candidate
	var close []candidate

	GetUserIndex().ForEachRecord(func(rec IndexUserRecord) bool {
		raw := string(bytes.TrimRight(rec.Username[:], "\x00"))
		lower := strings.ToLower(raw)

		if lower == needle {
			c := candidate{userId: int(rec.UserID), username: raw}
			exact = &c
			return false
		}
		if strings.HasPrefix(lower, needle) && len(close) < 100 {
			close = append(close, candidate{userId: int(rec.UserID), username: raw})
		}
		return true
	})

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

		if u := GetByUserId(c.userId); u != nil {
			result.Role = u.Role
			result.Email = u.EmailAddress
		} else {
			if loaded, err := LoadUser(c.username, true); err == nil {
				result.Role = loaded.Role
				result.Email = loaded.EmailAddress
			}
		}

		results = append(results, result)
	}

	return results
}
