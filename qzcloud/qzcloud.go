package qzcloud

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"qzkvm/src/pkg/app"
	"strconv"
	"strings"
	"time"
)

//创建
func CreateKvm(c *gin.Context)  {
	//Createkvm
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=CreateStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	//os.Open()
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	if data["re_create"]=="1"{
		rootdir:=os.Args[0]
		token_path:=rootdir+"/vnc/token/token.conf"
		token_path=strings.Replace(token_path,"//","/",-1)
		deltoken:=fmt.Sprintf("sed -i '/%s: %s:%s/d' %s",data["host_name"],data["ip"],data["vnc_port"],token_path)
		Exec_shell(deltoken)
		SetDomainStatus(7,data["host_name"])
		os.Remove("/var/lib/vnstat/"+data["host_name"])
		RemoveFile(data["backup_path"]+"/"+data["host_name"])
		UnNWFilter(data["host_name"]+"_public_firewall")
		RemoveFile(data["data_path"]+"/"+data["host_name"]) //删除文件
	}
	data["host_uuid"] = GetUUIDBuild()
	//copy system os
	template_path :=data["template_path"]+"/"+data["os_name"]
	data_path:=data["data_path"]+"/"+data["host_name"]
	os_path:=data_path+"/"+data["host_name"]+"_os.qcow2"
	if len(data["os_name"])<2{
		appG.Response(0,"os_name param error","")
		return
	}
	if len(data["template_path"])<2||len(data["data_path"])<2{
		appG.Response(0,"dir path error","")
		return
	}

	err:=mkdir(data_path)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	_,err=CopyFile(os_path,template_path)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	/**
	//create data Pool
	err=CreateStoragePool(data["host_name"],data_path)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	**/

    //create data Volume
	data_vol_path,err:=CreateVolume(data_path,data["host_data"],data["host_name"]+"_data1.qcow2")
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	//多ip
	otherip:=data["otherip"]
	ip:=data["ip"]
	if len(otherip)>6{
		ip = ip+","+otherip
	}

	//多ip防盗
	iparr:=strings.Split(ip,",")
	ipTemplate :=""
	for _,val:=range iparr{
		ipTemplate+= `<parameter name='IP' value='`+val+`'/>
`
	}

	cdrom:=""
	if cdrom,ok:=data["cdrom"];ok{
		cdrom = cdrom
	}

    //同步时间
	clock:=`utc`
	if data["os_type"]=="windows"{
		clock=`localtime`
	}
	//是否显示cpu
	cpu_model:="host-passthrough" //show cpu Host model
	if cpu_m,ok:=data["cpu_model"];ok{
		 if cpu_m=="hide"{
			 cpu_model = "host-model"
		 }
	}

	data["clock"] = clock
	data["cdrom"] = cdrom
	data["cpu_model"] = cpu_model
	data["filterip"] = ipTemplate
	data["bandwidth_in"] = Absum(data["bandwidth"])
	data["bandwidth_out"] = Absum(data["bandwidth"],"128")
	data["data_read"] = Absum(data["data_read"],"1000000")
	data["data_write"] = Absum(data["data_write"],"1000000")
	data["os_read"] = Absum(data["os_read"],"1000000")
	data["os_write"] = Absum(data["os_write"],"1000000")
	data["os_path"] = os_path
	data["data_path"] = data_vol_path
	password := data["password"]
	//创建一个空策略
	UnNWFilter(data["host_name"]+"_public_firewall")
	err=InitNWFilter(data["host_name"])
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	_,err=CreateDomain(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}

	//vnc token
	rootdir:=os.Args[0]
	token_path:=rootdir+"/vnc/token/token.conf"
	token_path=strings.Replace(token_path,"//","/",-1)
	addtoken:=fmt.Sprintf("echo %s: %s:%s >>%s",data["host_name"],data["ip"],data["vnc_port"],token_path)
	Exec_shell(addtoken)
    //setting ip
	if data["os_type"]!="windows" {
		//配置ip
		shname:="change.sh"
		if data["os_type"]=="centos"{
			shname = "change.sh"
		}else{
			shname = "change_other.sh"
		}
		buf:=fmt.Sprintf("chmod 777 /root/linux_ip&&/root/linux_ip\n")
		buf+=fmt.Sprintf("rm -rf /root/linux_ip\n")
		buf+=fmt.Sprintf("rm -rf /root/ip.txt\n")
		buf+=fmt.Sprintf("echo root:'%s'| chpasswd",password)
		ipconfig :=fmt.Sprintf(`%s|%s|%s|%s|%s|%s
%s|%s|%s|%s
%s`,data["ip"],data["gateway"],data["netmask"],data["dns1"],data["dns2"],data["mac"], data["ip1"],data["gateway1"],data["netmask1"],data["mac1"],data["host_name"])
		dir:="/tmp/"+strconv.FormatInt(time.Now().UnixNano(),10)+"/"
		os.MkdirAll(dir,0666)

		file_path := dir+shname
		err := ioutil.WriteFile(file_path, []byte(buf),0666)
		if err != nil {
			appG.Response(0,err.Error(),"")
			return
		}

		file_path_ip:= dir+"ip.txt"
		err = ioutil.WriteFile(file_path_ip, []byte(ipconfig),0666)

		if err != nil {
			appG.Response(0,err.Error(),"")
			return
		}

		dir_, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		cmdstr1:=fmt.Sprintf(`virt-copy-in -d %s %s /root`,data["host_name"],file_path)
		cmds := exec.Command("bash", "-c", cmdstr1)
		cmds.Run()

		cmdstr1=fmt.Sprintf(`virt-copy-in -d %s %s /root`,data["host_name"],file_path_ip)
		cmds = exec.Command("bash", "-c", cmdstr1)
		cmds.Run()


		cmdstr1=fmt.Sprintf(`virt-copy-in -d %s %s /root`,data["host_name"],dir_+"/linux_ip")
		cmds = exec.Command("bash", "-c", cmdstr1)
		cmds.Run()
		os.RemoveAll(dir)

	}else{
		ipconfig:=""
		ipconfig+="Call net user administrator "+password+" \r\n"
		ipconfig+="Call netsh interface ip set address name=\"本地连接\" static "+data["ip"]+" "+data["netmask"]+" "+data["gateway"]+" 1 \r\n"
		ipconfig+=fmt.Sprintf("Call netsh interface ip set dns \"本地连接\" static %s\r\n",data["dns1"])
		ipconfig+=fmt.Sprintf("Call netsh interface ip add dns \"本地连接\" %s \r\n",data["dns2"])
		ipconfig+=fmt.Sprintf("Call wmic computersystem where \"name='%%computername%%'\" call rename %s\r\n",data["host_name"])
		ipconfig+=fmt.Sprintf("Call netsh interface ip set address name=\"本地连接 2\" static %s %s\r\n",data["ip1"],data["netmask1"])
		ipconfig+=fmt.Sprintf("powershell start-service w32time\r\n")
		ipconfig+=fmt.Sprintf("powershell w32tm /config /update /manualpeerlist:time.windows.com /syncfromflags:manual /reliable:yes\r\n")
		ipconfig+=fmt.Sprintf("powershell w32tm /resync\r\n")
		ipconfig+="del %0\r\n"
		enc := mahonia.NewEncoder("GBK")
		output := enc.ConvertString(ipconfig)

		dir:="/tmp/"+strconv.FormatInt(time.Now().UnixNano(),10)+"/"
		os.MkdirAll(dir,0666)

		file_path := dir+"change.bat"
		err = ioutil.WriteFile(file_path, []byte(output),0666)
		if err != nil {
			appG.Response(0,err.Error(),"")
			return
		}

		cmdstr1:=fmt.Sprintf(`virt-copy-in -d %s %s /`,data["host_name"],file_path)
		cmds := exec.Command("bash", "-c", cmdstr1)
		cmds.Run()
		os.RemoveAll(dir)

	}
	err=SetDomainStatus(1,data["host_name"])
	if err != nil {
		appG.Response(0,err.Error(),"")
		return
	}
	//start count flow
	///
	os.Remove("/var/lib/vnstat/"+data["host_name"])
	CountFlow(data["host_name"])
	appG.Response(200,"ok","")
}

//编辑
func UpdateKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=CreateStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)

	//多ip
	otherip:=data["otherip"]
	ip:=data["ip"]
	if len(otherip)>6{
		ip = ip+","+otherip
	}

	//多ip防盗
	iparr:=strings.Split(ip,",")
	ipTemplate :=""
	for _,val:=range iparr{
		ipTemplate+= `<parameter name='IP' value='`+val+`'/>
`
	}

	cdrom:=""
	if cdrom,ok:=data["cdrom"];ok{
		cdrom = cdrom
	}

	//同步时间
	clock:=`utc`
	if data["os_type"]=="windows"{
		clock=`localtime`
	}
	//是否显示cpu
	cpu_model:="host-passthrough" //show cpu Host model
	if cpu_m,ok:=data["cpu_model"];ok{
		if cpu_m=="hide"{
			cpu_model = "host-model"
		}
	}
	data_path:=data["data_path"]+"/"+data["host_name"]
	os_path:=data_path+"/"+data["host_name"]+"_os.qcow2"
	data_vol_path:=data_path+"/"+data["host_name"]+"_data1.qcow2"
	data["clock"] = clock
	data["cdrom"] = cdrom
	data["cpu_model"] = cpu_model
	data["filterip"] = ipTemplate
	data["bandwidth_in"] = Absum(data["bandwidth"])
	data["bandwidth_out"] = Absum(data["bandwidth"],"128")
	data["data_read"] = Absum(data["data_read"],"1000000")
	data["data_write"] = Absum(data["data_write"],"1000000")
	data["os_read"] = Absum(data["os_read"],"1000000")
	data["os_write"] = Absum(data["os_write"],"1000000")
	data["os_path"] = os_path
	data["data_path"] = data_vol_path

	_,err:=CreateDomain(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}

	//修改硬盘 VolResize
	host_data:=Absum(data["host_data"],"1024","1024","1024")
	amount,_:=strconv.ParseInt(host_data,10,64)
	err = VolResize(data_vol_path,uint64(amount))
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	SetDomainStatus(10,data["host_name"])
	appG.Response(200,"ok","")
}

//删除
func RemoveKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=RemoveStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
   /***
	//删除备份 以及快照
	err:=ManageStoragePool(data["host_name"],"undefine")
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
   **/

	err:=SetDomainStatus(7,data["host_name"])
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	rootdir:=os.Args[0]
	token_path:=rootdir+"/vnc/token/token.conf"
	token_path=strings.Replace(token_path,"//","/",-1)
	deltoken:=fmt.Sprintf("sed -i '/%s: %s:%s/d' %s",data["host_name"],data["ip"],data["vnc_port"],token_path)
	Exec_shell(deltoken)
	os.Remove("/var/lib/vnstat/"+data["host_name"])
	RemoveFile(data["backup_path"]+"/"+data["host_name"])
	UnNWFilter(data["host_name"]+"_public_firewall")
	RemoveFile(data["data_path"]+"/"+data["host_name"]) //删除文件
	appG.Response(200,"ok","")
}

//重新安装操作系统
func ReinstallKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=ReinstallStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)

	//关机
	SetDomainStatus(4,data["host_name"])

	//copy system os
	template_path :=data["template_path"]+"/"+data["os_name"]
	data_path:=data["data_path"]+"/"+data["host_name"]
	os_path:=data_path+"/"+data["host_name"]+"_os.qcow2"
	password:=data["password"]
	//删除原系统
	os.Remove(os_path)

	_,err:=CopyFile(os_path,template_path)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}

	//setting ip
	if data["os_type"]!="windows" {
		//配置ip
		shname:="change.sh"
		if data["os_type"]=="centos"{
			shname = "change.sh"
		}else{
			shname = "change_other.sh"
		}
		buf:=fmt.Sprintf("chmod 777 /root/linux_ip&&/root/linux_ip\n")
		buf+=fmt.Sprintf("rm -rf /root/linux_ip\n")
		buf+=fmt.Sprintf("rm -rf /root/ip.txt\n")
		buf+=fmt.Sprintf(`echo root:'%s'| chpasswd`,password)
		ipconfig :=fmt.Sprintf(`%s|%s|%s|%s|%s|%s
%s|%s|%s|%s
%s`,data["ip"],data["gateway"],data["netmask"],data["dns1"],data["dns2"],data["mac"], data["ip1"],data["gateway1"],data["netmask1"],data["mac1"],data["host_name"])
		dir:="/tmp/"+strconv.FormatInt(time.Now().UnixNano(),10)+"/"
		os.MkdirAll(dir,0666)

		file_path := dir+shname
		err := ioutil.WriteFile(file_path, []byte(buf),0666)
		if err != nil {
			appG.Response(0,err.Error(),"")
			return
		}

		file_path_ip:= dir+"ip.txt"
		err = ioutil.WriteFile(file_path_ip, []byte(ipconfig),0666)

		if err != nil {
			appG.Response(0,err.Error(),"")
			return
		}

		dir_, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		cmdstr1:=fmt.Sprintf(`virt-copy-in -d %s %s /root`,data["host_name"],file_path)
		cmds := exec.Command("bash", "-c", cmdstr1)
		cmds.Run()

		cmdstr1=fmt.Sprintf(`virt-copy-in -d %s %s /root`,data["host_name"],file_path_ip)
		cmds = exec.Command("bash", "-c", cmdstr1)
		cmds.Run()


		cmdstr1=fmt.Sprintf(`virt-copy-in -d %s %s /root`,data["host_name"],dir_+"/linux_ip")
		cmds = exec.Command("bash", "-c", cmdstr1)
		cmds.Run()
		os.RemoveAll(dir)

	}else{
		ipconfig:=""
		ipconfig+="Call net user administrator "+password+" \r\n"
		ipconfig+="Call netsh interface ip set address name=\"本地连接\" static "+data["ip"]+" "+data["netmask"]+" "+data["gateway"]+" 1 \r\n"
		ipconfig+=fmt.Sprintf("Call netsh interface ip set dns \"本地连接\" static %s\r\n",data["dns1"])
		ipconfig+=fmt.Sprintf("Call netsh interface ip add dns \"本地连接\" %s \r\n",data["dns2"])
		ipconfig+=fmt.Sprintf("Call wmic computersystem where \"name='%%computername%%'\" call rename %s\r\n",data["host_name"])
		ipconfig+=fmt.Sprintf("Call netsh interface ip set address name=\"本地连接 2\" static %s %s\r\n",data["ip1"],data["netmask1"])
		ipconfig+=fmt.Sprintf("powershell start-service w32time\r\n")
		ipconfig+=fmt.Sprintf("powershell w32tm /config /update /manualpeerlist:time.windows.com /syncfromflags:manual /reliable:yes\r\n")
		ipconfig+=fmt.Sprintf("powershell w32tm /resync\r\n")
		ipconfig+="del %0\r\n"
		enc := mahonia.NewEncoder("GBK")
		output := enc.ConvertString(ipconfig)

		dir:="/tmp/"+strconv.FormatInt(time.Now().UnixNano(),10)+"/"
		os.MkdirAll(dir,0666)

		file_path := dir+"change.bat"
		err = ioutil.WriteFile(file_path, []byte(output),0666)
		if err != nil {
			appG.Response(0,err.Error(),"")
			return
		}

		cmdstr1:=fmt.Sprintf(`virt-copy-in -d %s %s /`,data["host_name"],file_path)
		cmds := exec.Command("bash", "-c", cmdstr1)
		cmds.Run()
		os.RemoveAll(dir)
		err=SetDomainStatus(1,data["host_name"])
		if err != nil {
			appG.Response(0,err.Error(),"")
			return
		}
	}
	//同步时间
	clock:=`utc`
	if data["os_type"]=="windows"{
		clock=`localtime`
	}

	data["clock"] = clock
	err = UpClock(data["host_name"],"<clock offset='"+clock+"'>")
	if err != nil {
		appG.Response(0,err.Error(),"")
		return
	}
	err = SetDomainStatus(1,data["host_name"])
	if err != nil {
		appG.Response(0,err.Error(),"")
		return
	}

	appG.Response(200,"ok","")

}

//重新设置系统密码
func UpdateSystemPassword(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=UpdateSystemPasswordStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)

	password:=data["password"]
	username:=data["username"]

	bool_:=GetDomainStatus(data["host_name"])
	if bool_==false{
		appG.Response(0,"重置密码云主机必须处在运行状态","")
		return
	}

	password=base64.StdEncoding.EncodeToString([]byte(password))

	cmdStr:=fmt.Sprintf(`{ "execute": "guest-set-user-password", "arguments": {"username":"%s","password":"%s","crypted":false} }`,username,password)

	err:=QemuAgentCommand(data["host_name"],cmdStr)
	if err!=nil{
		appG.Response(0,"重置密码失败","")
		return
	}
	appG.Response(200,"ok","")
}

//创建快照内部
func CreateSnapshotkvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=CreateSnapshotStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)

	err:=CreateSnapshot(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"ok","")
}

//还原一个指定快照
func RestoreSnapshotKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=RestoreSnapshotStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)

	err:=RestoreSnapshot(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"ok","")
}

//删除一个指定快照
func RemoveSnapshotKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=RemoveSnapshotStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)

	err:=RemoveSnapshot(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"ok","")
}

//修改ip
func UpdateIPKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=UpdateIPStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)

	//多ip
	otherip:=data["otherip"]
	ip:=data["ip"]
	if len(otherip)>6{
		ip = ip+","+otherip
	}

	//多ip防盗
	iparr:=strings.Split(ip,",")
	ipTemplate :=""
	for _,val:=range iparr{
		ipTemplate+= `<parameter name='IP' value='`+val+`'/>
`
	}
	data["filterip"] = ipTemplate
	err:=UpdateIPV4(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"ok","")
}

//添加一个防火墙策略
func AddNWFilterKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=AddNWFilterStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	err:=AddNWFilterRule(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"ok","")
}

//删除一个防火墙策略
func RemoveNWFilterKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=RemoveNWFilterStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	err:=RemoveNWFilterRule(data["host_name"],data["name"])
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"ok","")
}

//更新一个iso
func UpdateIsoKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=UpdateIsoStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	err:=UpdateIso(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"ok","")
}

//创建备份
func CreateBackupKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=BackupStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
    err:= CreateBackup(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}

	appG.Response(200,"ok","")
}

//还原备份
func RestoreBackupKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=BackupStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	err:= RestoreBackup(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}

	appG.Response(200,"ok","")
}

//删除备份
func RemoveBackupKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=BackupStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	err:= RemoveBackup(data)
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}

	appG.Response(200,"ok","")
}
 

 //监控 cpu 内存 网络
func MonitorKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=MonitorStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	useCpu:=GetDomainCpuUse(data["host_name"])
	useMemory:=GetDomainMemoryStats(data["host_name"])
	up,down,sumup,sumdown:=GetDomainNetworkIO(data["host_name"],data["network_name"])
	appG.Response(200,"ok",map[string]string{
		"use_cpu":useCpu,
		"use_memory":useMemory,
		"up":up,
		"down":down,
		"sum_up":sumup,
		"sum_down":sumdown,
	})
}

//启动项
func BootOrderKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=BootOrderStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	err:=BootOrder(data["boot"],data["host_name"])
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"","")
}

func GetNetworkFlowKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=MonitorStruct{}
	if err:=c.BindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	up,down,sumup,sumdown:=GetDomainNetworkIO(data["host_name"],data["network_name"])
	appG.Response(200,"ok",map[string]string{
		"use_cpu":"",
		"use_memory":"",
		"up":up,
		"down":down,
		"sum_up":sumup,
		"sum_down":sumdown,
	})
}

//kvm status
func GetStatusKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=HostNameStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)

	bool_:=GetDomainStatus(data["host_name"])
	appG.Response(200,"ok",bool_)
}

func SetStatusKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=SetStatusStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	state,_:=strconv.ParseInt(data["state"],10,64)
	err:=SetDomainStatus(int(state),data["host_name"])
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"","")
}

func GetIsoListKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=GetIsoListStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	list,err:=GetIsoList(data["iso_path"])
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"",list)
}

func NetworkCloseOrOpenKvm(c *gin.Context){
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=SetStatusStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	err:=NetworkCloseOrOpen(data["host_name"],data["state"])
	if err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	appG.Response(200,"","")
}

func CountFlowKvm(c *gin.Context)  {
	data:=make(map[string]string)
	appG := app.Gin{C: c}
	form:=HostNameStruct{}
	if err:=c.ShouldBindJSON(&form);err!=nil{
		appG.Response(0,err.Error(),"")
		return
	}
	json_,_:=json.Marshal(form)
	json.Unmarshal(json_,&data)
	tx,rx:=CountFlow(data["host_name"])

	appG.Response(200,"",map[string]int64{
		"tx":tx,
		"rx":rx,
	})
}