package zabbix

import "context"

const (
	// HostInterfaceAvailabilityUnknown Unknown availability of host, never has come online
	HostInterfaceAvailabilityUnknown = 0
	// HostInterfaceAvailabilityAvailable Host is available
	HostInterfaceAvailabilityAvailable = 1
	// HostInterfaceAvailabilityUnavailable Host is NOT available
	HostInterfaceAvailabilityUnavailable = 2

	// HostInterfaceTypeAgent Host interface type agent
	HostInterfaceTypeAgent = 1
	// HostInterfaceTypeSNMP Host interface type SNMP
	HostInterfaceTypeSNMP = 2
	// HostInterfaceTypeIPMI Host interface type IPMI
	HostInterfaceTypeIPMI = 3
	// HostInterfaceTypeJMX Host interface type JMX
	HostInterfaceTypeJMX = 4
)

// HostInterface  This class is designed to work with host interfaces.
//
// See https://www.zabbix.com/documentation/current/manual/api/reference/hostinterface/object#host_interface
type HostInterface struct {
	// InterfaceID is the unique ID of the Interface. (readonly)
	InterfaceID string `json:"interfaceid,omitempty"`

	// (readonly) Availability of host interface.
	Available int `json:"available,string,omitempty"`

	// DNS name used by the interface.
	DNS string `json:"dns"`

	// IP address used by the interface.
	IP string `json:"ip,omitempty"`

	// (readonly) Error text if host interface is unavailable.
	Error string `json:"error,omitempty"`

	// (readonly) Time when host interface became unavailable.
	ErrorsFrom *UnixTimestamp `json:"errors_from,string,omitempty"`

	// HostID - ID of the host the interface belongs to.
	HostID string `json:"hostid,omitempty"`

	// Whether the interface is used as default on the host. Only one interface of some type can be set as default on a host.
	// Possible values are:
	// 0 - not default;
	// 1 - default.
	Main ZBXBoolean `json:"main,string,omitempty"`

	// Port number used by the interface.
	Port string `json:"port,omitempty"`

	// Interface type.
	// Possible values are: 1 - agent; 2 - SNMP; 3 - IPMI; 4 - JMX.
	Type int `json:"type,string,omitempty"`

	// Whether the connection should be made via IP.
	// Possible values are:
	// 0 - connect using host DNS name;
	// 1 - connect using host IP address for this host interface.
	Useip ZBXBoolean `json:"useip,string,omitempty"`

	Details *HostInterfaceDetail `json:"details,omitempty"`
}

type HostInterfaceDetail struct {
	// SNMP interface version.
	// Possible values are: 1 - SNMPv1; 2 - SNMPv2c; 3 - SNMPv3
	Version string `json:"version,omitempty"`

	// Whether to use bulk SNMP requests.
	Bulk string `json:"bulk,omitempty"`

	// SNMP community (required). Used only by SNMPv1 and SNMPv2 interfaces.
	Community string `json:"community,omitempty"`

	// SNMPv3 security name.
	Securityname string `json:"securityname,omitempty"`

	// SNMPv3 security level. Used only by SNMPv3 interfaces.
	// Possible values are: 0 - (default) - noAuthNoPriv; 1 - authNoPriv; 2 - authPriv.
	Securitylevel string `json:"securitylevel,omitempty"`

	// SNMPv3 authentication passphrase.
	Authpassphrase string `json:"authpassphrase,omitempty"`

	// SNMPv3 privacy passphrase.
	Privpassphrase string `json:"privpassphrase,omitempty"`

	// SNMPv3 authentication protocol. Used only by SNMPv3 interfaces.
	// Possible values are: 0 - (default) - MD5; 1 - SHA1; 2 - SHA224; 3 - SHA256; 4 - SHA384; 5 - SHA512.
	Authprotocol string `json:"authprotocol,omitempty"`

	// SNMPv3 privacy protocol. Used only by SNMPv3 interfaces.
	// Possible values are: 0 - (default) - DES; 1 - AES128; 2 - AES192; 3 - AES256; 4 - AES192C; 5 - AES256C.
	Privprotocol string `json:"privprotocol,omitempty"`

	// SNMPv3 context name.
	Contextname string `json:"contextname,omitempty"`
}

type HostInterfaceGetParams struct {
	GetParameters

	// Return only host interfaces used by the given hosts.
	HostIDs []string `json:"hostids,omitempty"`

	// Return only host interfaces with the given IDs.
	InterfaceIDs []string `json:"interfaceids,omitempty"`

	// Return only host interfaces used by the given items.
	ItemIDs []string `json:"itemids,omitempty"`

	// Return only host interfaces used by items in the given triggers.
	TriggerIDs []string `json:"triggerids,omitempty"`
}

// GetHostInterfaces queries the Zabbix API for Hosts interfaces matching the given search
// parameters.
//
// ErrEventNotFound is returned if the search result set is empty.
// An error is returned if a transport, parsing or API error occurs.
func (c *Session) GetHostInterfaces(ctx context.Context, params HostInterfaceGetParams) ([]HostInterface, error) {
	hostInterfaces := make([]HostInterface, 0)
	err := c.Get(ctx, "hostinterface.get", params, &hostInterfaces)
	if err != nil {
		return nil, err
	}

	if len(hostInterfaces) == 0 {
		return nil, ErrNotFound
	}

	return hostInterfaces, nil
}
