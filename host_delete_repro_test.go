package zabbix

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// countString возвращает число вхождений s в slice.
func countString(slice []string, s string) int {
	n := 0
	for _, v := range slice {
		if v == s {
			n++
		}
	}
	return n
}

func itoa(i int) string {
	digits := "0123456789"
	if i < 10 {
		return digits[i : i+1]
	}
	return itoa(i/10) + digits[i%10:i%10+1]
}

// Хост входит в несколько Maintenance. Раньше это приводило к дубликату
// hostid в params повторного host.delete и ошибке Zabbix:
//
//	Invalid parameter "/7": value (14806) already exists. (-32602)
//
// Теперь HostDelete обязан исключать дубликаты из повторной попытки.
func TestHostDelete_MaintenanceRetryDeduplicatesHostIDs(t *testing.T) {
	m, session := newMockSession(t, "7.0.0")
	defer m.Close()

	// 8 хостов как в проде (дубликат был на позиции "/7" - 8-й элемент)
	hosts := []Host{
		{HostID: "15166"}, {HostID: "15001"}, {HostID: "14806"}, {HostID: "14999"},
		{HostID: "15178"}, {HostID: "15350"}, {HostID: "14277"}, {HostID: "14600"},
	}

	var deleteCalls [][]string

	// Первый host.delete - падает (любая ошибка: хост уже удалён, discovered и т.п.)
	m.handle("host.delete", func(req *Request) (interface{}, *APIError) {
		raw, _ := json.Marshal(req.Params)
		var ids []string
		if err := json.Unmarshal(raw, &ids); err != nil {
			t.Fatalf("host.delete params is not []string: %s", raw)
		}
		deleteCalls = append(deleteCalls, ids)

		if len(deleteCalls) == 1 {
			return nil, &APIError{
				Code:    -32602,
				Message: "Invalid params.",
				Data:    `Invalid parameter "/7": value (14806) already exists.`,
			}
		}

		// Повторная попытка: проверяем отсутствие дубликатов
		for i, id := range ids {
			for j := 0; j < i; j++ {
				if ids[j] == id {
					t.Errorf("дубликат hostid %q на позиции /%d в повторном host.delete: %v", id, i, ids)
				}
			}
		}
		return map[string][]string{"hostids": ids}, nil
	})

	// Два maintenance, перекрывающихся по хосту 14806
	m.handle("maintenance.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{
				"maintenanceid": "10",
				"name":          "maint-A",
				"hosts": []map[string]string{
					{"hostid": "15166", "host": "h1"},
					{"hostid": "15001", "host": "h2"},
					{"hostid": "14806", "host": "h3"},
				},
			},
			{
				"maintenanceid": "11",
				"name":          "maint-B",
				"hosts": []map[string]string{
					{"hostid": "14806", "host": "h3"},
					{"hostid": "14999", "host": "h4"},
					{"hostid": "15178", "host": "h5"},
					{"hostid": "15350", "host": "h6"},
					{"hostid": "14277", "host": "h7"},
					{"hostid": "14600", "host": "h8"},
				},
			},
		}, nil
	})

	maintenanceDeletes := 0
	m.handle("maintenance.delete", func(req *Request) (interface{}, *APIError) {
		maintenanceDeletes++
		return map[string][]string{"maintenanceids": {"x"}}, nil
	})

	resp, err := session.HostDelete(context.Background(), hosts...)

	if len(deleteCalls) != 2 {
		t.Fatalf("expected 2 host.delete calls (batch + retry), got %d", len(deleteCalls))
	}
	if n := countString(deleteCalls[1], "14806"); n != 1 {
		t.Errorf("expected unique hostid 14806 in retry params, got %d times: %v", n, deleteCalls[1])
	}
	if maintenanceDeletes != 2 {
		t.Errorf("expected 2 maintenance.delete calls, got %d", maintenanceDeletes)
	}
	if err != nil {
		t.Errorf("expected nil error after successful retry, got %v", err)
	}
	if len(resp) != len(hosts) {
		t.Errorf("expected all %d hostids in response, got %d: %v", len(hosts), len(resp), resp)
	}
}

// Если maintenance нет (maintenance.get -> ErrNotFound), наружу должна
// возвращаться исходная ошибка удаления, а не ErrNotFound.
func TestHostDelete_NoMaintenanceReturnsOriginalError(t *testing.T) {
	m, session := newMockSession(t, "7.0.0")
	defer m.Close()

	hosts := []Host{{HostID: "15166"}, {HostID: "15001"}}

	m.handle("host.delete", func(req *Request) (interface{}, *APIError) {
		return nil, &APIError{
			Code:    -32500,
			Message: "Application error.",
			Data:    "No permissions to referred object or it does not exist!",
		}
	})
	m.handle("maintenance.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{}, nil
	})

	maintenanceDeletes := 0
	m.handle("maintenance.delete", func(req *Request) (interface{}, *APIError) {
		maintenanceDeletes++
		return map[string][]string{"maintenanceids": {"x"}}, nil
	})

	resp, err := session.HostDelete(context.Background(), hosts...)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "No permissions") {
		t.Errorf("expected original delete error, got: %v", err)
	}
	if strings.Contains(err.Error(), ErrNotFound.Error()) {
		t.Errorf("ErrNotFound must not replace the original error, got: %v", err)
	}
	if maintenanceDeletes != 0 {
		t.Errorf("expected no maintenance.delete calls, got %d", maintenanceDeletes)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty response, got %v", resp)
	}
}

// Если повторная попытка тоже падает, наружу должны возвращаться обе ошибки.
func TestHostDelete_RetryFailureCombinesErrors(t *testing.T) {
	m, session := newMockSession(t, "7.0.0")
	defer m.Close()

	hosts := []Host{{HostID: "15166"}, {HostID: "15001"}}

	attempt := 0
	m.handle("host.delete", func(req *Request) (interface{}, *APIError) {
		attempt++
		return nil, &APIError{
			Code:    -32500,
			Message: "Application error.",
			Data:    "attempt " + itoa(attempt) + " failed",
		}
	})
	m.handle("maintenance.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{
				"maintenanceid": "10",
				"name":          "maint-A",
				"hosts": []map[string]string{
					{"hostid": "15166", "host": "h1"},
					{"hostid": "15001", "host": "h2"},
				},
			},
		}, nil
	})
	m.handle("maintenance.delete", func(req *Request) (interface{}, *APIError) {
		return map[string][]string{"maintenanceids": {"x"}}, nil
	})

	_, err := session.HostDelete(context.Background(), hosts...)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "attempt 1 failed") {
		t.Errorf("original error lost, got: %v", err)
	}
	if !strings.Contains(err.Error(), "attempt 2 failed") {
		t.Errorf("retry error lost, got: %v", err)
	}
}
