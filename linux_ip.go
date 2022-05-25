/*
Package guest_set_user_password - sync host<->guest communication

Example:
{ "execute": "guest-get-cloud-init", "arguments": {"cmd":"cloud_kvm 192.168.10.2 192.168.10.1 255.255.255.0 8.8.8.8 8.8.8.6 "} }
*/
package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)
const (
	centos              = "/etc/sysconfig/network-scripts/ifcfg-"
	ubuntu              = "/etc/network/interfaces"
	ubuntu1             = "/etc/network/interfaces.d/cloud.cfg"
	ubuntu2             = "/etc/netplan/50-cloud-init.yaml"
	centos_network_conf = `TYPE=Ethernet
BOOTPROTO=static
HWADDR={{.mac}}
IPADDR={{.ip}}
NETMASK={{.mask}}
GATEWAY={{.gateway}}
DNS1={{.dns1}}
DNS2={{.dns2}}
DEFROUTE=yes
IPV4_FAILURE_FATAL=yes
IPV6INIT=no
NAME={{.name}}
DEVICE={{.name}}
ONBOOT=yes`
	centos_network_conf1 = `TYPE=Ethernet
BOOTPROTO=static
IPADDR={{.ip1}}
NETMASK={{.mask1}}
DEFROUTE=yes
IPV4_FAILURE_FATAL=yes
NAME={{.name1}}
DEVICE={{.name1}}
ONBOOT=yes`
	ubuntu1_network_conf = `# This file describes the network interfaces available on your system
# and how to activate them. For more information, see interfaces(5).

source /etc/network/interfaces.d/*
# The loopback network interface
auto lo
iface lo inet loopback
`
	ubuntu2_network_conf1 = `auto {{.name}}
iface {{.name}} inet static
	address {{.ip}}
	dns-nameservers {{.dns1}} {{.dns2}}
	netmask {{.mask}}
	gateway {{.gateway}}
auto {{.name1}}
iface {{.name1}} inet static
	address {{.ip1}}
	netmask {{.mask1}}`
	ubuntu2_network_conf2 = `network:
    renderer: networkd
    ethernets:
        {{.name}}:
            addresses: [{{.ip}}/{{.mask}}]
            gateway4: {{.gateway}}
            nameservers:
                addresses: [{{.dns1}}, {{.dns2}}]
                search: []
            optional: true
        {{.name1}}:
            addresses: [{{.ip1}}/{{.mask1}}]
            optional: true
    version: 2`
)

func main()  {
	f, err := os.Open("/root/ip.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	ipInfo:=map[string]string{}
	scanner := bufio.NewScanner(f)
	os_,ver_:=OsAndVer()
    i:=0
	for scanner.Scan(){
		i+=1
		line:=scanner.Text() //以'\n'为结束符读入一行
		line = strings.Replace(line,"\n","",1)
		line = strings.Trim(line,"")
		//ip|gateway|mask|dns1|dns2|mac
		/**
		  192.168.10.1.2|192.168.10.1|255.255.255.0|8.8.8.8|8.8.8.6|mac
		  192.168.10.1.2|192.168.10.1|255.255.255.0|eth1
		  hostname
		 */
		ip:=strings.Split(line,"|")
		if i==1{
			ipInfo["ip"] = ip[0]
			ipInfo["gateway"] = ip[1]
			ipInfo["mask"] = ip[2]
			ipInfo["dns1"] = ip[3]
			ipInfo["dns2"] = ip[4]
			ipInfo["mac"] = ip[5]
		}else if(i==2){
			ipInfo["ip1"] = ip[0]
			ipInfo["gateway1"] = ip[1]
			ipInfo["mask1"] = ip[2]
			ipInfo["mac1"] = ip[3]
		}else{
			SetHostName(os_,line)
		}

	}
	SetIp(ipInfo,os_,ver_)
}
func SetIp(ipInfo map[string]string, os_, ver string) {
	tag := InterfacesTag(ipInfo["mac"])
	tag1 := InterfacesTag(ipInfo["mac1"])
	ipInfo["name"] = tag
	ipInfo["name1"] = tag1
	if strings.Index(strings.ToLower(os_),"centos")>-1  || strings.Index(strings.ToLower(os_),"fedora")>-1  {
		os.Remove(centos+"eth0")
		path := centos + tag
		os.Remove(path)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		t := template.New("ipset")
		t.Parse(centos_network_conf)
		t.Execute(file, ipInfo)

		os.Remove(centos+"eth1")
		path1:= centos + tag1
		file1, err := os.OpenFile(path1, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			//panic(err)
		}
		defer file1.Close()
		t = template.New("ipset1")
		t.Parse(centos_network_conf1)
		t.Execute(file1, ipInfo)

		var ver1 float64
		if ver != "" {
			ver1, _ = strconv.ParseFloat(ver, 10)
		}
		if strings.ToLower(os_) == "fedora" {
			cmd := exec.Command("bash", "-c", "nmcli con reload;nmcli con reload;nmcli con up "+tag+";systemctl restart network")
			cmd.Run()
		}
		if ver1 >= 7 && ver1<8{
			cmd := exec.Command("bash", "-c", "systemctl stop NetworkManager;systemctl disable NetworkManager")
			cmd.Run()

			com:=fmt.Sprintf("ifconfig %s down;ifconfig %s up;ifconfig %s down;ifconfig %s up",tag,tag,tag1,tag1)
			cmd = exec.Command("bash", "-c", com)
			cmd.Run()

			cmd = exec.Command("bash", "-c", "systemctl restart network")
			cmd.Run()

		}else if ver1 >= 8 && ver1<9{
			cmd := exec.Command("bash", "-c", "systemctl start NetworkManager;systemctl enable NetworkManager")
			cmd.Run()

			com:=fmt.Sprintf("ifdown %s;ifup %s;ifdown %s ;ifup %s up",tag,tag,tag1,tag1)
			cmd = exec.Command("bash", "-c", com)
			cmd.Run()

			cmd = exec.Command("bash", "-c", "systemctl restart NetworkManager")
			cmd.Run()
		} else {
			cmd := exec.Command("bash", "-c", "service  NetworkManager stop;chkconfig NetworkManager off;service network restart")
			cmd.Run()
		}
	}
	if strings.ToLower(os_) == "ubuntu" || strings.ToLower(os_) == "debian" {
		if ver != "" {
			ver2, _ := strconv.ParseFloat(ver, 10)
			if ver2 <= 17 || strings.ToLower(os_) == "debian" {
				ioutil.WriteFile(ubuntu, []byte(ubuntu1_network_conf), 0666) //写入文件(字节数组)
				file, err := os.OpenFile(ubuntu1, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					panic(err)
				}
				defer file.Close()
				t := template.New("ipset")
				t.Parse(ubuntu2_network_conf1)
				t.Execute(file, ipInfo)

                dns:=fmt.Sprintf(`echo nameserver %s > /etc/resolv.conf;echo nameserver %s >> /etc/resolv.conf`,ipInfo["dns1"],ipInfo["dns2"])

				cmd := exec.Command("bash", "-c", dns)
				cmd.Run()
                cmd = exec.Command("bash", "-c", "/etc/init.d/networking restart")
				cmd.Run()
			}
			if ver2 > 17 {
				cmd1 := exec.Command("bash", "-c", "sudo rm -r /etc/netplan/*.ymal")
				cmd1.Run()
				file, err := os.OpenFile(ubuntu2, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
				if err != nil {
					panic(err)
				}
				defer file.Close()
				t := template.New("ipset")
				t.Parse(ubuntu2_network_conf2)
				ipInfo["mask"] = MASK2CIDR(ipInfo["mask"])
				t.Execute(file, ipInfo)
				cmd := exec.Command("bash", "-c", "sudo netplan apply;sudo service network-manager stop")
				cmd.Run()
			}
		}
	}
}

func SetHostName(os_, newHostName string) {
	if strings.ToLower(os_) == "centos"  {
		network_, _ := ioutil.ReadFile("/etc/sysconfig/network")
		osRegexp_ := regexp.MustCompile("[^_]HOSTNAME=\"?(.*)\"?")
		hostName := osRegexp_.FindSubmatch(network_)
		if len(hostName)>0{
			hostName_ := strings.Trim(string(hostName[1]), " ")
			hostName1_ := strings.Trim(string(hostName_), "\n")
			com1 := fmt.Sprintf(`sudo sed -i "s/%s/%s/g" /etc/sysconfig/network`, hostName1_, newHostName)
			com2 := fmt.Sprintf(`sudo sed -i "s/%s/%s/g" /etc/hosts`, hostName1_, newHostName)
			cmd1 := exec.Command("bash", "-c", com1+";"+com2+";hostname "+newHostName)
			cmd1.Run()
		}else{
			hostName_, _ := ioutil.ReadFile("/etc/hostname")
			hostName1_ := strings.Trim(string(hostName_), "\n")
			hostName1_ = strings.Trim(string(hostName1_), " ")
			com1 := fmt.Sprintf(`sed -i "s/%s/%s/g" /etc/hostname`, hostName1_, newHostName)
			com2 := fmt.Sprintf(`sed -i "s/%s/%s/g" /etc/hosts`, hostName1_, newHostName)
			cmd1 := exec.Command("bash", "-c", com1+";"+com2+";hostname "+newHostName)
			cmd1.Run()
		}

	} else if strings.ToLower(os_) == "debian" {
		hostName_, _ := ioutil.ReadFile("/etc/hostname")
		hostName1_ := strings.Trim(string(hostName_), "\n")
		hostName1_ = strings.Trim(string(hostName1_), " ")
		com1 := fmt.Sprintf(`sed -i "s/%s/%s/g" /etc/hostname`, hostName1_, newHostName)
		com2 := fmt.Sprintf(`sed -i "s/%s/%s/g" /etc/hosts`, hostName1_, newHostName)
		cmd1 := exec.Command("bash", "-c", com1+";"+com2+";hostname "+newHostName)
		cmd1.Run()
	}else if strings.ToLower(os_) == "fedora" {
		hostName_, _ := ioutil.ReadFile("/etc/hostname")
		hostName1_ := strings.Trim(string(hostName_), "\n")
		hostName1_ = strings.Trim(string(hostName1_), " ")
		if len(strings.Trim(hostName1_,""))<2{
			network_, _ := ioutil.ReadFile("/etc/sysconfig/network")
			osRegexp_ := regexp.MustCompile("[^_]HOSTNAME=\"?(.*)\"?")
			hostName := osRegexp_.FindSubmatch(network_)
			if len(hostName)>0{
				hostName_ := strings.Trim(string(hostName[1]), " ")
				hostName1_ = strings.Trim(string(hostName_), "\n")
			}
		}
		com1 := fmt.Sprintf(`sed -i "s/%s/%s/g" /etc/hostname`, hostName1_, newHostName)
		com2 := fmt.Sprintf(`sed -i "s/%s/%s/g" /etc/hosts`, hostName1_, newHostName)
		cmd1 := exec.Command("bash", "-c", com1+";"+com2+";hostname "+newHostName)
		cmd1.Run()
	} else {
		hostName_, _ := ioutil.ReadFile("/etc/hostname")
		hostName1_ := strings.Trim(string(hostName_), "\n")
		hostName1_ = strings.Trim(string(hostName1_), " ")
		com1 := fmt.Sprintf(`sudo sed -i "s/%s/%s/g" /etc/hostname`, hostName1_, newHostName)
		com2 := fmt.Sprintf(`sudo sed -i "s/%s/%s/g" /etc/hosts`, hostName1_, newHostName)
		cmd1 := exec.Command("bash", "-c", com1+";"+com2+";sudo hostname "+newHostName)
		cmd1.Run()
	}
}

func OsAndVer() (string, string) {
	var osStr, verStr string
	if IsFileExist("/etc/os-release") {
		release, _ := ioutil.ReadFile("/etc/os-release")
		osRegexp := regexp.MustCompile("[^_]ID=\"?(.*)\"?")
		os_ := osRegexp.FindSubmatch(release)
		verRegexp := regexp.MustCompile("VERSION_ID=\"(.*)\"")
		ver := verRegexp.FindSubmatch(release)
		if len(os_) > 0 {
			osStr = string(os_[1])
		}
		if len(ver) > 0 {
			verStr = string(ver[1])
		}
	} else if IsFileExist("/etc/centos-release") {
		release, _ := ioutil.ReadFile("/etc/centos-release")
		osRegexp := regexp.MustCompile("(CentOS)")
		releaseInfo := osRegexp.FindSubmatch(release)
		osStr = string(releaseInfo[1])
	} else {
		return "", ""
	}
	return strings.Trim(osStr,`"`),strings.Trim(verStr,`"`)
}

func InterfacesTag(name string) string {
	name = strings.Replace(name,"\n","",1)
	name = strings.Trim(name,"")
	tag, err := net.Interfaces()
	if err!=nil{
		return "eth0"
	}
	for k, t := range tag {
		if  k==0{
			continue
		}
		if  strings.EqualFold(t.HardwareAddr.String(),name){
			return t.Name
		}
		if strings.EqualFold(t.Name,name){
			return t.Name
		}
	}
	return "eth0"
}

func IsFileExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func MASK2CIDR(mask string) string {
	masks := strings.Split(mask, ".")
	mask1, _ := strconv.ParseInt(masks[0], 10, 64)
	mask2, _ := strconv.ParseInt(masks[1], 10, 64)
	mask3, _ := strconv.ParseInt(masks[2], 10, 64)
	mask4, _ := strconv.ParseInt(masks[3], 10, 64)
	ones, _ := net.IPv4Mask(byte(mask1), byte(mask2), byte(mask3), byte(mask4)).Size()
	return strconv.FormatInt(int64(ones), 10)
}

