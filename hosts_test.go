package zabbix

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHostInventoryUnmarshal(t *testing.T) {
	for _, data := range []string{
		`{"inventory":[]}`,
		`{"inventory":{"date_hw_install":"2026-01-02"}}`,
		`{"interfaces":[{"details":[]}]}`,
	} {
		var host Host
		if err := json.Unmarshal([]byte(data), &host); err != nil {
			t.Errorf("cannot unmarshal %s: %v", data, err)
		}
	}
}

func TestHosts(t *testing.T) {
	session := GetTestSession(t)

	params := HostGetParams{
		GetParameters:         GetParameters{ResultLimit: 10},
		IncludeTemplates:      true,
		SelectGroups:          SelectExtendedOutput,
		SelectApplications:    SelectExtendedOutput,
		SelectDiscoveries:     SelectExtendedOutput,
		SelectDiscoveryRule:   SelectExtendedOutput,
		SelectGraphs:          SelectExtendedOutput,
		SelectHostDiscovery:   SelectExtendedOutput,
		SelectWebScenarios:    SelectExtendedOutput,
		SelectInterfaces:      SelectExtendedOutput,
		SelectInventory:       SelectExtendedOutput,
		SelectItems:           SelectExtendedOutput,
		SelectMacros:          SelectExtendedOutput,
		SelectParentTemplates: SelectExtendedOutput,
		SelectScreens:         SelectExtendedOutput,
		SelectTriggers:        SelectExtendedOutput,
	}

	hosts, err := session.GetHosts(context.Background(), params)
	if err != nil {
		t.Fatalf("Error getting Hosts: %v", err)
	}

	if len(hosts) == 0 {
		t.Fatal("No Hosts found")
	}

	for i, host := range hosts {
		if host.HostID == "" {
			t.Fatalf("Host %d returned in response body has no Host ID", i)
		}
	}

	t.Logf("Validated %d Hosts", len(hosts))
}
