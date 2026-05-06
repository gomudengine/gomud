package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// resolveUserId resolves a path segment that is a numeric user ID.
// Returns 0 and writes an error response if the value is not a valid integer.
func resolveUserId(w http.ResponseWriter, idStr string) int {
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid user id: "+idStr)
		return 0
	}
	return id
}

// loadUserRecord loads a UserRecord by id, checking the online cache first.
// Writes a 404 error response and returns nil when the user does not exist.
func loadUserRecord(w http.ResponseWriter, userId int) *users.UserRecord {
	if u := users.GetByUserId(userId); u != nil {
		return u
	}
	u, err := users.LoadUserById(userId)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "user not found")
		return nil
	}
	return u
}

// GET /admin/api/v1/users/{userid}
func apiV1GetUser(w http.ResponseWriter, r *http.Request) {
	userId := resolveUserId(w, r.PathValue("userid"))
	if userId == 0 {
		return
	}

	u := loadUserRecord(w, userId)
	if u == nil {
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*users.UserRecord]{
		Success: true,
		Data:    u,
	})
}

// PATCH /admin/api/v1/users/{userid}
func apiV1PatchUser(w http.ResponseWriter, r *http.Request) {
	userId := resolveUserId(w, r.PathValue("userid"))
	if userId == 0 {
		return
	}

	u := loadUserRecord(w, userId)
	if u == nil {
		return
	}

	updated := u.Clone()

	if err := applyUserPatch(&updated, r.Body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	// Preserve the canonical ID; callers cannot change it via PATCH.
	updated.UserId = userId

	if updated.Character.Gold < 0 {
		updated.Character.Gold = 0
	}
	if updated.Character.Bank < 0 {
		updated.Character.Bank = 0
	}

	updated.Character.Validate()

	if err := users.SaveUser(updated); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	users.UpdateOnlineUser(updated)

	writeJSON(w, http.StatusOK, APIResponse[*users.UserRecord]{
		Success: true,
		Data:    &updated,
	})
}

// POST /admin/api/v1/users/{userid}/password
func apiV1ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	userId := resolveUserId(w, r.PathValue("userid"))
	if userId == 0 {
		return
	}

	u := loadUserRecord(w, userId)
	if u == nil {
		return
	}

	var body struct {
		Password string `json:"Password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if body.Password == "" {
		writeAPIError(w, http.StatusBadRequest, "Password is required")
		return
	}

	updated := u.Clone()
	if err := updated.SetPassword(body.Password); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := users.SaveUser(updated); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	users.UpdateOnlineUser(updated)

	writeJSON(w, http.StatusOK, APIResponse[*users.UserRecord]{
		Success: true,
		Data:    &updated,
	})
}

func applyUserPatch(u *users.UserRecord, body io.Reader) error {
	var fields map[string]json.RawMessage
	if err := json.NewDecoder(body).Decode(&fields); err != nil {
		return err
	}

	for key, raw := range fields {
		switch strings.ToLower(key) {
		case "emailaddress":
			if err := json.Unmarshal(raw, &u.EmailAddress); err != nil {
				return err
			}
		case "muted":
			if err := json.Unmarshal(raw, &u.Muted); err != nil {
				return err
			}
		case "screenreader":
			if err := json.Unmarshal(raw, &u.ScreenReader); err != nil {
				return err
			}
		case "character":
			if err := applyCharacterPatch(u.Character, raw); err != nil {
				return err
			}
		case "userid", "role", "username", "password", "joined", "macros", "aliases", "configoptions", "tipscomplete":
			return jsonFieldError(key, "field cannot be changed through this endpoint")
		default:
			return jsonFieldError(key, "unsupported user patch field")
		}
	}

	return nil
}

func applyCharacterPatch(c *characters.Character, body json.RawMessage) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		return err
	}

	for key, raw := range fields {
		switch strings.ToLower(key) {
		case "name":
			var name string
			if err := json.Unmarshal(raw, &name); err != nil {
				return err
			}
			c.Name = name
		case "description":
			if err := json.Unmarshal(raw, &c.Description); err != nil {
				return err
			}
		case "raceid":
			if err := json.Unmarshal(raw, &c.RaceId); err != nil {
				return err
			}
		case "experience":
			if err := json.Unmarshal(raw, &c.Experience); err != nil {
				return err
			}
		case "roomid":
			if err := json.Unmarshal(raw, &c.RoomId); err != nil {
				return err
			}
		case "alignment":
			if err := json.Unmarshal(raw, &c.Alignment); err != nil {
				return err
			}
		case "gold":
			if err := json.Unmarshal(raw, &c.Gold); err != nil {
				return err
			}
		case "bank":
			if err := json.Unmarshal(raw, &c.Bank); err != nil {
				return err
			}
		case "trainingpoints":
			if err := json.Unmarshal(raw, &c.TrainingPoints); err != nil {
				return err
			}
		case "statpoints":
			if err := json.Unmarshal(raw, &c.StatPoints); err != nil {
				return err
			}
		case "extralives":
			if err := json.Unmarshal(raw, &c.ExtraLives); err != nil {
				return err
			}
		case "stats":
			if err := json.Unmarshal(raw, &c.Stats); err != nil {
				return err
			}
		case "skills":
			if err := json.Unmarshal(raw, &c.Skills); err != nil {
				return err
			}
		case "equipment":
			if err := json.Unmarshal(raw, &c.Equipment); err != nil {
				return err
			}
		case "items":
			if err := json.Unmarshal(raw, &c.Items); err != nil {
				return err
			}
		case "shop":
			if err := json.Unmarshal(raw, &c.Shop); err != nil {
				return err
			}
		case "spellbook":
			if err := json.Unmarshal(raw, &c.SpellBook); err != nil {
				return err
			}
		case "pet":
			if err := json.Unmarshal(raw, &c.Pet); err != nil {
				return err
			}
		default:
			return jsonFieldError("Character."+key, "unsupported character patch field")
		}
	}

	return nil
}

func jsonFieldError(field string, reason string) error {
	return fmt.Errorf("%s: %s", field, reason)
}

// POST /admin/api/v1/users
func apiV1CreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"Username"`
		Password string `json:"Password"`
		Email    string `json:"EmailAddress"`
		Role     string `json:"Role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	if body.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "Username is required")
		return
	}

	if users.Exists(body.Username) {
		writeAPIError(w, http.StatusConflict, "username already exists")
		return
	}
	if body.Role != "" {
		writeAPIError(w, http.StatusBadRequest, "Role cannot be set through this endpoint")
		return
	}

	u := users.NewUserRecord(0, 0)
	if err := u.SetUsername(body.Username); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.Password != "" {
		if err := u.SetPassword(body.Password); err != nil {
			writeAPIError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.Email != "" {
		u.EmailAddress = body.Email
	}

	if err := users.CreateUser(u); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse[*users.UserRecord]{
		Success: true,
		Data:    u,
	})
}

// GET /admin/api/v1/users/search
func apiV1SearchUsers(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "name query parameter is required")
		return
	}

	results := users.SearchUsers(name)
	if results == nil {
		results = []users.UserSearchResult{}
	}

	writeJSON(w, http.StatusOK, APIResponse[[]users.UserSearchResult]{
		Success: true,
		Data:    results,
	})
}
