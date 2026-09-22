package zabbix

import (
	"context"
	"strings"
	"testing"
)

// Хост может входить сразу в несколько обслуживаний. Все обслуживания должны
// быть очищены до единственного вызова host.delete.
func TestHostDelete_CleansOverlappingMaintenancesBeforeDeletingHosts(t *testing.T) {
	m, session := newMockSession(t, "7.0.0")
	defer m.Close()

	hosts := []Host{
		{HostID: "15166"}, {HostID: "15001"}, {HostID: "14806"}, {HostID: "14999"},
		{HostID: "15178"}, {HostID: "15350"}, {HostID: "14277"}, {HostID: "14600"},
	}

	maintenanceDeletes := 0
	hostDeleteCalls := 0
	m.handle("maintenance.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{
				"maintenanceid": "10",
				"hosts": []map[string]string{
					{"hostid": "15166", "host": "h1"},
					{"hostid": "15001", "host": "h2"},
					{"hostid": "14806", "host": "h3"},
				},
				"groups": []interface{}{},
			},
			{
				"maintenanceid": "11",
				"hosts": []map[string]string{
					{"hostid": "14806", "host": "h3"},
					{"hostid": "14999", "host": "h4"},
					{"hostid": "15178", "host": "h5"},
					{"hostid": "15350", "host": "h6"},
					{"hostid": "14277", "host": "h7"},
					{"hostid": "14600", "host": "h8"},
				},
				"groups": []interface{}{},
			},
		}, nil
	})
	m.handle("maintenance.delete", func(req *Request) (interface{}, *APIError) {
		maintenanceDeletes++
		return map[string][]string{"maintenanceids": {"x"}}, nil
	})
	m.handle("host.delete", func(req *Request) (interface{}, *APIError) {
		hostDeleteCalls++
		if maintenanceDeletes != 2 {
			t.Errorf("host.delete вызван до очистки всех обслуживаний: %d", maintenanceDeletes)
		}
		var ids []string
		decodeRequestParams(t, req, &ids)
		return map[string][]string{"hostids": ids}, nil
	})

	response, err := session.HostDelete(context.Background(), hosts...)
	if err != nil {
		t.Fatalf("HostDelete завершился с ошибкой: %v", err)
	}
	if maintenanceDeletes != 2 {
		t.Errorf("ожидалось 2 вызова maintenance.delete, получено %d", maintenanceDeletes)
	}
	if hostDeleteCalls != 1 {
		t.Errorf("ожидался 1 вызов host.delete, получено %d", hostDeleteCalls)
	}
	if len(response) != len(hosts) {
		t.Errorf("ожидалось %d удалённых хостов, получено %d: %v", len(hosts), len(response), response)
	}
}

// Отсутствие обслуживаний не должно мешать удалению хостов и подменять ошибку
// метода host.delete на ErrNotFound.
func TestHostDelete_NoMaintenanceReturnsDeleteError(t *testing.T) {
	m, session := newMockSession(t, "7.0.0")
	defer m.Close()

	m.handle("maintenance.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{}, nil
	})
	m.handle("host.delete", func(req *Request) (interface{}, *APIError) {
		return nil, &APIError{
			Code:    -32500,
			Message: "Application error.",
			Data:    "No permissions to referred object or it does not exist!",
		}
	})

	response, err := session.HostDelete(context.Background(), Host{HostID: "15166"})
	if err == nil {
		t.Fatal("ожидалась ошибка host.delete")
	}
	if !strings.Contains(err.Error(), "No permissions") {
		t.Errorf("ожидалась исходная ошибка удаления, получено: %v", err)
	}
	if strings.Contains(err.Error(), ErrNotFound.Error()) {
		t.Errorf("ErrNotFound не должна подменять ошибку удаления: %v", err)
	}
	if len(response) != 0 {
		t.Errorf("ожидался пустой ответ, получено: %v", response)
	}
}

// После предварительной очистки обслуживаний ошибка host.delete возвращается
// вызывающему коду без скрытой повторной попытки.
func TestHostDelete_DeleteFailureIsReturnedWithoutRetry(t *testing.T) {
	m, session := newMockSession(t, "7.0.0")
	defer m.Close()

	attempts := 0
	m.handle("maintenance.get", func(req *Request) (interface{}, *APIError) {
		return []map[string]interface{}{
			{
				"maintenanceid": "10",
				"hosts":         []map[string]string{{"hostid": "15166", "host": "h1"}},
				"groups":        []interface{}{},
			},
		}, nil
	})
	m.handle("maintenance.delete", func(req *Request) (interface{}, *APIError) {
		return map[string][]string{"maintenanceids": {"10"}}, nil
	})
	m.handle("host.delete", func(req *Request) (interface{}, *APIError) {
		attempts++
		return nil, &APIError{Code: -32500, Message: "Application error.", Data: "delete failed"}
	})

	_, err := session.HostDelete(context.Background(), Host{HostID: "15166"})
	if err == nil || !strings.Contains(err.Error(), "delete failed") {
		t.Fatalf("ожидалась ошибка удаления, получено: %v", err)
	}
	if attempts != 1 {
		t.Errorf("ожидалась одна попытка host.delete, получено %d", attempts)
	}
}
