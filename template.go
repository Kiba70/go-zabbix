package zabbix

type Template struct {
	// (readonly) ID of the template
	TemplateID string `json:"templateid,omitempty"`

	// Technical name of the template
	Host string `json:"host,string,omitempty"`

	// Description of the template
	Description string `json:"description,string,omitempty"`

	// Visible name of the template
	Name string `json:"name,string,omitempty"`

	// Universal unique identifier, used for linking imported templates to already existing ones. Auto-generated, if not given.
	UUID string `json:"uuid,string,omitempty"`
}
