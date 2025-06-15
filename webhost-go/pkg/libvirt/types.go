package libvirt

type VMConfig struct {
	Name     string
	MemoryMB int
	VCPUs    int
	DiskPath string
	ISOPath  string
}

type VMInfo struct {
	Name   string
	State  string
	Memory uint64
	VCPU   uint
	UUID   string
}
