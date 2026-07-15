package zabbix

import (
	"context"
	"errors"
	"testing"
)

func skipIfNotFound(t *testing.T, err error) {
	t.Helper()
	if errors.Is(err, ErrNotFound) {
		t.Skip("No objects found")
	}
}

func TestItems(t *testing.T) {
	params := ItemGetParams{GetParameters: GetParameters{ResultLimit: 10}}
	items, err := GetTestSession(t).GetItems(context.Background(), params)
	skipIfNotFound(t, err)
	if err != nil {
		t.Fatalf("Error getting items: %v", err)
	}
	for i, item := range items {
		if item.ItemID == 0 {
			t.Fatalf("Item %d has no Item ID", i)
		}
	}
	t.Logf("Validated %d Items", len(items))
}

func TestTriggers(t *testing.T) {
	params := TriggerGetParams{GetParameters: GetParameters{ResultLimit: 10}}
	triggers, err := GetTestSession(t).GetTriggers(context.Background(), params)
	skipIfNotFound(t, err)
	if err != nil {
		t.Fatalf("Error getting triggers: %v", err)
	}
	for i, trigger := range triggers {
		if trigger.TriggerID == "" {
			t.Fatalf("Trigger %d has no Trigger ID", i)
		}
	}
	t.Logf("Validated %d Triggers", len(triggers))
}

func TestHistories(t *testing.T) {
	params := HistoryGetParams{GetParameters: GetParameters{ResultLimit: 10}}
	histories, err := GetTestSession(t).GetHistories(context.Background(), params)
	skipIfNotFound(t, err)
	if err != nil {
		t.Fatalf("Error getting histories: %v", err)
	}
	for i, history := range histories {
		if history.ItemID == 0 {
			t.Fatalf("History %d has no Item ID", i)
		}
	}
	t.Logf("Validated %d History values", len(histories))
}

func TestHostInterfaces(t *testing.T) {
	params := HostInterfaceGetParams{GetParameters: GetParameters{ResultLimit: 10}}
	interfaces, err := GetTestSession(t).GetHostInterfaces(context.Background(), params)
	skipIfNotFound(t, err)
	if err != nil {
		t.Fatalf("Error getting host interfaces: %v", err)
	}
	for i, hostInterface := range interfaces {
		if hostInterface.InterfaceID == "" {
			t.Fatalf("Host interface %d has no Interface ID", i)
		}
	}
	t.Logf("Validated %d Host Interfaces", len(interfaces))
}
