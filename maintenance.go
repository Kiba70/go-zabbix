package zabbix

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type MaintenanceType int
type TagsEvaltype int

var ErrMaintenanceHostNotFound = errors.New("Failed to find ID by host name")

const (
	withDataCollection MaintenanceType = iota
	withoutDataCollection

	and TagsEvaltype = iota * 2
	or

	Once int = iota
	EveryDay
	EveryWeek
	EveryMonth
)

type Maintenance struct {
	Session       *Session
	MaintenanceID string
	Name          string
	ActiveSince   time.Time
	ActiveTill    time.Time
	Description   string
	// Service period in seconds
	ServicePeriod       int
	Type                MaintenanceType
	ActionEvalTypeAndOr TagsEvaltype

	Hosts       []Host
	Hostgroups  []Hostgroup
	Tags        []MaintenanceTag
	Timeperiods []Timeperiod
}

type Maintenances []Maintenance

type MaintenanceTag struct {
	Tag      string
	Operator int
	Value    string
}

type Timeperiod struct {
	TimeperiodId int

	// Day of the month when the maintenance must come into effect.
	// Required only for monthly time periods.
	Day int

	// Days are stored in binary form with each bit representing the corresponding day.
	// For example, 4 equals 100 in binary and means, that maintenance will be enabled on Wednesday.
	Dayofweek int

	// For daily and weekly periods every defines day or week intervals at which the maintenance must come into effect.
	// For monthly periods every defines the week of the month when the maintenance must come into effect.
	// Possible values:
	// 1 - first week;
	// 2 - second week;
	// 3 - third week;
	// 4 - fourth week;
	// 5 - last week.
	Every int

	// Months when the maintenance must come into effect.
	// Months are stored in binary form with each bit representing the corresponding month.
	// For example, 5 equals 101 in binary and means, that maintenance will be enabled in January and March.
	// Required only for monthly time periods.
	Month int

	// Duration of the maintenance period in seconds.
	// Default: 3600.
	Period time.Duration

	// Date when the maintenance period must come into effect.
	// Required only for one time periods.
	// Default: current date.
	StartDate time.Time

	// Time of day when the maintenance starts in seconds.
	// Required for daily, weekly and monthly periods.
	StartTime int

	// Type of time period.
	// Possible values:
	// 0 - (default) one time only;
	// 2 - daily;
	// 3 - weekly;
	// 4 - monthly.
	TimeperiodType int
}

type MaintenanceGetParams struct {
	GetParameters

	// Sort the result by the given properties.
	// Possible values are: maintenanceid, name and maintenance_type.
	SortField []string `json:"sortfield,omitempty"`

	// Return the maintenance's time periods in the timeperiods property.
	SelectTimeperiods SelectQuery `json:"selectTimeperiods,omitempty"`

	// Return hosts assigned to the maintenance in the hosts property.
	SelectHosts SelectQuery `json:"selectHosts,omitempty"`

	// Return host groups assigned to the maintenance in the groups property.
	SelectGroups SelectQuery `json:"selectGroups,omitempty"`

	// Return tags assigned to the maintenance in the tags property.
	SelectTags SelectQuery `json:"selectTags,omitempty"`

	// Return only maintenances with the given IDs.
	Maintenanceids []string `json:"maintenanceids,omitempty"`

	// Return only maintenances that are assigned to the given hosts.
	Hostids []string `json:"hostids,omitempty"`

	// Return only maintenances that are assigned to the given host groups.
	Groupids []string `json:"groupids,omitempty"`
}

type MaintenanceCreateParams struct {
	Maintenance

	Groupids []string `json:"groupids,omitempty"`
	// Hosts name
	HostNames   []string         `json:"-"`
	HostIDs     []string         `json:"hostids"`
	Timeperiods []Timeperiod     `json:"timeperiods"`
	Tags        []MaintenanceTag `json:"tags,omitempty"`
}

type jMaintenanceCreateParams struct {
	JMaintenance

	Groupids []string `json:"groupids,omitempty"` // Только до версии 6.0
	// Hosts name
	HostNames   []string          `json:"-,omitempty"`       // Только до версии 6.0
	HostIDs     []string          `json:"hostids,omitempty"` // Только до версии 6.0
	Timeperiods []jTimeperiod     `json:"timeperiods,omitempty"`
	Tags        []jMaintenanceTag `json:"tags,omitempty"`

	// Добавлено для Zabbix 6.0
	Groups jHostgroups `json:"groups,omitempty"`
	Hosts  jHosts      `json:"hosts,omitempty"`
}

type MaintenanceCreateResponse struct {
	IDs []string `json:"maintenanceids"`
}

// GetMaintenance queries the Zabbix API for Maintenance matching the given search
// parameters.
func (s *Session) GetMaintenance(params *MaintenanceGetParams) ([]Maintenance, error) {
	jmaintenance := make([]jMaintenanceGet, 0)
	err := s.Get("maintenance.get", params, &jmaintenance)
	if err != nil {
		return nil, err
	}

	if len(jmaintenance) == 0 {
		return nil, ErrNotFound
	}

	out := make([]Maintenance, len(jmaintenance))
	for i, jaction := range jmaintenance {
		maintenance, err := jaction.Maintenance()
		if err != nil {
			return nil, fmt.Errorf("Error mapping Maintenance %d in response: %v", i, err)
		}

		out[i] = *maintenance
		out[i].Session = s
	}

	return out, nil
}

func (m *Maintenance) Create() (response MaintenanceCreateResponse, err error) {
	newM := &jMaintenanceCreateParams{}
	newM.Name = m.Name
	newM.ActiveSince = m.ActiveSince.Unix()
	newM.ActiveTill = m.ActiveTill.Unix()
	newM.Description = m.Description
	newM.MaintenanceType = int(m.Type)
	newM.TagsEvaltype = int(m.ActionEvalTypeAndOr)

	for _, tp := range m.Timeperiods {
		t := jTimeperiod{}
		switch tp.TimeperiodType {
		case 0:
			// t.TimeperiodId = 0
			// t.Day = 0
			// t.Dayofweek = 0
			// t.Every = 1 // Нельзя указывать 0
			// t.Month = 0
			t.Period = int(tp.Period.Seconds())
			t.StartDate = tp.StartDate.Unix()
			t.StartTime = tp.StartTime
			t.TimeperiodType = tp.TimeperiodType
		}
		newM.Timeperiods = append(newM.Timeperiods, t)
	}
	// newM.Tags = append(newM.Tags, jMaintenanceTag(params.Tags)...)
	for _, i := range m.Tags {
		t := jMaintenanceTag(i)
		newM.Tags = append(newM.Tags, t)
	}

	switch m.Session.ApiVersion.Major {
	case 4, 5:
		for _, host := range m.Hosts {
			newM.HostNames = append(newM.HostNames, host.Hostname)
		}
	case 6:
		for _, host := range m.Hosts {
			jHost := &jHost{}
			jHost.Hostname = host.Hostname
			newM.Hosts = append(newM.Hosts, *jHost)
		}
	}

	// Заполняем HostID на основе заполненных выше имён хостов
	if err = newM.FillHostIDs(m.Session); err != nil {
		return
	}

	err = m.Session.Get("maintenance.create", newM, &response)

	return
}

func (m *Maintenance) Delete() error {
	ID := []string{m.MaintenanceID}
	response := make(map[string]interface{})
	if err := m.Session.Get("maintenance.delete", ID, &response); err != nil {
		return err
	}
	return nil
}

func (m *Maintenance) Update() (response MaintenanceCreateResponse, err error) {
	newM := &jMaintenanceCreateParams{}
	if m.MaintenanceID == "" {
		// Нет ID
		// err = error.Printf("Error")
		return
	}
	newM.MaintenanceID = m.MaintenanceID
	if m.Name != "" {
		newM.Name = m.Name
	}
	// if m.ActiveSince !=  {
	// 	newM.ActiveSince = m.ActiveSince.Unix()
	// }
	// newM.ActiveTill = m.ActiveTill.Unix()
	if m.Description != "" {
		newM.Description = m.Description
	}
	if m.Type != 0 {
		newM.MaintenanceType = int(m.Type)
	}
	if m.ActionEvalTypeAndOr != 0 {
		newM.TagsEvaltype = int(m.ActionEvalTypeAndOr)
	}
	for _, tp := range m.Timeperiods {
		t := jTimeperiod{}
		switch tp.TimeperiodType {
		case 0:
			t.TimeperiodId = 0
			t.Day = 0
			t.Dayofweek = 0
			t.Every = 1 // Нельзя указывать 0
			t.Month = 0
			t.Period = int(tp.Period.Seconds())
			t.StartDate = tp.StartDate.Unix()
			t.StartTime = tp.StartTime
			t.TimeperiodType = tp.TimeperiodType
		}
		newM.Timeperiods = append(newM.Timeperiods, t)
	}
	// newM.Tags = append(newM.Tags, jMaintenanceTag(params.Tags)...)
	for _, i := range m.Tags {
		t := jMaintenanceTag(i)
		newM.Tags = append(newM.Tags, t)
	}

	switch m.Session.ApiVersion.Major {
	case 4, 5:
		for _, host := range m.Hosts {
			newM.HostNames = append(newM.HostNames, host.Hostname)
		}
	case 6:
		for _, host := range m.Hosts {
			jHost := &jHost{}
			jHost.Hostname = host.Hostname
			newM.Hosts = append(newM.Hosts, *jHost)
		}
	}

	if err = newM.FillHostIDs(m.Session); err != nil {
		return
	}

	err = m.Session.Get("maintenance.update", newM, &response)

	return
}

func (m *jMaintenanceCreateParams) FillHostIDs(session *Session) error {
	hgp := &HostGetParams{}
	// switch session.ApiVersion.Major {
	// case 4, 5:
	// 	for _, hostname := range m.HostNames {
	// 		hgp.GetParameters.Filter["host"] = string(hostname)
	// 	}
	// case 6:
	// 	for _, host := range m.Hosts {
	// 		hgp.GetParameters.Filter["host"] = string(host.Hostname)
	// 	}
	// }

	hosts, err := session.GetHosts(*hgp)
	if err != nil {
		return err
	}

	err = ErrMaintenanceHostNotFound
	switch session.ApiVersion.Major {
	case 4, 5:
		for _, name := range m.HostNames {
			for _, host := range hosts {
				// if strings.ToUpper(strings.Trim(host.Hostname, " ")) == strings.ToUpper(strings.Trim(name, " ")) {
				if strings.EqualFold(strings.Trim(host.Hostname, " "), strings.Trim(name, " ")) {
					m.HostIDs = append(m.HostIDs, host.HostID)

					err = nil
				}
			}
		}
	case 6:
		for i, mHost := range m.Hosts {
			for _, host := range hosts {
				if strings.EqualFold(strings.Trim(host.Hostname, " "), strings.Trim(mHost.Hostname, " ")) {
					m.Hosts[i].HostID = host.HostID
					m.Hosts[i].Hostname = "" // Очищаем т.к. API Zabbix ругается на заполненное поле

					err = nil
				}
			}
		}
	}

	return err
}

func (c *jMaintenanceCreateParams) FillFields(Object *Maintenance) *jMaintenanceCreateParams {
	c.ActiveSince = Object.ActiveSince.Unix()
	c.ActiveTill = Object.ActiveSince.Add(time.Hour * time.Duration(Object.ServicePeriod)).Unix()
	c.Description = Object.Description
	c.MaintenanceID = Object.MaintenanceID
	c.Name = Object.Name
	c.TagsEvaltype = int(Object.Type)
	c.MaintenanceType = int(Object.ActionEvalTypeAndOr)

	return c
}
