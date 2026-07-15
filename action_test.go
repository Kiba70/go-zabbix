package zabbix

import (
	"context"
	"testing"
)

func TestParseActionStepDuration(t *testing.T) {
	tests := map[string]int{
		"30":  30,
		"30s": 30,
		"5m":  300,
		"1h":  3600,
		"2d":  172800,
		"1w":  604800,
	}

	for value, want := range tests {
		got, err := parseActionStepDuration(value)
		if err != nil {
			t.Errorf("parseActionStepDuration(%q) returned an error: %v", value, err)
			continue
		}
		if got != want {
			t.Errorf("parseActionStepDuration(%q) = %d, want %d", value, got, want)
		}
	}

	for _, value := range []string{"", "abc", "1y"} {
		if _, err := parseActionStepDuration(value); err == nil {
			t.Errorf("parseActionStepDuration(%q) did not return an error", value)
		}
	}
}

func TestActions(t *testing.T) {
	session := GetTestSession(t)

	params := ActionGetParams{}

	actions, err := session.GetActions(context.Background(), params)
	if err != nil {
		t.Fatalf("Error getting actions: %v", err)
	}

	if len(actions) == 0 {
		t.Fatal("No actions found")
	}

	for i, action := range actions {
		if action.ActionID == "" {
			t.Fatalf("Action %d has no Action ID", i)
		}

		if action.Name == "" {
			t.Fatalf("Action %d has no name", i)
		}

	}

	t.Logf("Validated %d Actions", len(actions))
}
