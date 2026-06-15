package zabbix

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ---------- Host CRUD ----------

func TestGetHosts_Success(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})
	m.handle("user.login", func(req *Request) (interface{}, *APIError) {
		return "token", nil
	})
	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"hostid": "1", "host": "server1", "name": "Server 1"},
			{"hostid": "2", "host": "server2", "name": "Server 2"},
		}, nil
	})

	session, err := NewSession(context.Background(), m.URL(), "Admin", "zabbix", "")
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	hosts, err := session.GetHosts(context.Background(), HostGetParams{})
	if err != nil {
		t.Fatalf("GetHosts failed: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}
	if hosts[0].HostID != "1" {
		t.Errorf("expected HostID '1', got %q", hosts[0].HostID)
	}
	if hosts[1].Hostname != "server2" {
		t.Errorf("expected Hostname 'server2', got %q", hosts[1].Hostname)
	}
}

func TestGetHosts_NotFound(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []interface{}{}, nil
	})

	_, err := session.GetHosts(context.Background(), HostGetParams{})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestHostCreate_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("host.create", func(req *Request) (interface{}, *APIError) {
		return map[string]interface{}{
			"hostids": []string{"10084"},
		}, nil
	})

	ids, err := session.HostCreate(context.Background(), Host{
		Hostname: "new-host",
	})
	if err != nil {
		t.Fatalf("HostCreate failed: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 ID, got %d", len(ids))
	}
	if ids[0] != "10084" {
		t.Errorf("expected '10084', got %q", ids[0])
	}
}

func TestHostDelete_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("host.delete", func(req *Request) (interface{}, *APIError) {
		return map[string]interface{}{
			"hostids": []string{"10084"},
		}, nil
	})

	ids, err := session.HostDelete(context.Background(), Host{HostID: "10084"})
	if err != nil {
		t.Fatalf("HostDelete failed: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 deleted ID, got %d", len(ids))
	}
}

// ---------- Hostgroup CRUD ----------

func TestGetHostgroups_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("hostgroup.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"groupid": "1", "name": "Linux servers"},
			{"groupid": "2", "name": "Windows servers"},
		}, nil
	})

	groups, err := session.GetHostgroups(context.Background(), HostgroupGetParams{})
	if err != nil {
		t.Fatalf("GetHostgroups failed: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Name != "Linux servers" {
		t.Errorf("unexpected Name: %q", groups[0].Name)
	}
}

// ---------- UserMacro CRUD ----------

func TestGetUserMacro_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("usermacro.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"hostmacroid": "1", "macro": "{$SNMP_COMMUNITY}", "value": "public"},
		}, nil
	})

	macros, err := session.GetUserMacro(context.Background(), UserMacroGetParams{})
	if err != nil {
		t.Fatalf("GetUserMacro failed: %v", err)
	}
	if len(macros) != 1 {
		t.Fatalf("expected 1 macro, got %d", len(macros))
	}
	if macros[0].Macro != "{$SNMP_COMMUNITY}" {
		t.Errorf("unexpected Macro: %q", macros[0].Macro)
	}
	if macros[0].Value != "public" {
		t.Errorf("unexpected Value: %q", macros[0].Value)
	}
}

func TestCreateUserMacros_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("usermacro.create", func(req *Request) (interface{}, *APIError) {
		return map[string]interface{}{
			"hostmacroids": []string{"1", "2"},
		}, nil
	})

	ids, err := session.CreateUserMacros(context.Background(),
		HostMacro{Macro: "{$M1}", Value: "v1"},
		HostMacro{Macro: "{$M2}", Value: "v2"},
	)
	if err != nil {
		t.Fatalf("CreateUserMacros failed: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(ids))
	}
}

func TestDeleteUserMacros_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("usermacro.delete", func(req *Request) (interface{}, *APIError) {
		return map[string]interface{}{
			"hostmacroids": []string{"1"},
		}, nil
	})

	ids, err := session.DeleteUserMacros(context.Background(), "1")
	if err != nil {
		t.Fatalf("DeleteUserMacros failed: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 ID, got %d", len(ids))
	}
}

func TestUpdateUserMacros_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("usermacro.update", func(req *Request) (interface{}, *APIError) {
		return map[string]interface{}{
			"hostmacroids": []string{"1"},
		}, nil
	})

	ids, err := session.UpdateUserMacros(context.Background(),
		HostMacro{HostMacroID: "1", Value: "new-value"},
	)
	if err != nil {
		t.Fatalf("UpdateUserMacros failed: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 ID, got %d", len(ids))
	}
}

// ---------- Maintenance CRUD ----------

func TestGetMaintenance_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("maintenance.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{
				"maintenanceid":     "1",
				"name":              "Weekly",
				"active_since":      "1609459200",
				"active_till":       "1609545600",
				"description":       "Weekly downtime",
				"maintenance_type":  "0",
				"tags_evaltype":     "0",
				"hosts":             []interface{}{},
				"groups":            []interface{}{},
				"tags":              []interface{}{},
				"timeperiods":       []interface{}{},
			},
		}, nil
	})

	maintenances, err := session.GetMaintenance(context.Background(), &MaintenanceGetParams{})
	if err != nil {
		t.Fatalf("GetMaintenance failed: %v", err)
	}
	if len(maintenances) != 1 {
		t.Fatalf("expected 1 maintenance, got %d", len(maintenances))
	}
	if maintenances[0].MaintenanceID != "1" {
		t.Errorf("expected MaintenanceID '1', got %q", maintenances[0].MaintenanceID)
	}
	if maintenances[0].Name != "Weekly" {
		t.Errorf("unexpected Name: %q", maintenances[0].Name)
	}
	if maintenances[0].Session == nil {
		t.Error("expected Session to be set")
	}
}

func TestMaintenance_Delete(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("maintenance.delete", func(req *Request) (interface{}, *APIError) {
		return map[string]interface{}{
			"maintenanceids": []string{"1"},
		}, nil
	})

	maint := &Maintenance{
		Session:       session,
		MaintenanceID: "1",
	}
	if err := maint.Delete(context.Background()); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// ---------- Trigger, Event, Action, Alert, Item, History via mock ----------

func TestGetTriggers_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("trigger.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"triggerid": "1", "value": "0", "description": "Test", "status": "0",
				"lastchange": "1609459200", "priority": "1", "state": "0"},
		}, nil
	})

	triggers, err := session.GetTriggers(context.Background(), TriggerGetParams{})
	if err != nil {
		t.Fatalf("GetTriggers failed: %v", err)
	}
	if len(triggers) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(triggers))
	}
	if triggers[0].TriggerID != "1" {
		t.Errorf("expected TriggerID '1', got %q", triggers[0].TriggerID)
	}
}

func TestGetEvents_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("event.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"eventid": "1", "acknowledged": "0", "clock": "1609459200",
				"ns": "0", "object": "0", "objectid": "100", "source": "0",
				"value": "1", "value_changed": "1"},
		}, nil
	})

	events, err := session.GetEvents(context.Background(), EventGetParams{})
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Value != 1 {
		t.Errorf("expected Value 1, got %d", events[0].Value)
	}
}

func TestGetActions_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("action.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"actionid": "1", "esc_period": "60", "eventsource": "0",
				"name": "Alert", "status": "0", "recovery_msg": "0",
				"def_shortdata": "subj", "def_longdata": "body",
				"r_shortdata": "", "r_longdata": "", "evaltype": ""},
		}, nil
	})

	actions, err := session.GetActions(context.Background(), ActionGetParams{})
	if err != nil {
		t.Fatalf("GetActions failed: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Name != "Alert" {
		t.Errorf("unexpected Name: %q", actions[0].Name)
	}
}

func TestGetItems_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("item.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"hostid": "10084", "itemid": "1", "name": "CPU load",
				"lastclock": "1609459200", "lastvalue": "1.5", "value_type": "0"},
		}, nil
	})

	items, err := session.GetItems(context.Background(), ItemGetParams{})
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ItemName != "CPU load" {
		t.Errorf("unexpected ItemName: %q", items[0].ItemName)
	}
}

func TestGetHistories_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("history.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"itemid": "1", "clock": "1609459200", "ns": "0", "value": "42.5"},
		}, nil
	})

	histories, err := session.GetHistories(context.Background(), HistoryGetParams{})
	if err != nil {
		t.Fatalf("GetHistories failed: %v", err)
	}
	if len(histories) != 1 {
		t.Fatalf("expected 1 history, got %d", len(histories))
	}
	if histories[0].Value != "42.5" {
		t.Errorf("unexpected Value: %q", histories[0].Value)
	}
}

func TestGetAlerts_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("alert.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"alertid": "1", "actionid": "10", "alerttype": "0",
				"clock": "1609459200", "esc_step": "1", "eventid": "100",
				"mediatypeid": "1", "message": "msg", "retries": "0",
				"sendto": "admin@test.com", "status": "0", "subject": "subj",
				"userid": "1"},
		}, nil
	})

	alerts, err := session.GetAlerts(context.Background(), AlertGetParams{})
	if err != nil {
		t.Fatalf("GetAlerts failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
}

func TestGetHostInterfaces_Success(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("hostinterface.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"interfaceid": "1", "hostid": "10084", "ip": "127.0.0.1",
				"dns": "localhost", "port": "10050", "type": "1",
				"main": "0", "useip": "1", "available": "1"},
		}, nil
	})

	ifaces, err := session.GetHostInterfaces(context.Background(), HostInterfaceGetParams{})
	if err != nil {
		t.Fatalf("GetHostInterfaces failed: %v", err)
	}
	if len(ifaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(ifaces))
	}
	if ifaces[0].HostID != "10084" {
		t.Errorf("expected HostID '10084', got %q", ifaces[0].HostID)
	}
}

// ---------- Maintenance FillHostIDs ----------

func TestFillHostIDs_V6_Found(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"hostid": "10084", "host": "my-host"},
		}, nil
	})

	params := &jMaintenanceCreateParams{
		Hosts: jHosts{{Hostname: "my-host"}},
	}
	err := params.FillHostIDs(context.Background(), session)
	if err != nil {
		t.Fatalf("FillHostIDs failed: %v", err)
	}
	if len(params.Hosts) != 1 || params.Hosts[0].HostID != "10084" {
		t.Errorf("expected HostID '10084' to be filled, got %+v", params.Hosts)
	}
}

func TestFillHostIDs_V6_NotFound(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{"hostid": "10084", "host": "other-host"},
		}, nil
	})

	params := &jMaintenanceCreateParams{
		Hosts: jHosts{{Hostname: "missing-host"}},
	}
	err := params.FillHostIDs(context.Background(), session)
	if !errors.Is(err, ErrMaintenanceHostNotFound) {
		t.Errorf("expected ErrMaintenanceHostNotFound, got %v", err)
	}
}

// ---------- context propagation to API calls ----------

func TestGetHosts_ContextCancelled(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		time.Sleep(100 * time.Millisecond)
		return []interface{}{}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := session.GetHosts(ctx, HostGetParams{})
	if err == nil {
		t.Fatal("expected context deadline error, got nil")
	}
}
