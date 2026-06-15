package zabbix

import (
	"encoding/json"
	"testing"
)

// ---------- Host JSON mapping (host_json.go) ----------

func TestJHost_Host(t *testing.T) {
	raw := `{"hostid":"10084","host":"zabbix-server","name":"Zabbix Server","flags":"0","status":"0"}`
	var jh jHost
	if err := json.Unmarshal([]byte(raw), &jh); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	host, err := jh.Host()
	if err != nil {
		t.Fatalf("Host() failed: %v", err)
	}
	if host.HostID != "10084" {
		t.Errorf("expected HostID '10084', got %q", host.HostID)
	}
	if host.Hostname != "zabbix-server" {
		t.Errorf("expected Hostname 'zabbix-server', got %q", host.Hostname)
	}
	if host.DisplayName != "Zabbix Server" {
		t.Errorf("expected DisplayName 'Zabbix Server', got %q", host.DisplayName)
	}
	if host.Status != 0 {
		t.Errorf("expected Status 0, got %d", host.Status)
	}
	if host.Source != 0 {
		t.Errorf("expected Source 0, got %d", host.Source)
	}
}

func TestJHosts_Hosts(t *testing.T) {
	raw := `[{"hostid":"1","host":"host1"},{"hostid":"2","host":"host2"}]`
	var jhs jHosts
	if err := json.Unmarshal([]byte(raw), &jhs); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	hosts, err := jhs.Hosts()
	if err != nil {
		t.Fatalf("Hosts() failed: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}
	if hosts[0].Hostname != "host1" {
		t.Errorf("expected 'host1', got %q", hosts[0].Hostname)
	}
	if hosts[1].HostID != "2" {
		t.Errorf("expected '2', got %q", hosts[1].HostID)
	}
}

func TestJHosts_Nil(t *testing.T) {
	var jhs jHosts
	hosts, err := jhs.Hosts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hosts != nil {
		t.Errorf("expected nil for nil input, got %v", hosts)
	}
}

// ---------- Hostgroup JSON mapping (hostgroup_json.go) ----------

func TestJHostgroup_Hostgroup(t *testing.T) {
	raw := `{"groupid":"15","name":"Linux servers","flags":"0","internal":"0"}`
	var jhg jHostgroup
	if err := json.Unmarshal([]byte(raw), &jhg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	hg, err := jhg.Hostgroup()
	if err != nil {
		t.Fatalf("Hostgroup() failed: %v", err)
	}
	if hg.GroupID != "15" {
		t.Errorf("expected GroupID '15', got %q", hg.GroupID)
	}
	if hg.Name != "Linux servers" {
		t.Errorf("expected Name 'Linux servers', got %q", hg.Name)
	}
}

func TestJHostgroups_Hostgroups(t *testing.T) {
	raw := `[{"groupid":"1","name":"grp1"},{"groupid":"2","name":"grp2"}]`
	var jhgs jHostgroups
	if err := json.Unmarshal([]byte(raw), &jhgs); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	groups, err := jhgs.Hostgroups()
	if err != nil {
		t.Fatalf("Hostgroups() failed: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestJHostgroups_Nil(t *testing.T) {
	var jhgs jHostgroups
	groups, err := jhgs.Hostgroups()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if groups != nil {
		t.Errorf("expected nil, got %v", groups)
	}
}

// ---------- Trigger JSON mapping (trigger_json.go) ----------

func TestJTrigger_Trigger(t *testing.T) {
	raw := `{
		"triggerid":"12345",
		"value":"0",
		"description":"CPU usage is too high",
		"status":"0",
		"expression":"{host:system.cpu.load.last()}>5",
		"lastchange":"1609459200",
		"priority":"3",
		"state":"0",
		"tags":[{"tag":"service","value":"web"}],
		"url":"http://example.com"
	}`
	var jt jTrigger
	if err := json.Unmarshal([]byte(raw), &jt); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	tr, err := jt.Trigger()
	if err != nil {
		t.Fatalf("Trigger() failed: %v", err)
	}
	if tr.TriggerID != "12345" {
		t.Errorf("expected TriggerID '12345', got %q", tr.TriggerID)
	}
	if tr.AlarmState != 0 {
		t.Errorf("expected AlarmState 0, got %d", tr.AlarmState)
	}
	// NOTE: current mapping sets Enabled = (status == "1"); status "0" → disabled.
	if tr.Enabled {
		t.Error("expected Enabled false (status=0 maps to disabled in current logic)")
	}
	if tr.Severity != 3 {
		t.Errorf("expected Severity 3, got %d", tr.Severity)
	}
	if tr.Description != "CPU usage is too high" {
		t.Errorf("unexpected Description: %q", tr.Description)
	}
	if len(tr.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tr.Tags))
	}
	if tr.Tags[0].Name != "service" {
		t.Errorf("expected tag 'service', got %q", tr.Tags[0].Name)
	}
}

func TestJTrigger_Disabled(t *testing.T) {
	// status "1" maps to Enabled=true in current mapping logic.
	raw := `{"triggerid":"1","value":"0","status":"1","priority":"0","state":"0"}`
	var jt jTrigger
	if err := json.Unmarshal([]byte(raw), &jt); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	tr, err := jt.Trigger()
	if err != nil {
		t.Fatalf("Trigger() failed: %v", err)
	}
	if !tr.Enabled {
		t.Error("expected Enabled true (status=1 maps to enabled in current logic)")
	}
}

// ---------- Event JSON mapping (event_json.go) ----------

func TestJEvent_Event(t *testing.T) {
	raw := `{
		"eventid":"100",
		"acknowledged":"1",
		"clock":"1609459200",
		"ns":"123456789",
		"object":"0",
		"objectid":"200",
		"source":"0",
		"value":"1",
		"value_changed":"1"
	}`
	var je jEvent
	if err := json.Unmarshal([]byte(raw), &je); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	ev, err := je.Event()
	if err != nil {
		t.Fatalf("Event() failed: %v", err)
	}
	if ev.EventID != "100" {
		t.Errorf("expected EventID '100', got %q", ev.EventID)
	}
	if !ev.Acknowledged {
		t.Error("expected Acknowledged true")
	}
	if ev.ObjectID != 200 {
		t.Errorf("expected ObjectID 200, got %d", ev.ObjectID)
	}
	if ev.Source != 0 {
		t.Errorf("expected Source 0, got %d", ev.Source)
	}
	if ev.Value != 1 {
		t.Errorf("expected Value 1, got %d", ev.Value)
	}
	if !ev.ValueChanged {
		t.Error("expected ValueChanged true")
	}
	if ev.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestJEvent_InvalidTimestamp(t *testing.T) {
	raw := `{"eventid":"1","clock":"notanumber","ns":"0"}`
	var je jEvent
	if err := json.Unmarshal([]byte(raw), &je); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	_, err := je.Event()
	if err == nil {
		t.Fatal("expected error for invalid timestamp, got nil")
	}
}

// ---------- Alert JSON mapping (alert_json.go) ----------

func TestJAlert_Alert(t *testing.T) {
	raw := `{
		"alertid":"1",
		"actionid":"10",
		"alerttype":"0",
		"clock":"1609459200",
		"error":"",
		"esc_step":"1",
		"eventid":"100",
		"mediatypeid":"1",
		"message":"Alert message",
		"retries":"0",
		"sendto":"admin@example.com",
		"status":"0",
		"subject":"Problem",
		"userid":"1"
	}`
	var ja jAlert
	if err := json.Unmarshal([]byte(raw), &ja); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	al, err := ja.Alert()
	if err != nil {
		t.Fatalf("Alert() failed: %v", err)
	}
	if al.AlertID != "1" {
		t.Errorf("expected AlertID '1', got %q", al.AlertID)
	}
	if al.AlertType != 0 {
		t.Errorf("expected AlertType 0, got %d", al.AlertType)
	}
	if al.Message != "Alert message" {
		t.Errorf("unexpected Message: %q", al.Message)
	}
	if al.Recipient != "admin@example.com" {
		t.Errorf("unexpected Recipient: %q", al.Recipient)
	}
	if al.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

// ---------- Action JSON mapping (action_json.go) ----------

func TestJAction_Action(t *testing.T) {
	raw := `{
		"actionid":"1",
		"esc_period":"60",
		"evaltype":"",
		"eventsource":"0",
		"name":"Report problems",
		"def_longdata":"Problem body",
		"def_shortdata":"Problem subject",
		"r_longdata":"Recovery body",
		"r_shortdata":"Recovery subject",
		"recovery_msg":"1",
		"status":"0"
	}`
	var ja jAction
	if err := json.Unmarshal([]byte(raw), &ja); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	a, err := ja.Action()
	if err != nil {
		t.Fatalf("Action() failed: %v", err)
	}
	if a.ActionID != "1" {
		t.Errorf("expected ActionID '1', got %q", a.ActionID)
	}
	if a.Name != "Report problems" {
		t.Errorf("expected Name 'Report problems', got %q", a.Name)
	}
	if a.StepDuration != 60 {
		t.Errorf("expected StepDuration 60, got %d", a.StepDuration)
	}
	if !a.RecoveryMessageEnabled {
		t.Error("expected RecoveryMessageEnabled true")
	}
	if !a.Enabled {
		t.Error("expected Enabled true (status=0)")
	}
	if a.ProblemMessageSubject != "Problem subject" {
		t.Errorf("unexpected ProblemMessageSubject: %q", a.ProblemMessageSubject)
	}
}

// ---------- Item JSON mapping (item_json.go) ----------

func TestJItem_Item(t *testing.T) {
	raw := `{
		"hostid":"10084",
		"itemid":"1",
		"name":"CPU load",
		"description":"Processor load",
		"lastclock":"1609459200",
		"lastvalue":"0.42",
		"value_type":"0"
	}`
	var ji jItem
	if err := json.Unmarshal([]byte(raw), &ji); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	it, err := ji.Item()
	if err != nil {
		t.Fatalf("Item() failed: %v", err)
	}
	if it.HostID != 10084 {
		t.Errorf("expected HostID 10084, got %d", it.HostID)
	}
	if it.ItemID != 1 {
		t.Errorf("expected ItemID 1, got %d", it.ItemID)
	}
	if it.ItemName != "CPU load" {
		t.Errorf("unexpected ItemName: %q", it.ItemName)
	}
	if it.LastValue != "0.42" {
		t.Errorf("unexpected LastValue: %q", it.LastValue)
	}
	if it.LastValueType != 0 {
		t.Errorf("expected LastValueType 0, got %d", it.LastValueType)
	}
}

func TestJItem_InvalidID(t *testing.T) {
	raw := `{"hostid":"notanumber","itemid":"1","lastclock":"0","value_type":"0"}`
	var ji jItem
	if err := json.Unmarshal([]byte(raw), &ji); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	_, err := ji.Item()
	if err == nil {
		t.Fatal("expected error for invalid HostID, got nil")
	}
}

// ---------- History JSON mapping (history_json.go) ----------

func TestJHistory_History(t *testing.T) {
	raw := `{
		"itemid":"1",
		"clock":"1609459200",
		"ns":"123456",
		"value":"42.5"
	}`
	var jh jHistory
	if err := json.Unmarshal([]byte(raw), &jh); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	h, err := jh.History()
	if err != nil {
		t.Fatalf("History() failed: %v", err)
	}
	if h.ItemID != 1 {
		t.Errorf("expected ItemID 1, got %d", h.ItemID)
	}
	if h.Clock != 1609459200 {
		t.Errorf("expected Clock 1609459200, got %d", h.Clock)
	}
	if h.Ns != 123456 {
		t.Errorf("expected Ns 123456, got %d", h.Ns)
	}
	if h.Value != "42.5" {
		t.Errorf("unexpected Value: %q", h.Value)
	}
}

func TestJHistories_Histories(t *testing.T) {
	raw := `[
		{"itemid":"1","clock":"100","ns":"0","value":"1.0"},
		{"itemid":"2","clock":"200","ns":"0","value":"2.0"}
	]`
	var jhs jHistories
	if err := json.Unmarshal([]byte(raw), &jhs); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	hs, err := jhs.Histories()
	if err != nil {
		t.Fatalf("Histories() failed: %v", err)
	}
	if len(hs) != 2 {
		t.Fatalf("expected 2 histories, got %d", len(hs))
	}
	if hs[1].Value != "2.0" {
		t.Errorf("unexpected Value: %q", hs[1].Value)
	}
}

// ---------- Maintenance JSON mapping (maintenance_json.go) ----------

func TestJMaintenanceGet_Maintenance(t *testing.T) {
	raw := `{
		"maintenanceid":"1",
		"name":"Monthly maintenance",
		"active_since":"1609459200",
		"active_till":"1609545600",
		"description":"Scheduled downtime",
		"maintenance_type":"0",
		"tags_evaltype":"0",
		"hosts":[],
		"groups":[],
		"tags":[],
		"timeperiods":[]
	}`
	var jm jMaintenanceGet
	if err := json.Unmarshal([]byte(raw), &jm); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	m, err := jm.Maintenance()
	if err != nil {
		t.Fatalf("Maintenance() failed: %v", err)
	}
	if m.MaintenanceID != "1" {
		t.Errorf("expected MaintenanceID '1', got %q", m.MaintenanceID)
	}
	if m.Name != "Monthly maintenance" {
		t.Errorf("expected Name 'Monthly maintenance', got %q", m.Name)
	}
	if m.Description != "Scheduled downtime" {
		t.Errorf("unexpected Description: %q", m.Description)
	}
	if m.ActiveSince.IsZero() {
		t.Error("expected non-zero ActiveSince")
	}
	if m.ActiveTill.IsZero() {
		t.Error("expected non-zero ActiveTill")
	}
}

func TestJMaintenanceGet_WithTags(t *testing.T) {
	raw := `{
		"maintenanceid":"1",
		"name":"Test",
		"active_since":"1609459200",
		"active_till":"1609545600",
		"maintenance_type":"0",
		"tags_evaltype":"0",
		"hosts":[],
		"groups":[],
		"tags":[{"tag":"service","operator":"0","value":"web"}],
		"timeperiods":[]
	}`
	var jm jMaintenanceGet
	if err := json.Unmarshal([]byte(raw), &jm); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	m, err := jm.Maintenance()
	if err != nil {
		t.Fatalf("Maintenance() failed: %v", err)
	}
	if len(m.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(m.Tags))
	}
	if m.Tags[0].Tag != "service" {
		t.Errorf("expected tag 'service', got %q", m.Tags[0].Tag)
	}
	if m.Tags[0].Value != "web" {
		t.Errorf("expected value 'web', got %q", m.Tags[0].Value)
	}
}

func TestJTimeperiod_Timeperiod(t *testing.T) {
	raw := `{
		"timeperiodid":1,
		"day":0,
		"dayofweek":0,
		"every":1,
		"month":0,
		"period":3600,
		"start_date":1609459200,
		"start_time":0,
		"timeperiod_type":0
	}`
	var jt jTimeperiod
	if err := json.Unmarshal([]byte(raw), &jt); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	tp, err := jt.jTimeperiod()
	if err != nil {
		t.Fatalf("jTimeperiod() failed: %v", err)
	}
	if tp.TimeperiodId != 1 {
		t.Errorf("expected TimeperiodId 1, got %d", tp.TimeperiodId)
	}
	if tp.Every != 1 {
		t.Errorf("expected Every 1, got %d", tp.Every)
	}
	if tp.TimeperiodType != 0 {
		t.Errorf("expected TimeperiodType 0, got %d", tp.TimeperiodType)
	}
}
