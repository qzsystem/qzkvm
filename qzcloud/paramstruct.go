package qzcloud

type CreateStruct struct {
	HostName string `json:"host_name" binding:"required"`
	HostUuid string `json:"host_uuid" `
	MaxRam string `json:"max_ram" binding:"required"`
	MinRam string `json:"min_ram" binding:"required"`
	Cpu string `json:"cpu" binding:"required"`
	CpuModel string `json:"cpu_model" binding:"required"`
	OsType string `json:"os_type" binding:"required"`
	Bandwidth string `json:"bandwidth" binding:"required"`
	Otherip string `json:"otherip" `
	IP string `json:"ip" binding:"required"`
	Gateway string `json:"gateway" binding:"required"`
	Netmask string `json:"netmask" binding:"required"`
	DNS1 string `json:"dns1" binding:"required"`
	DNS2 string `json:"dns2" binding:"required"`
	MAC string `json:"mac" binding:"required"`
	ReCreate string `json:"re_create" `

	IP1 string `json:"ip1" binding:"required"`
	Gateway1 string `json:"gateway1" `
	Netmask1 string `json:"netmask1" binding:"required"`
	MAC1 string `json:"mac1" binding:"required"`
	TemplatePath string `json:"template_path" binding:"required"`
	DataPath string `json:"data_path" binding:"required"`
	OsName  string `json:"os_name" binding:"required"`
	HostData string `json:"host_data" binding:"required"`

	DataRead  string `json:"data_read" binding:"required"`
	DataWrite  string `json:"data_write" binding:"required"`
	OsRead  string `json:"os_read" binding:"required"`
	OsWrite  string `json:"os_write" binding:"required"`
	DataIops  string `json:"data_iops" binding:"required"`
	OsIops  string `json:"os_iops" binding:"required"`

	VncPort  string `json:"vnc_port" binding:"required"`
	VncPassword  string `json:"vnc_password" binding:"required"`
	Password string `json:"password" binding:"required"`

}

type RemoveStruct struct {
	HostName string `json:"host_name" binding:"required"`
	DataPath string `json:"data_path" binding:"required"`
	BackupPath string `json:"backup_path" binding:"required"`
}

type ReinstallStruct struct {
	HostName string `json:"host_name" binding:"required"`
	OsType string `json:"os_type" binding:"required"`
	IP string `json:"ip" binding:"required"`
	Gateway string `json:"gateway" binding:"required"`
	Netmask string `json:"netmask" binding:"required"`
	DNS1 string `json:"dns1" binding:"required"`
	DNS2 string `json:"dns2" binding:"required"`
	MAC string `json:"mac" binding:"required"`

	IP1 string `json:"ip1" binding:"required"`
	Gateway1 string `json:"gateway1"`
	Netmask1 string `json:"netmask1" binding:"required"`
	MAC1 string `json:"mac1" binding:"required"`
	Password string `json:"password" binding:"required"`

	TemplatePath string `json:"template_path" binding:"required"`
	DataPath string `json:"data_path" binding:"required"`
	OsName  string `json:"os_name" binding:"required"`

}

type UpdateSystemPasswordStruct struct {
	HostName string `json:"host_name" binding:"required"`
	Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type CreateSnapshotStruct struct {
	HostName string `json:"host_name" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type RestoreSnapshotStruct struct {
	HostName string `json:"host_name" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type RemoveSnapshotStruct struct {
	HostName string `json:"host_name" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type UpdateIPStruct struct {
	HostName string `json:"host_name" binding:"required"`
	Otherip string `json:"otherip" `
	IP string `json:"ip" binding:"required"`
}

type RemoveNWFilterStruct struct {
	HostName string `json:"host_name" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type AddNWFilterStruct struct {
	HostName string `json:"host_name" binding:"required"`
	Name string `json:"name" binding:"required"`
	Protocol string `json:"protocol" binding:"required"`
	Action string `json:"action" binding:"required"`
	Direction string `json:"direction" binding:"required"`
	Priority string `json:"priority" binding:"required"`
	Port string `json:"port" binding:"required"`
	StartIp string `json:"start_ip" binding:"required"`
	EndIp string `json:"end_ip"`
}

type UpdateIsoStruct struct {
	HostName string `json:"host_name" binding:"required"`
	IsoPath string `json:"iso_path" `
}

type BackupStruct struct {
	HostName string `json:"host_name" binding:"required"`
	BackupPath string `json:"backup_path" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type DomainMemoryStats struct{
	Actual uint64 //是启动虚机时设置的最大内存
	Swap_out uint64
	Swap_in uint64
	Major_fault uint64
	Minor_fault uint64
	Unused uint64 //虚拟机未被使用内存
	Available uint64 //虚拟机内存
	Last_update uint64
	Rss uint64 //在宿主机上所占用的内存
}

type MonitorStruct struct {
	HostName string `json:"host_name" binding:"required"`
	NetworkName string `json:"network_name" binding:"required"`
	DiskDev string `json:"disk_dev" `
}

type BootOrderStruct struct {
	HostName string `json:"host_name" binding:"required"`
	NetworkName string `json:"network_name" binding:"required"`
	DiskDev string `json:"disk_dev" `
}

type HostNameStruct struct {
	HostName string `json:"host_name" binding:"required"`
}

type SetStatusStruct struct {
	HostName string `json:"host_name" binding:"required"`
	State string `json:"state" binding:"required"`
}

type GetIsoListStruct struct {
	IsoPath string `json:"iso_path" binding:"required"`
}


