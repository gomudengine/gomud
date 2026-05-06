package users

import "testing"

func TestUserRecordCloneOwnsUserMutableState(t *testing.T) {
	original := NewUserRecord(1, 100)
	original.Macros = map[string]string{"a": "look"}
	original.Aliases = map[string]string{"l": "look"}
	original.ConfigOptions = map[string]any{"theme": "dark"}
	original.TipsComplete = map[string]bool{"movement": true}
	original.EventLog = UserLog{{Category: "test", What: "original"}}

	cloned := original.Clone()
	cloned.Macros["a"] = "inventory"
	cloned.Aliases["l"] = "listen"
	cloned.ConfigOptions["theme"] = "light"
	cloned.TipsComplete["movement"] = false
	cloned.EventLog[0].What = "clone"

	if original.Macros["a"] != "look" {
		t.Fatalf("original macro = %q, want look", original.Macros["a"])
	}
	if original.Aliases["l"] != "look" {
		t.Fatalf("original alias = %q, want look", original.Aliases["l"])
	}
	if original.ConfigOptions["theme"] != "dark" {
		t.Fatalf("original theme = %q, want dark", original.ConfigOptions["theme"])
	}
	if !original.TipsComplete["movement"] {
		t.Fatal("original tip completion changed after mutating clone")
	}
	if original.EventLog[0].What != "original" {
		t.Fatalf("original event log entry = %q, want original", original.EventLog[0].What)
	}
}
