package zabbix

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestWriteMethods(t *testing.T) {
	ctx := context.Background()
	session := GetTestSession(t)

	groups, err := session.GetHostgroups(ctx, HostgroupGetParams{
		GetParameters: GetParameters{
			EditableOnly: true,
			ResultLimit:  1,
		},
	})
	if errors.Is(err, ErrNotFound) {
		t.Skip("The account has read-only access")
	}
	if err != nil {
		t.Fatalf("Error checking write access: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	hostName := "go-zabbix-integration-" + suffix
	hostDisplayName := "go-zabbix integration " + suffix
	host := Host{
		Hostname:    hostName,
		DisplayName: hostDisplayName,
		Status:      HostStatusUnmonitored,
		Groups:      []Hostgroup{{GroupID: groups[0].GroupID}},
	}

	hostIDs, err := session.HostCreate(ctx, host)
	if err != nil {
		if isWritePermissionError(err) {
			t.Skipf("The account has read-only access: %v", err)
		}
		t.Fatalf("Error creating test host: %v", err)
	}
	if len(hostIDs) != 1 || hostIDs[0] == "" {
		t.Fatalf("Unexpected host.create response: %v", hostIDs)
	}
	host.HostID = hostIDs[0]
	hostDeleted := false
	t.Cleanup(func() {
		if hostDeleted {
			return
		}
		if _, cleanupErr := session.HostDelete(context.Background(), host); cleanupErr != nil {
			t.Errorf("Cannot delete test host %s: %v", host.HostID, cleanupErr)
		}
	})

	createdHosts, err := session.GetHosts(ctx, HostGetParams{HostIDs: []string{host.HostID}})
	if err != nil || len(createdHosts) != 1 || createdHosts[0].Hostname != hostName {
		t.Fatalf("Cannot read created host: hosts=%v, err=%v", createdHosts, err)
	}

	host.Description = "updated by go-zabbix integration test"
	host.DisplayName += " updated"
	updatedHostIDs, err := session.HostUpdate(ctx, host)
	if err != nil {
		t.Fatalf("Error updating test host: %v", err)
	}
	if !containsString(updatedHostIDs, host.HostID) {
		t.Fatalf("Unexpected host.update response: %v", updatedHostIDs)
	}
	updatedHosts, err := session.GetHosts(ctx, HostGetParams{HostIDs: []string{host.HostID}})
	if err != nil || len(updatedHosts) != 1 || updatedHosts[0].Description != host.Description {
		t.Fatalf("Cannot verify updated host: hosts=%v, err=%v", updatedHosts, err)
	}

	macro := HostMacro{
		HostID:      host.HostID,
		Macro:       "{$GO_ZABBIX_INTEGRATION}",
		Value:       "created",
		Description: "go-zabbix integration test",
	}
	macroIDs, err := session.CreateUserMacros(ctx, macro)
	if err != nil {
		t.Fatalf("Error creating test macro: %v", err)
	}
	if len(macroIDs) != 1 || macroIDs[0] == "" {
		t.Fatalf("Unexpected usermacro.create response: %v", macroIDs)
	}
	macro.HostMacroID = macroIDs[0]
	macroDeleted := false
	t.Cleanup(func() {
		if macroDeleted {
			return
		}
		if _, cleanupErr := session.DeleteUserMacros(context.Background(), macro.HostMacroID); cleanupErr != nil && !errors.Is(cleanupErr, ErrNotFound) {
			t.Errorf("Cannot delete test macro %s: %v", macro.HostMacroID, cleanupErr)
		}
	})

	createdMacros, err := session.GetUserMacro(ctx, UserMacroGetParams{HostMacroIDs: []string{macro.HostMacroID}})
	if err != nil || len(createdMacros) != 1 || createdMacros[0].Value != macro.Value {
		t.Fatalf("Cannot read created macro: macros=%v, err=%v", createdMacros, err)
	}

	macro.Value = "updated"
	updatedMacroIDs, err := session.UpdateUserMacros(ctx, HostMacro{
		HostMacroID: macro.HostMacroID,
		Macro:       macro.Macro,
		Value:       macro.Value,
		Description: macro.Description,
	})
	if err != nil {
		t.Fatalf("Error updating test macro: %v", err)
	}
	if !containsString(updatedMacroIDs, macro.HostMacroID) {
		t.Fatalf("Unexpected usermacro.update response: %v", updatedMacroIDs)
	}
	updatedMacros, err := session.GetUserMacro(ctx, UserMacroGetParams{HostMacroIDs: []string{macro.HostMacroID}})
	if err != nil || len(updatedMacros) != 1 || updatedMacros[0].Value != macro.Value {
		t.Fatalf("Cannot verify updated macro: macros=%v, err=%v", updatedMacros, err)
	}

	now := time.Now().Truncate(time.Second)
	maintenance := Maintenance{
		Session:     session,
		Name:        "go-zabbix integration " + suffix,
		Description: "created by go-zabbix integration test",
		ActiveSince: now,
		ActiveTill:  now.Add(time.Hour),
		Hosts:       []Host{{Hostname: hostName}},
		Timeperiods: []Timeperiod{{
			TimeperiodType: Once,
			StartDate:      now,
			Period:         time.Hour,
		}},
	}
	maintenanceResponse, err := maintenance.Create(ctx)
	if err != nil {
		t.Fatalf("Error creating test maintenance: %v", err)
	}
	if len(maintenanceResponse.IDs) != 1 || maintenanceResponse.IDs[0] == "" {
		t.Fatalf("Unexpected maintenance.create response: %v", maintenanceResponse.IDs)
	}
	maintenance.MaintenanceID = maintenanceResponse.IDs[0]
	maintenanceDeleted := false
	t.Cleanup(func() {
		if maintenanceDeleted {
			return
		}
		if cleanupErr := maintenance.Delete(context.Background()); cleanupErr != nil {
			t.Errorf("Cannot delete test maintenance %s: %v", maintenance.MaintenanceID, cleanupErr)
		}
	})

	createdMaintenances, err := session.GetMaintenance(ctx, &MaintenanceGetParams{
		Maintenanceids: []string{maintenance.MaintenanceID},
	})
	if err != nil || len(createdMaintenances) != 1 || createdMaintenances[0].Name != maintenance.Name {
		t.Fatalf("Cannot read created maintenance: maintenances=%v, err=%v", createdMaintenances, err)
	}

	maintenance.Description = "updated by go-zabbix integration test"
	updatedMaintenanceResponse, err := maintenance.Update(ctx)
	if err != nil {
		t.Fatalf("Error updating test maintenance: %v", err)
	}
	if !containsString(updatedMaintenanceResponse.IDs, maintenance.MaintenanceID) {
		t.Fatalf("Unexpected maintenance.update response: %v", updatedMaintenanceResponse.IDs)
	}
	updatedMaintenances, err := session.GetMaintenance(ctx, &MaintenanceGetParams{
		Maintenanceids: []string{maintenance.MaintenanceID},
	})
	if err != nil || len(updatedMaintenances) != 1 || updatedMaintenances[0].Description != maintenance.Description {
		t.Fatalf("Cannot verify updated maintenance: maintenances=%v, err=%v", updatedMaintenances, err)
	}

	if err := maintenance.Delete(ctx); err != nil {
		t.Fatalf("Error deleting test maintenance: %v", err)
	}
	maintenanceDeleted = true
	if _, err := session.GetMaintenance(ctx, &MaintenanceGetParams{Maintenanceids: []string{maintenance.MaintenanceID}}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Deleted maintenance is still available: %v", err)
	}

	deletedMacroIDs, err := session.DeleteUserMacros(ctx, macro.HostMacroID)
	if err != nil {
		t.Fatalf("Error deleting test macro: %v", err)
	}
	if !containsString(deletedMacroIDs, macro.HostMacroID) {
		t.Fatalf("Unexpected usermacro.delete response: %v", deletedMacroIDs)
	}
	macroDeleted = true
	if _, err := session.GetUserMacro(ctx, UserMacroGetParams{HostMacroIDs: []string{macro.HostMacroID}}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Deleted macro is still available: %v", err)
	}

	deletedHostIDs, err := session.HostDelete(ctx, host)
	if err != nil {
		t.Fatalf("Error deleting test host: %v", err)
	}
	if !containsString(deletedHostIDs, host.HostID) {
		t.Fatalf("Unexpected host.delete response: %v", deletedHostIDs)
	}
	hostDeleted = true
	if _, err := session.GetHosts(ctx, HostGetParams{HostIDs: []string{host.HostID}}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Deleted host is still available: %v", err)
	}
}

func isWritePermissionError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "permission") ||
		strings.Contains(message, "not authorized") ||
		strings.Contains(message, "not allowed")
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
