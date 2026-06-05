package users

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestUserRecordIsAISerializes(t *testing.T) {
	u := UserRecord{Username: "tester", IsAI: true}

	out, err := yaml.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), "isai: true") {
		t.Errorf("expected 'isai: true' in YAML, got:\n%s", out)
	}

	var back UserRecord
	if err := yaml.Unmarshal(out, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !back.IsAI {
		t.Errorf("IsAI should round-trip true, got false")
	}
}

func TestUserRecordIsAIOmittedWhenFalse(t *testing.T) {
	u := UserRecord{Username: "human"}
	out, err := yaml.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "isai:") {
		t.Errorf("isai should be omitted when false, got:\n%s", out)
	}
}
