package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func TestAPIV1PatchUserRejectsSensitiveFields(t *testing.T) {
	setupAuthTestUsers(t, "correct-password", map[string]string{
		"player": users.RoleUser,
	})

	tests := []struct {
		name string
		body string
	}{
		{name: "user id", body: `{"UserId":2}`},
		{name: "role", body: `{"Role":"admin"}`},
		{name: "username", body: `{"Username":"renamed"}`},
		{name: "password", body: `{"Password":"new-password"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/admin/api/v1/users/1", bytes.NewBufferString(tt.body))
			req.SetPathValue("userid", "1")
			rec := httptest.NewRecorder()

			apiV1PatchUser(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}

			u, err := users.LoadUserById(1)
			if err != nil {
				t.Fatalf("LoadUserById: %v", err)
			}
			if u.Role != users.RoleUser {
				t.Fatalf("Role = %q, want %q", u.Role, users.RoleUser)
			}
			if u.UserId != 1 {
				t.Fatalf("UserId = %d, want 1", u.UserId)
			}
			if u.Username != "player" {
				t.Fatalf("Username = %q, want player", u.Username)
			}
			if !u.PasswordMatches("correct-password") {
				t.Fatal("stored password no longer matches original password")
			}
		})
	}
}

func TestAPIV1PatchUserAllowsProfileAndCharacterFields(t *testing.T) {
	setupAuthTestUsers(t, "correct-password", map[string]string{
		"player": users.RoleUser,
	})

	req := httptest.NewRequest(http.MethodPatch, "/admin/api/v1/users/1", bytes.NewBufferString(`{
		"EmailAddress":"player@example.com",
		"Muted":true,
		"ScreenReader":true,
		"character":{"Name":"Hero","Gold":-50,"Bank":-25}
	}`))
	req.SetPathValue("userid", "1")
	rec := httptest.NewRecorder()

	apiV1PatchUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	u, err := users.LoadUserById(1)
	if err != nil {
		t.Fatalf("LoadUserById: %v", err)
	}
	if u.EmailAddress != "player@example.com" {
		t.Fatalf("EmailAddress = %q, want player@example.com", u.EmailAddress)
	}
	if !u.Muted || !u.ScreenReader {
		t.Fatalf("Muted/ScreenReader = %t/%t, want true/true", u.Muted, u.ScreenReader)
	}
	if u.Character.Name != "Hero" {
		t.Fatalf("Character.Name = %q, want Hero", u.Character.Name)
	}
	if u.Character.Gold != 0 || u.Character.Bank != 0 {
		t.Fatalf("Gold/Bank = %d/%d, want 0/0", u.Character.Gold, u.Character.Bank)
	}
}

func TestUserRecordCloneDoesNotShareMutableCharacterFields(t *testing.T) {
	original := users.NewUserRecord(1, 0)
	original.Character = characters.New()
	original.Character.Skills["stealth"] = 3
	original.Character.Items = []items.Item{{
		ItemId:     1001,
		Adjectives: []string{"old"},
	}}
	original.Character.Equipment.Weapon = items.Item{
		ItemId:     2001,
		Adjectives: []string{"sharp"},
	}
	original.Character.Pet.Items = []items.Item{{
		ItemId:     3001,
		Adjectives: []string{"pet-old"},
	}}
	original.Character.SpellBook = map[string]int{"spark": 1}
	original.Character.Shop = characters.Shop{{ItemId: 4001, Quantity: 1}}

	updated := original.Clone()
	updated.Character.Skills["stealth"] = 9
	updated.Character.Items[0].ItemId = 1002
	updated.Character.Items[0].Adjectives[0] = "new"
	updated.Character.Equipment.Weapon.ItemId = 2002
	updated.Character.Equipment.Weapon.Adjectives[0] = "dull"
	updated.Character.Pet.Items[0].ItemId = 3002
	updated.Character.Pet.Items[0].Adjectives[0] = "pet-new"
	updated.Character.SpellBook["spark"] = 2
	updated.Character.Shop[0].Quantity = 5

	if original.Character.Skills["stealth"] != 3 {
		t.Fatalf("original skill level = %d, want 3", original.Character.Skills["stealth"])
	}
	if original.Character.Items[0].ItemId != 1001 || original.Character.Items[0].Adjectives[0] != "old" {
		t.Fatalf("original item = %+v, want item 1001 with old adjective", original.Character.Items[0])
	}
	if original.Character.Equipment.Weapon.ItemId != 2001 || original.Character.Equipment.Weapon.Adjectives[0] != "sharp" {
		t.Fatalf("original weapon = %+v, want item 2001 with sharp adjective", original.Character.Equipment.Weapon)
	}
	if original.Character.Pet.Items[0].ItemId != 3001 || original.Character.Pet.Items[0].Adjectives[0] != "pet-old" {
		t.Fatalf("original pet item = %+v, want item 3001 with pet-old adjective", original.Character.Pet.Items[0])
	}
	if original.Character.SpellBook["spark"] != 1 {
		t.Fatalf("original spell level = %d, want 1", original.Character.SpellBook["spark"])
	}
	if original.Character.Shop[0].Quantity != 1 {
		t.Fatalf("original shop quantity = %d, want 1", original.Character.Shop[0].Quantity)
	}
}

func TestAPIV1CreateUserRejectsRoleAssignment(t *testing.T) {
	setupAuthTestUsers(t, "correct-password", map[string]string{
		"admin": users.RoleAdmin,
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/users", bytes.NewBufferString(`{
		"Username":"newadmin",
		"Password":"correct-password",
		"role":"admin"
	}`))
	rec := httptest.NewRecorder()

	apiV1CreateUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if users.Exists("newadmin") {
		t.Fatal("newadmin should not have been created")
	}
}

func TestAPIV1ResetUserPasswordUpdatesStoredHash(t *testing.T) {
	setupAuthTestUsers(t, "correct-password", map[string]string{
		"player": users.RoleUser,
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/users/1/password", bytes.NewBufferString(`{
		"Password":"new-password"
	}`))
	req.SetPathValue("userid", "1")
	rec := httptest.NewRecorder()

	apiV1ResetUserPassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	u, err := users.LoadUserById(1)
	if err != nil {
		t.Fatalf("LoadUserById: %v", err)
	}
	if !u.PasswordMatches("new-password") {
		t.Fatal("stored password does not match new password")
	}
	if u.Password == "new-password" {
		t.Fatal("password was stored in plaintext")
	}
}

func TestAPIV1ResetUserPasswordRequiresPassword(t *testing.T) {
	setupAuthTestUsers(t, "correct-password", map[string]string{
		"player": users.RoleUser,
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/users/1/password", bytes.NewBufferString(`{}`))
	req.SetPathValue("userid", "1")
	rec := httptest.NewRecorder()

	apiV1ResetUserPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	u, err := users.LoadUserById(1)
	if err != nil {
		t.Fatalf("LoadUserById: %v", err)
	}
	if !u.PasswordMatches("correct-password") {
		t.Fatal("stored password no longer matches original password")
	}
}

func TestAPIV1CreateUserDefaultsToUserRole(t *testing.T) {
	setupAuthTestUsers(t, "correct-password", map[string]string{
		"admin": users.RoleAdmin,
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/users", bytes.NewBufferString(`{
		"Username":"newuser",
		"Password":"correct-password",
		"EmailAddress":"newuser@example.com"
	}`))
	rec := httptest.NewRecorder()

	apiV1CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var response APIResponse[*users.UserRecord]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal response: %v", err)
	}
	if response.Data.Role != users.RoleUser {
		t.Fatalf("Role = %q, want %q", response.Data.Role, users.RoleUser)
	}
}
