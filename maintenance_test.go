package zabbix

import (
	"context"
	"testing"
)

func TestMaintenance(t *testing.T) {
	session := GetTestSession(t)

	params := &MaintenanceGetParams{
		SelectHosts:       SelectExtendedOutput,
		SelectGroups:      SelectExtendedOutput,
		SelectTimeperiods: SelectExtendedOutput,
		SelectTags:        SelectExtendedOutput,
	}

	maintenances, err := session.GetMaintenance(context.Background(), params)
	if err != nil {
		t.Fatalf("Error getting maintenances: %v", err)
	}

	if len(maintenances) == 0 {
		t.Fatal("No maintenance found")
	}

	for i, maintenance := range maintenances {
		if maintenance.MaintenanceID == "" {
			t.Fatalf("Maintenance %d returned in response body has no ID", i)
		}
		if maintenance.Session == nil {
			t.Fatalf("Maintenance %d returned in response body has no session", i)
		}
	}

	t.Logf("Validated %d maintenances", len(maintenances))
}
