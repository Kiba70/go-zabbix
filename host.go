package zabbix

const (
	// HostSourceDefault indicates that a Host was created in the normal way.
	HostSourceDefault = 0

	// HostSourceDiscovery indicates that a Host was created by Host discovery.
	HostSourceDiscovery = 4

	// HostAvailabilityUnknown Unknown availability of host, never has come online
	HostAvailabilityUnknown = 0

	// HostAvailabilityAvailable Host is available
	HostAvailabilityAvailable = 1

	// HostAvailabilityUnavailable Host is NOT available
	HostAvailabilityUnavailable = 2

	// HostInventoryModeDisabled Host inventory in disabled
	HostInventoryModeDisabled = -1

	// HostInventoryModeManual Host inventory is managed manually
	HostInventoryModeManual = 0

	// HostInventoryModeAutomatic Host inventory is managed automatically
	HostInventoryModeAutomatic = 1

	// HostTLSConnectUnencryped connect unencrypted to or from host
	HostTLSConnectUnencryped = 1

	// HostTLSConnectPSK connect with PSK to or from host
	HostTLSConnectPSK = 2

	// HostTLSConnectCertificate connect with certificate to or from host
	HostTLSConnectCertificate = 4

	// HostStatusMonitored Host is monitored
	HostStatusMonitored = 0

	// HostStatusUnmonitored Host is not monitored
	HostStatusUnmonitored = 1
)

// Host represents a Zabbix Host returned from the Zabbix API.
//
// See: https://www.zabbix.com/documentation/2.2/manual/config/hosts
type Host struct {
	// HostID is the unique ID of the Host.
	HostID string `json:"hostid,omitempty"`

	// Hostname is the technical name of the Host.
	Hostname string `json:"host,omitempty"`

	// DisplayName is the visible name of the Host.
	DisplayName string `json:"name,omitempty"`

	// Source is the origin of the Host and must be one of the HostSource
	// constants.
	Source int `json:"flags,string,omitempty"`

	// Статус хоста
	Status int `json:"status,string,omitempty"`

	// Macros contains all Host Macros assigned to the Host.
	Macros []HostMacro `json:"macros,omitempty"`

	// Macros contains all Host Macros assigned to the Host.
	Tags []HostTag `json:"tags,omitempty"`

	// Groups contains all Host Groups assigned to the Host.
	Groups []Hostgroup `json:"groups,omitempty"`

	// Interfaces of the host
	Interfaces []HostInterface `json:"interfaces,omitempty"`

	Templates []Template `json:"templates,omitempty"`

	TemplatesClear []Template `json:"templates_clear,omitempty"`

	Proxy             string    `json:"proxy_hostid,omitempty"`
	MaintenanceStatus string    `json:"maintenance_status,omitempty"`
	MaintenanceID     string    `json:"maintenanceid,omitempty"`
	MaintenanceType   string    `json:"maintenance_type,omitempty"`
	MaintenanceFrom   string    `json:"maintenance_from,omitempty"`
	InventoryMode     int       `json:"inventory_mode,string,omitempty"`
	Inventory         inventory `json:"inventory,omitempty"`

	// Description of host
	Description string `json:"description,omitempty"`

	// How should we connect to host
	TLSConnect int `json:"tls_connect,string,omitempty"`

	// What type of connections we accept from host
	TLSAccept int `json:"tls_accept,string,omitempty"`

	TLSIssuer      string `json:"tls_issuer,omitempty"`
	TLSSubject     string `json:"tls_subject,omitempty"`
	TLSPSKIdentity string `json:"tls_psk_identity,omitempty"`
	TLSPSK         string `json:"tls_psk,omitempty"`
}

type inventory struct {
	DateHwInstall  string `json:"date_hw_install,omitempty"`
	DateHwPurchase string `json:"date_hw_purchase,omitempty"`
}

// HostGetParams represent the parameters for a `host.get` API call.
//
// See: https://www.zabbix.com/documentation/2.2/manual/api/reference/host/get#parameters
type HostGetParams struct {
	GetParameters

	// GroupIDs filters search results to hosts that are members of the given
	// Group IDs.
	GroupIDs []string `json:"groupids,omitempty"`

	// ApplicationIDs filters search results to hosts that have items in the
	// given Application IDs.
	ApplicationIDs []string `json:"applicationids,omitempty"`

	// DiscoveredServiceIDs filters search results to hosts that are related to
	// the given discovered service IDs.
	DiscoveredServiceIDs []string `json:"dserviceids,omitempty"`

	// GraphIDs filters search results to hosts that have the given graph IDs.
	GraphIDs []string `json:"graphids,omitempty"`

	// HostIDs filters search results to hosts that matched the given Host IDs.
	HostIDs []string `json:"hostids,omitempty"`

	// WebCheckIDs filters search results to hosts with the given Web Check IDs.
	WebCheckIDs []string `json:"httptestids,omitempty"`

	// InterfaceIDs filters search results to hosts that use the given Interface
	// IDs.
	InterfaceIDs []string `json:"interfaceids,omitempty"`

	// ItemIDs filters search results to hosts with the given Item IDs.
	ItemIDs []string `json:"itemids,omitempty"`

	// MaintenanceIDs filters search results to hosts that are affected by the
	// given Maintenance IDs
	MaintenanceIDs []string `json:"maintenanceids,omitempty"`

	// MonitoredOnly filters search results to return only monitored hosts.
	MonitoredOnly bool `json:"monitored_hosts,omitempty"`

	// ProxyOnly filters search results to hosts which are Zabbix proxies.
	ProxiesOnly bool `json:"proxy_host,omitempty"`

	// ProxyIDs filters search results to hosts monitored by the given Proxy
	// IDs.
	ProxyIDs []string `json:"proxyids,omitempty"`

	// IncludeTemplates extends search results to include Templates.
	IncludeTemplates bool `json:"templated_hosts,omitempty"`

	// SelectGroups causes the Host Groups that each Host belongs to to be
	// attached in the search results.
	SelectGroups SelectQuery `json:"selectGroups,omitempty"`

	// SelectApplications causes the Applications from each Host to be attached
	// in the search results.
	SelectApplications SelectQuery `json:"selectApplications,omitempty"`

	// SelectDiscoveries causes the Low-Level Discoveries from each Host to be
	// attached in the search results.
	SelectDiscoveries SelectQuery `json:"selectDiscoveries,omitempty"`

	// SelectDiscoveryRule causes the Low-Level Discovery Rule that created each
	// Host to be attached in the search results.
	SelectDiscoveryRule SelectQuery `json:"selectDiscoveryRule,omitempty"`

	// SelectGraphs causes the Graphs from each Host to be attached in the
	// search results.
	SelectGraphs SelectQuery `json:"selectGraphs,omitempty"`

	SelectHostDiscovery SelectQuery `json:"selectHostDiscovery,omitempty"`

	SelectWebScenarios SelectQuery `json:"selectHttpTests,omitempty"`

	SelectInterfaces SelectQuery `json:"selectInterfaces,omitempty"`

	SelectInventory SelectQuery `json:"selectInventory,omitempty"`

	SelectItems SelectQuery `json:"selectItems,omitempty"`

	SelectMacros SelectQuery `json:"selectMacros,omitempty"`

	SelectParentTemplates SelectQuery `json:"selectParentTemplates,omitempty"`
	SelectScreens         SelectQuery `json:"selectScreens,omitempty"`
	SelectTriggers        SelectQuery `json:"selectTriggers,omitempty"`
}

type HostTag struct {
	Tag   string `json:"tag,omitempty"`
	Value string `json:"value,omitempty"`
}

type HostResponse struct {
	IDs []string `json:"hostids"`
}

// GetHosts queries the Zabbix API for Hosts matching the given search
// parameters.
//
// ErrEventNotFound is returned if the search result set is empty.
// An error is returned if a transport, parsing or API error occurs.
func (s *Session) GetHosts(params HostGetParams) ([]Host, error) {
	hosts := make([]Host, 0)
	err := s.Get("host.get", params, &hosts)
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return nil, ErrNotFound
	}

	return hosts, nil
}

func (s *Session) HostCreate(hosts ...Host) (response []string, err error) {
	var hcr HostResponse
	// var err2 error

	response = make([]string, 0, len(hosts))

	for _, h := range hosts {
		err2 := s.Get("host.create", h, &hcr)
		if err2 != nil {
			err = err2
		}
		response = append(response, hcr.IDs...)
	}

	return
}

/* func (c *Session) HostUpdate(hosts ...Host) (response []string, err error) {
	var hcr HostResponse
	// var err2 error

	response = make([]string, 0, len(hosts))

	for _, h := range hosts {
		err2 := c.Get("host.update", h, &hcr)
		if err2 != nil {
			err = err2
		}
		response = append(response, hcr.IDs...)
	}

	return
} */

func (s *Session) HostUpdate(hosts ...Host) (response []string, err error) {
	var hcr HostResponse

	response = make([]string, 0, len(hosts))

	err = s.Get("host.update", hosts, &hcr)

	response = append(response, hcr.IDs...)

	return
}

func (s *Session) HostDelete(hosts ...Host) (response []string, err error) {
	var hcr HostResponse

	response = make([]string, 0, len(hosts))
	toDelete := make([]string, 0, len(hosts))

	for _, h := range hosts {
		toDelete = append(toDelete, h.HostID)
	}

	err = s.Get("host.delete", toDelete, &hcr)

	response = append(response, hcr.IDs...)

	if len(toDelete) == len(response) {
		return // Всё удалено - выходим
	}

	// Получаем список HostID которые не удалось удалить
	for _, r := range response {
		for i, d := range toDelete {
			if r == d {
				toDelete = removeSliceIndex(toDelete, i)
				break
			}
		}
	}

	// Получаем список Maintenance в которых участвуют не удалённые хосты
	mgp := MaintenanceGetParams{
		SelectHosts: SelectExtendedOutput,
	}
	mgp.Hostids = make([]string, 0, len(toDelete))
	mgp.Hostids = append(mgp.Hostids, toDelete...)
	maintenances, err := s.GetMaintenance(&mgp)
	if err != nil {
		return // Ошибка или ничего не найдено - выходим
	}

	// Удаляем хосты из maintenance
	tryDelete := make([]string, 0, len(toDelete))
	for _, m := range maintenances {
		allHosts := true
		hosts2 := make([]string, 0)
		for _, h := range m.Hosts {
			if !have(toDelete, h.HostID) {
				allHosts = false
			} else {
				hosts2 = append(hosts2, h.HostID)
			}
		}
		if allHosts {
			m.Delete() // Удаляем весь maintenance
			tryDelete = append(tryDelete, hosts2...)
		} else {
			// В maintenance только часть хостов
		}
	}

	// Вторая попытка удаления
	if len(tryDelete) > 0 {
		err = s.Get("host.delete", tryDelete, &hcr)
		response = append(response, hcr.IDs...)
	}

	return
}

func have(s []string, i string) bool {
	for _, ss := range s {
		if ss == i {
			return true
		}
	}

	return false
}

func removeSliceIndex(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
