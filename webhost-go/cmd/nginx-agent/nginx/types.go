package nginx

type AgentInfo struct {
	Username string `json:"username"`
	Hostname string `json:"hostname"`
	VMIP     string `json:"VMIP"`
	SSHPort  int    `json:"SSHPort"`
}
