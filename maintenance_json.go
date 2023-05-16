package zabbix

import (
	"fmt"
	"time"
)

// JMaintenance is a private map for the Zabbix API Maintenance object.
// See: https://www.zabbix.com/documentation/2.2/manual/api/reference/maintenance/object
type JMaintenance struct {
	MaintenanceID   string `json:"maintenanceid"`
	Name            string `json:"name,omitempty"`
	ActiveSince     int64  `json:"active_since,string,omitempty"`
	ActiveTill      int64  `json:"active_till,string,omitempty"`
	Description     string `json:"description,omitempty"`
	MaintenanceType int    `json:"maintenance_type,string,omitempty"`
	TagsEvaltype    int    `json:"tags_evaltype,string,omitempty"`
}

type jMaintenanceGet struct {
	JMaintenance

	Hosts       jHosts           `json:"hosts"`
	Hostgroups  jHostgroups      `json:"groups"`
	Tags        jMaintenanceTags `json:"tags"`
	Timeperiods jTimeperiods     `json:"timeperiods"`
}

type jMaintenanceTag struct {
	Tag      string `json:"tag"`
	Operator int    `json:"operator,string"`
	Value    string `json:"value"`
}

type jMaintenanceTags []jMaintenanceTag

// Timeperiods is a private map for the Zabbix API Maintenance object.
// See: https://www.zabbix.com/documentation/2.2/manual/api/reference/maintenance/object
type jTimeperiod struct {
	TimeperiodId   int   `json:"timeperiodid,string"`
	Day            int   `json:"day,string"`
	Dayofweek      int   `json:"dayofweek,string"`
	Every          int   `json:"every,string"`
	Month          int   `json:"month,string"`
	Period         int   `json:"period,string"`
	StartDate      int64 `json:"start_date,string"`
	StartTime      int   `json:"start_time,string"`
	TimeperiodType int   `json:"timeperiod_type,string"`
}

type jTimeperiods []jTimeperiod

// Maintenance returns a native Go Maintenance struct mapped from the given JSON Maintenance
// data.
func (c *jMaintenanceGet) Maintenance() (result *Maintenance, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	maintenance := &Maintenance{
		ActionEvalTypeAndOr: TagsEvaltype(c.MaintenanceType),
		Type:                MaintenanceType(c.TagsEvaltype),
		ActiveSince:         time.Unix(c.ActiveSince, 0),
		ActiveTill:          time.Unix(c.ActiveTill, 0),
		Description:         c.Description,
		MaintenanceID:       c.MaintenanceID,
		Name:                c.Name,
	}

	// Hosts:               c.Hosts.Hosts(),
	if len(c.Hosts) > 0 {
		if hosts, err := c.Hosts.Hosts(); err == nil {
			maintenance.Hosts = hosts
		}
	}

	// Hostgroups:          c.Hostgroups,
	if len(c.Hostgroups) > 0 {
		if hostgroups, err := c.Hostgroups.Hostgroups(); err == nil {
			maintenance.Hostgroups = hostgroups
		}
	}

	// Tags:                c.Tags,
	if len(c.Tags) > 0 {
		if tags, err := c.Tags.Tags(); err == nil {
			maintenance.Tags = tags
		}
	}

	// Timeperiods:                c.Timeperiods,
	if len(c.Timeperiods) > 0 {
		if timeperiods, err := c.Timeperiods.Timeperiods(); err == nil {
			maintenance.Timeperiods = timeperiods
		}
	}

	return maintenance, nil
}

// Tag returns a native Go Tag struct mapped from the given JSON Tag data.
func (c *jMaintenanceTag) jTag() (*MaintenanceTag, error) {
	//var err error

	tag := &MaintenanceTag{}
	tag.Tag = c.Tag
	tag.Operator = c.Operator
	tag.Value = c.Value

	return tag, nil
}

// Hosts returns a native Go slice of Hosts mapped from the given JSON Hosts
// data.
func (c jMaintenanceTags) Tags() ([]MaintenanceTag, error) {
	if c != nil {
		tags := make([]MaintenanceTag, len(c))
		for i, jtag := range c {
			tag, err := jtag.jTag()
			if err != nil {
				return nil, fmt.Errorf("Error unmarshalling Tags %d in JSON data: %v", i, err)
			}

			tags[i] = *tag
		}

		return tags, nil
	}

	return nil, nil
}

// Tag returns a native Go Tag struct mapped from the given JSON Tag data.
func (c *jTimeperiod) jTimeperiod() (*Timeperiod, error) {
	//var err error

	t := &Timeperiod{}
	t.TimeperiodId = c.TimeperiodId
	t.Day = c.Day
	t.Dayofweek = c.Dayofweek
	t.Every = c.Every
	t.Month = c.Month
	t.Period = time.Duration(c.Period * 1000)
	t.StartDate = time.Unix(c.StartDate, 0)
	t.StartTime = c.StartTime
	t.TimeperiodType = c.TimeperiodType

	return t, nil
}

// Hosts returns a native Go slice of Hosts mapped from the given JSON Hosts
// data.
func (c jTimeperiods) Timeperiods() ([]Timeperiod, error) {
	if c != nil {
		timeperiods := make([]Timeperiod, len(c))
		for i, jtimeperiod := range c {
			timeperiod, err := jtimeperiod.jTimeperiod()
			if err != nil {
				return nil, fmt.Errorf("Error unmarshalling Tags %d in JSON data: %v", i, err)
			}

			timeperiods[i] = *timeperiod
		}

		return timeperiods, nil
	}

	return nil, nil
}
