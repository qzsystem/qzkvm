package qzcloud
import "C"
import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/libvirt/libvirt-go"
	"github.com/lmittmann/ppm"
	"github.com/subchen/go-xmldom"
	"html/template"
	"image/jpeg"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ArchStruct struct {
	Arch string `xml:"host>cpu>arch"`
}

//创建硬盘
func CreateVolume(path ,capacity,name string)(string,error){
	/**
	xmlTemplate :=GetCreateVolumeXml()
	t:=template.New("getCreateVolumeXml")
	t,_=t.Parse(xmlTemplate)
	var buf bytes.Buffer
	data := make(map[string]string)
	data["capacity"]=capacity
	data["path"]=path
	data["name"]=name
	t.Execute(&buf,data)
	xmlconf := buf.String()
	pool, err := Conn.LookupStoragePoolByTargetPath(path)
	if err!=nil{
		return "",err
	}

	defer pool.Free()
	vol,err:=pool.StorageVolCreateXML(xmlconf,0)
	fmt.Println(err)
	if err!=nil{
		return "",err
	}
	defer vol.Free()
	vol_path,_:= vol.GetPath()
	return vol_path,nil
	**/

	cmdstr1:=fmt.Sprintf(`qemu-img create %s/%s -f qcow2 %sG`,path,name,capacity)
	cmds := exec.Command("bash", "-c", cmdstr1)
	err:=cmds.Run()
	if err!=nil{
		return  "",err
	}
    return path+"/"+name,nil
}

//創建一個存儲池
func CreateStoragePool(name ,path string) error {
	err:=os.MkdirAll(path,0666)
	 if err!=nil{
	 	return err
	 }
	xmlTemplate :=GetCreatePoolXml()
	t:=template.New("getCreatePoolXml")
	t,_=t.Parse(xmlTemplate)
	var buf bytes.Buffer
	data := make(map[string]string)
	data["name"]=name
	data["path"]=path
	t.Execute(&buf,data)

	//fmt.Print(xmlconf)

	xmlconf := buf.String()
	storagePool,err:=Conn.StoragePoolDefineXML(xmlconf,0)
	if err!=nil{
		return  err
	}
	err=storagePool.Create(libvirt.STORAGE_POOL_CREATE_NORMAL)
	if err!=nil{
		return  err
	}
	storagePool.Refresh(0)
	err=storagePool.SetAutostart(true)
	if err!=nil{
		return  err
	}

	return nil

}

//存储池启动删除停止
func ManageStoragePool(name ,command string)error  {
	pool, err := Conn.LookupStoragePoolByName(name)
	if err!=nil{
		return err
	}

	if command=="start"{
		err =pool.Create(libvirt.STORAGE_POOL_CREATE_NORMAL)
	}else if command=="stop"{
		err=pool.Destroy()
	}else if command=="remove"{
		err=pool.Destroy()
		err=pool.Delete(libvirt.STORAGE_POOL_DELETE_NORMAL)
	}else if command=="undefine"{
		bool,err:=pool.IsActive()
		if err!=nil{
			return  err
		}
		if bool==true{
			err=pool.Destroy()
		}

		err=pool.Undefine()
		pool.Free()
	}

	return err
}

//设置存储池是否开机启动 true随机启动  false不随机启动
func SetPoolAutoStart(name string,auto bool)error{
	pool, err := Conn.LookupStoragePoolByName(name)
	if err!=nil{
		return  err
	}
	err=pool.SetAutostart(auto)
	return  err
}

func CreateDomain(data map[string]string)(map[string]string,error) {
	var Arch ArchStruct
	xmlcontent,_:=Conn.GetCapabilities()
	xml.Unmarshal([]byte(xmlcontent),&Arch)
	var buf bytes.Buffer
	xml_string := GetCreateDomainXml()

	if is_forward,exit:=data["is_forward"];is_forward=="forward"&&exit==true{
		reg := regexp.MustCompile(`<filterref[\S\s]+</filterref>`)
		xml_string = reg.ReplaceAllString(xml_string,"")
	}
	data["arch"] = Arch.Arch

	t := template.New("create_vm_xml")
	t = t.Funcs(template.FuncMap{"unescaped": unescaped})
	t.Parse(xml_string)
	t.Execute(&buf, data)
	dom,err:=Conn.DomainDefineXML(buf.String())
	if err!=nil{
		return data,err
	}

	err=dom.SetAutostart(true)
	if err!=nil{
		return data,err
	}
	return data,err

}

//删除全部策略
func UnNWFilter(name string){
	Filter,err:=Conn.LookupNWFilterByName(name)
	if err==nil{
		err=Filter.Undefine()
	}
}

//初始化一个策略
func InitNWFilter(name string)error{
	NW,err:=Conn.LookupNWFilterByName(name)
	if err==nil{
		NW.Undefine()
	}
	xml:="<filter name='"+name+"_public_firewall' chain='root'><filterref filter='clean-traffic' /></filter>"

	_,err=Conn.NWFilterDefineXML(xml)
	return  err
}
//添加一个防火墙策略
//host_name,name,protocol,action,direction,priority,port,start_ip,end_ip  string
func AddNWFilterRule(param map[string]string)error {
	protocol:= strings.ToLower(param["protocol"])
	name:= strings.ToLower(param["name"])
	action:= strings.ToLower(param["action"])
	direction:= strings.ToLower(param["direction"])
	port:= strings.ToLower(param["port"])
	start_ip:= strings.ToLower(param["start_ip"])
	end_ip:= strings.ToLower(param["end_ip"])
	priority:=strings.ToLower(param["priority"])

	NW,err:=Conn.LookupNWFilterByName(param["host_name"]+"_public_firewall")
	if err!=nil{
		return err
	}
	xml,err:=NW.GetXMLDesc(0)
	if err!=nil{
		return err
	}
	root:=xmldom.Must(xmldom.ParseXML(xml)).Root
	add_rule:=root.CreateNode("rule")
	add_rule.SetAttributeValue("action",action).SetAttributeValue("direction",direction).SetAttributeValue("priority",priority)
	if strings.ToLower(protocol) == "any"{
		protocol ="all"
	}
	add_protocol:=add_rule.CreateNode(protocol).SetAttributeValue("comment",param["host_name"]+"-"+name)
	port=strings.ToLower(port)
	if port!="-1"&&port!="any"&&port!=""{
		add_protocol.SetAttributeValue("dstportstart",port)
	}

	if start_ip!="0.0.0.0"&&start_ip!=""&&start_ip!="any"{
		add_protocol.SetAttributeValue("srcipfrom",start_ip)
	}

	if end_ip!="0.0.0.0"&&end_ip!=""&&end_ip!="any"{
		add_protocol.SetAttributeValue("srcipto",end_ip)
	}
	fmt.Println(add_rule.XML())
	xml = root.XML()
	_,err=Conn.NWFilterDefineXML(xml)
	if err!=nil{
		return err
	}
	return  nil
}

//删除一个防火墙策略
func RemoveNWFilterRule(host_name ,name string)error{
	NW,err:=Conn.LookupNWFilterByName(host_name+"_public_firewall")
	if err!=nil{
		return err
	}
	xml,err:=NW.GetXMLDesc(0)
	if err!=nil{
		return err
	}
	root:=xmldom.Must(xmldom.ParseXML(xml)).Root
	rules:=root.GetChildren("rule")
	for _,rule:=range rules{
		if rule.FirstChild().GetAttributeValue("comment")==host_name+"-"+name{
			root.RemoveChild(rule)
		}
	}
	xml = root.XML()
	_,err=Conn.NWFilterDefineXML(xml)
	if err!=nil{
		return err
	}
	return  nil
}

//设置虚拟机状态以及销毁等操作
func SetDomainStatus(status int,host_name string) error {
	domain,err:=Conn.LookupDomainByName(host_name)
	if err!=nil{
		return err
	}
	if status==1{
		err=domain.Create() //启动
		if err!=nil{
			err1:=err.(libvirt.Error)
			if err1.Code==55{
				return nil
			}
		}
	}else if status==2{
		err=domain.Shutdown() //关机
		if err!=nil{
			err1:=err.(libvirt.Error)
			if err1.Code==55{
				return nil
			}
		}
	}else if status==3{
		err=domain.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT) //重启动
	}else if status==4{
		err=domain.Destroy() //硬件关机
		if err!=nil{
			err1:=err.(libvirt.Error)
			if err1.Code==55{
				return nil
			}
		}
	}else if status==5{
		err=domain.SetAutostart(true) //auto
	}else if status==6{
		err=domain.SetAutostart(false) //auto
	}else if status==7{
		bool_,_:=domain.IsActive()
		if bool_==true{
			domain.Destroy()
		}
		err=domain.Undefine()//删除虚拟机
	}else if status==8{
		bool_,_:=domain.IsActive()
		if bool_==true{
			domain.Resume()//运行
		}else{
			err=domain.Create() //启动
			if err!=nil{
				err1:=err.(libvirt.Error)
				if err1.Code==55{
					return nil
				}
			}
		}

	}else if status==9{
		bool_,_:=domain.IsActive()
		if bool_==true{
			err = domain.Suspend() //暂停
		}
	}else if status==10{//硬件重启
		bool_,_:=domain.IsActive()
		if bool_==true{
			domain.Destroy()
		}
		domain.Create()
	}
	if err!=nil{
		return err
	}
	return nil

}

//修改硬盘大小
func VolResize(path string,capacity uint64)error{
	vol,err:=Conn.LookupStorageVolByPath(path)
	vinfo,_:=vol.GetInfo()
	if err!=nil{
		return  err
	}
	err=vol.Resize(capacity-vinfo.Capacity,2)
	if err!=nil{
		return  err
	}
	pool,err:=vol.LookupPoolByVolume()
	if err!=nil{
		return  err
	}
	pool.Refresh(0)
	vol.Free()
	pool.Free()
	return  err
}

//删除一个硬盘
func DeleteVolume(poolName ,volumeName string) error {
	pool, err := Conn.LookupStoragePoolByName(poolName)
	if err!=nil{
		return  err
	}
	defer pool.Free()
	volume,err:=pool.LookupStorageVolByName(volumeName)
	if err!=nil{
		return  err
	}
	volume.Delete(0)
	return  err
}

//更新clock设备
func UpClock(host_name ,clock string)error{
	domain, err := Conn.LookupDomainByName(host_name)
	if err!=nil{
		return err
	}
	xml,_:=domain.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)
	re,_:=regexp.Compile("\\<clock offset\\='[a-zA-z]+'\\>")
	domainXml :=re.ReplaceAllString(xml,clock)
    _,err=Conn.DomainDefineXML(domainXml)
	if err!=nil{
		return err
	}
	return nil
}

//QemuAgentCommand
func QemuAgentCommand(host_name ,command string) error{
	domain, err := Conn.LookupDomainByName(host_name)
	if err!=nil{
		return  err
	}
	_,err=domain.QemuAgentCommand(command,libvirt.DOMAIN_QEMU_AGENT_COMMAND_MIN,0)
	if err!=nil{
		return  err
	}
	return  err
}

//获取虚拟机状态
func GetDomainStatus(host_name string)bool{
	domain,err:=Conn.LookupDomainByName(host_name)
	if err!=nil{
		return false
	}

	bool,err:=domain.IsActive()

	if err!=nil{
		return false
	}
	//conn.Close()
	return  bool
}

//创建快照内部
func CreateSnapshot(data map[string]string)error{
	xmlTemplate:=GetSnapshotXml()
	t:=template.New("getSnapshotXml")
	t,_=t.Parse(xmlTemplate)
	var buf bytes.Buffer
	t.Execute(&buf,data)
	xmlConfig := buf.String()
	domain,err:=Conn.LookupDomainByName(data["host_name"])
	if err==nil{
		snapshot,err:=domain.CreateSnapshotXML(xmlConfig,0)
		if err!=nil{
			return  err
		}
		snapshot.IsCurrent(0)
	}

	return  err
}

//还原一个指定快照
func RestoreSnapshot(data map[string]string)error{
	domain,err:=Conn.LookupDomainByName(data["host_name"])
	if err==nil{
		snapshot,err:=domain.SnapshotLookupByName(data["name"],0)
		if err!=nil{
			return  err
		}
		err=snapshot.RevertToSnapshot(0)
	}
	return  err
}

//删除一个指定快照
func RemoveSnapshot(data map[string]string)error{
	domain,err:=Conn.LookupDomainByName(data["host_name"])
	if err==nil{
		snapshot,err:=domain.SnapshotLookupByName(data["name"],0)
		if err!=nil{
			return  err
		}
		err=snapshot.Delete(0)
	}
	return  err
}

//修改ip
func UpdateIPV4(data map[string]string)error{
	domain,err:=Conn.LookupDomainByName(data["host_name"])
	if err!=nil{
		return  err
	}
	defer domain.Free()
	domainXml,err:=domain.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)
	if err!=nil{
		return  err
	}
	root := xmldom.Must(xmldom.ParseXML(domainXml)).Root
	inter:=root.GetChild("devices").GetChildren("interface")[0].XML()
	reg := regexp.MustCompile(`(?U)<filterref(.*)>(.*)</filterref>`)
	xml_string := reg.ReplaceAllString(inter,"<filterref$1>"+data["filterip"]+"</filterref>")

	if ok,_:=domain.IsActive();ok==true{
		err=domain.UpdateDeviceFlags(xml_string,libvirt.DOMAIN_DEVICE_MODIFY_CONFIG|libvirt.DOMAIN_DEVICE_MODIFY_LIVE)
	}else{
		err=domain.UpdateDeviceFlags(xml_string,libvirt.DOMAIN_DEVICE_MODIFY_CONFIG)
	}
	return  nil
}

//更新一个iso
func UpdateIso(data map[string]string)error{
	domain,err:=Conn.LookupDomainByName(data["host_name"])
	iso_path:=data["iso_path"]
	if err!=nil{
		return  err
	}
	defer domain.Free()
	domainXml,err:=domain.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)
	if err!=nil{
		return  err
	}
	root := xmldom.Must(xmldom.ParseXML(domainXml)).Root
	disks:=root.GetChild("devices").GetChildren("disk")
	var disk_xml string
	for _,disk:=range disks{ //device
		if disk.GetAttributeValue("device")=="cdrom"{
			if iso_path==""{ //卸载
				source_node:=disk.GetChild("source")
				if source_node!=nil {
					disk.RemoveChild(source_node)
				}
			}else{ //挂载
				source_node:=disk.GetChild("source")
				if source_node==nil{
					disk.CreateNode("source").SetAttributeValue("file",iso_path)
				}else{
					source_node.SetAttributeValue("file",iso_path)
				}
			}
			disk_xml= disk.XML()
		}
	}
    if disk_xml==""{
		return   errors.New("xml error")
	}
	if ok,_:=domain.IsActive();ok==true{
		err=domain.UpdateDeviceFlags(disk_xml,libvirt.DOMAIN_DEVICE_MODIFY_CONFIG|libvirt.DOMAIN_DEVICE_MODIFY_LIVE)
	}else{
		err=domain.UpdateDeviceFlags(disk_xml,libvirt.DOMAIN_DEVICE_MODIFY_CONFIG)
	}
	if iso_path==""{
		BootOrder("ide",data["host_name"])
	}else{
		BootOrder("iso",data["host_name"])
	}
	return  nil
}
 
//创建一个备份
func CreateBackup(data map[string]string)error{
	host_name:=data["host_name"]
	name:=data["name"]
	domain,err:=Conn.LookupDomainByName(host_name)
	backup_path:=data["backup_path"]
	if err!=nil{
		return  err
	}
	all_path:=backup_path+"/"+host_name+"/"+name+"/"
	all_path=strings.ReplaceAll(all_path,"\\","/")
	all_path = strings.ReplaceAll(all_path,"//","/")
	err = mkdir(all_path)
	if err!=nil{
		return  err
	}
	defer domain.Free()
	domainXml,err:=domain.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)
	if err!=nil{
		return  err
	}
	root := xmldom.Must(xmldom.ParseXML(domainXml)).Root
	disks:=root.GetChild("devices").GetChildren("disk")
	for _,disk:=range disks{ //device
		if disk.GetAttributeValue("device")=="cdrom"{
			continue
		}
		file_path:=disk.GetChild("source").GetAttributeValue("file")
		if file_path==""{
			continue
		}
		file_name:=filepath.Base(file_path)
		_,err=CopyFile(all_path+"/"+file_name,file_path);
		if err!=nil{
			break
		}

	}
	if err!=nil{
		return  err
	}
	return  nil
}

//还原一个备份
func RestoreBackup(data map[string]string)error{
	host_name:=data["host_name"]
	name:=data["name"]
	domain,err:=Conn.LookupDomainByName(host_name)
	backup_path:=data["backup_path"]
	if err!=nil{
		return  err
	}
	err=SetDomainStatus(4,data["host_name"])
	if err!=nil{
		return  err
	}
	all_path:=backup_path+"/"+host_name+"/"+name+"/"
	all_path=strings.ReplaceAll(all_path,"\\","/")
	all_path = strings.ReplaceAll(all_path,"//","/")
	defer domain.Free()
	domainXml,err:=domain.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)
	if err!=nil{
		return  err
	}
	root := xmldom.Must(xmldom.ParseXML(domainXml)).Root
	disks:=root.GetChild("devices").GetChildren("disk")
	for _,disk:=range disks{ //device
		if disk.GetAttributeValue("device")=="cdrom"{
			continue
		}
		file_path:=disk.GetChild("source").GetAttributeValue("file")
		if file_path==""{
			continue
		}
		file_name:=filepath.Base(file_path)
		_,err=CopyFile(file_path,all_path+"/"+file_name);
		if err!=nil{
			break
		}

	}
	if err!=nil{
		return  err
	}
	err=SetDomainStatus(1,data["host_name"])
	if err!=nil{
		return  err
	}
	return  nil
}

//删除一个备份
func RemoveBackup(data map[string]string)error{
	host_name:=data["host_name"]
	name:=data["name"]
	backup_path:=data["backup_path"]
	all_path:=backup_path+"/"+host_name+"/"+name
	all_path=strings.ReplaceAll(all_path,"\\","/")
	all_path = strings.ReplaceAll(all_path,"//","/")
	err:=os.RemoveAll(all_path)
	if err!=nil{
		return  err
	}
	return  nil
}

//cpu使用情况
func GetDomainCpuUse(host_name string)string{
	domain,err:= Conn.LookupDomainByName(host_name)
	if err !=nil {
		return "0"
	}
	t1 := (time.Now()).Unix()
	domainInfo,err:=domain.GetInfo()
	if err !=nil {
		return "0"
	}
	c1:=domainInfo.CpuTime
	time.Sleep(time.Second)
	t2 := (time.Now()).Unix()
	domainInfo2,err:=domain.GetInfo()
	if err !=nil {
		return "0"
	}
	c2:=domainInfo2.CpuTime
	c_nums := domainInfo2.NrVirtCpu
	usage := float64(float64(c2-c1)*100)/float64((float64(t2)-float64(t1))*float64(c_nums)*1e9)
	return strconv.FormatFloat(usage,'f',-1,64)
}

//网络使用情况 name = network name or mac
func GetDomainNetworkIO(host_name,name string) (string,string,string,string) {
	domain,err:= Conn.LookupDomainByName(host_name)
	if err !=nil {
		return "0","0","0","0"
	}
	domainInterfaceStats,err:=domain.InterfaceStats(name)  //rx bytes 流入（接受）数据  Tx bytes 流出（发送）数据  pakets 包
	if err !=nil {
		return "0","0","0","0"
	}
	//intterval:=1
	RxBytes:=float64(domainInterfaceStats.RxBytes)
	TxBytes:=float64(domainInterfaceStats.TxBytes)
	time.Sleep(time.Second*1)
	domainInterfaceStats1,err:=domain.InterfaceStats(name)  //rx bytes 流入（接受）数据  Tx bytes 流出（发送）数据  pakets 包
	if err !=nil {
		return "0","0","0","0"
	}
	RxBytes1:=float64(domainInterfaceStats1.RxBytes)
	TxBytes1:=float64(domainInterfaceStats1.TxBytes)
	up:= (TxBytes1-TxBytes)
	down:=(RxBytes1-RxBytes)
	up,down,sumup,sumdown:=formatFloat(up/1000),formatFloat(down/1000),formatFloat(TxBytes1/1000),formatFloat(RxBytes1/1000)
	return strconv.FormatFloat(float64(up),'f',-1,64),strconv.FormatFloat(float64(down),'f',-1,64),strconv.FormatFloat(float64(sumup),'f',-1,64),strconv.FormatFloat(float64(sumdown),'f',-1,64)

}

//内存使用情况 name = network name or mac
func GetDomainMemoryStats(host_name string)(string){
	domain,err:=Conn.LookupDomainByName(host_name)
	if err !=nil {
		return "0"
	}
	err=domain.SetMemoryStatsPeriod(10,0)
	if err !=nil {
		return "0"
	}
	s,err:=domain.MemoryStats(0x8,0)
	if err !=nil {
		return "0"
	}
	MemoryStats:=DomainMemoryStats{}
	for _,v:=range s{
		if v.Tag==6{
			MemoryStats.Actual=v.Val
		}else if v.Tag==1{
			MemoryStats.Swap_out=v.Val
		}else if v.Tag==0{
			MemoryStats.Swap_in=v.Val
		}else if v.Tag==2{
			MemoryStats.Major_fault=v.Val
		}else if v.Tag==3{
			MemoryStats.Minor_fault=v.Val
		}else if v.Tag==4{
			MemoryStats.Unused=v.Val
		}else if v.Tag==5{
			MemoryStats.Available=v.Val
		}else if v.Tag==9{
			MemoryStats.Last_update=v.Val
		}else if v.Tag==7{
			MemoryStats.Rss=v.Val
		}
	}

	free:=(MemoryStats.Available-MemoryStats.Unused)
	free_:=float32(free)
	util_mem:=(free_/float32(MemoryStats.Available))*100
	return strconv.FormatFloat(formatFloat(float64(util_mem)),'f',-1,64)
}

//监控硬盘容量信息
func GetVolumeInfo(host_name ,dev string)(string,string){
	domain,err:=Conn.LookupDomainByName(host_name)
	if err!=nil{
		return "0","0"
	}
	block,err:=domain.GetBlockInfo(dev,0)
	if err!=nil{
		return "0","0"
	}else{
		use:=formatFloat(float64(block.Allocation)/(1024*1024*1024))
		capacity:=formatFloat(float64(block.Capacity)/(1024*1024*1024))
		return strconv.FormatFloat(use,'f',-1,64),strconv.FormatFloat(capacity,'f',-1,64)
	}
}

//更新引导顺序
func BootOrder(boot,host_name string)error{
	domain_, err := Conn.LookupDomainByName(host_name)
	if err !=nil {
		return err
	}
	domainXml,err:=domain_.GetXMLDesc(0)
	if err!=nil{
		return err
	}

	root := xmldom.Must(xmldom.ParseXML(domainXml)).Root
	nodeList := root.Query("//disk")
	sort:=map[string]string{}
	for _, node := range nodeList {
		if node ==nil{
			continue
		}
		target := node.GetChild("target")
		if target ==nil{
			continue
		}
		dev := target.GetAttribute("dev")
		if dev.Value=="vda"{
			boot:=node.GetChild("boot")
			if boot==nil{
				continue
			}
			order:=boot.GetAttribute("order").Value
			sort["vda"] = order
		}
		if dev.Value=="hdc"{
			boot:=node.GetChild("boot")
			if boot==nil{
				continue
			}
			order:=boot.GetAttribute("order").Value
			sort["hdc"] = order
		}

	}

	if boot=="iso"{
		hdc,ok:=sort["hdc"]
		if ok==true{
			expr:=fmt.Sprintf(`<boot order='%s'(.*)/>`,hdc)
			r,_:=regexp.Compile(expr)
			domainXml=r.ReplaceAllString(domainXml,`<boot order='N1'/>`)
		}else{
			r,_:=regexp.Compile(`<target dev='hdc' bus='ide'/>`)
			domainXml = r.ReplaceAllString(domainXml,`<target dev='hdc' bus='ide'/> 
      <boot order='N1'/>`)

		}
		vda:=sort["vda"]
		expr:=fmt.Sprintf(`<boot order='%s'(.*)/>`,vda)

		r,_:=regexp.Compile(expr)
		domainXml=r.ReplaceAllString(domainXml,`<boot order='2'/>`)

	}else{
		hdc,ok:=sort["hdc"]
		if ok==true{
			expr:=fmt.Sprintf(`<boot order='%s'(.*)/>`,hdc)
			r,_:=regexp.Compile(expr)
			domainXml=r.ReplaceAllString(domainXml,`<boot order='N2'/>`)
		}else{
			r,_:=regexp.Compile(`<target dev='hdc' bus='ide'/>`)
			domainXml = r.ReplaceAllString(domainXml,`<target dev='hdc' bus='ide'/> 
      <boot order='N2'/>`)
		}

		vda:=sort["vda"]
		expr:=fmt.Sprintf(`<boot order='%s'(.*)/>`,vda)
		r,_:=regexp.Compile(expr)
		domainXml=r.ReplaceAllString(domainXml,`<boot order='1'/>`)
	}
	r,_:=regexp.Compile(`<boot order=['"]N`)
	domainXml=r.ReplaceAllString(domainXml,`<boot order='`)
	_,err=Conn.DomainDefineXML(domainXml)
	return err
}

func GetIsoList(iso_path string)(map[string]string,error){
	list:=map[string]string{}
	dir,err:=ioutil.ReadDir(iso_path)
	if err!=nil{
		return nil,err
	}
	PthSep:=string(os.PathSeparator)
	for _,file:=range dir{
		if file.IsDir(){
			continue
		}
		list[file.Name()]=iso_path+PthSep+file.Name()
	}

	return list,nil

}

func NetworkCloseOrOpen(host_name ,state string)error{
	domain,err:=Conn.LookupDomainByName(host_name)
	if err!=nil{
		return  err
	}
	defer domain.Free()
	domainXml,err:=domain.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)
	if err!=nil{
		return  err
	}
	root := xmldom.Must(xmldom.ParseXML(domainXml)).Root
	interfaces:=root.GetChild("devices").GetChildren("interface")
	xml:=""
	for _,interface_:=range interfaces{ //device
		if state=="1"{ //open
			if interface_.GetAttributeValue("type")=="network"{
				interfacexml_:=interface_.XML()
				xml =strings.Replace(interfacexml_,"network","bridge",-1)
				xml =strings.Replace(interfacexml_,"default","br0",-1)
			}
		}else{//close
			if interface_.GetAttributeValue("type")=="bridge"{
				interfacexml_:=interface_.XML()
				xml =strings.Replace(interfacexml_,"bridge","network",-1)
				xml =strings.Replace(interfacexml_,"br0","default",-1)
			}
		}
        if xml!=""{
			if ok,_:=domain.IsActive();ok==true{
				err=domain.UpdateDeviceFlags(xml,libvirt.DOMAIN_DEVICE_MODIFY_CONFIG|libvirt.DOMAIN_DEVICE_MODIFY_LIVE)
			}else{
				err=domain.UpdateDeviceFlags(xml,libvirt.DOMAIN_DEVICE_MODIFY_CONFIG)
			}
			if err!=nil{
				return  err
			}
		}

	}

	return  nil
}

func CountFlow(host_name string)(tx int64,rx int64){
	_,error:=Exec_shell("vnstat  -u -i "+host_name)
	if error!=nil{
		return 0,0
	}
	str,error:=Exec_shell("vnstat  --dumpdb -i "+host_name)
	if error!=nil{
		return 0,0
	}

	lines:=strings.Split(str,"\n")
	if len(lines)<20{
		return 0,0
	}
	//var tx,rx int64
	for _,line:=range  lines {
		m:=strings.Split(line,";")
		if m[0]=="m"&& m[1]=="0"{
			tx1,_:=strconv.ParseInt(m[4],10,64)
			tx2,_:=strconv.ParseInt(m[6],10,64)
			rx1,_:=strconv.ParseInt(m[3],10,64)
			rx2,_:=strconv.ParseInt(m[5],10,64)
			tx=tx1*1024+tx2
			rx=rx1*1024+rx2
		}
	}
	return tx,rx
}

