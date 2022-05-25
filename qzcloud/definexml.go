package qzcloud

func GetCreateDomainXml() string {
	xml := `
		<domain type='kvm'>
		  <name>{{.host_name}}</name>
		  <uuid>{{.host_uuid}}</uuid>
		  <memory unit='MiB'>{{.max_ram}}</memory>
		  <currentMemory unit='MiB'>{{.min_ram}}</currentMemory>
		  <vcpu placement='static' current='{{.cpu}}'>{{.cpu}}</vcpu>
          <cpu mode='{{.cpu_model}}' />
		  <os>
			<type arch='{{.arch}}' machine='pc-i440fx-rhel7.0.0'>hvm</type>
			<bootmenu enable='yes'/>
		  </os>
		  <features>
			<acpi/>
			<apic/>
		  </features>
		  <clock offset='{{.clock}}'>
			<timer name='rtc' tickpolicy='catchup'/>
			<timer name='pit' tickpolicy='delay'/>
			<timer name='hpet' present='no'/>
		  </clock>
		  <on_poweroff>destroy</on_poweroff>
		  <on_reboot>restart</on_reboot>
		  <on_crash>destroy</on_crash>
		  <pm>
			<suspend-to-mem enabled='no'/>
			<suspend-to-disk enabled='no'/>
		  </pm>
		  <devices>
			<emulator>/usr/libexec/qemu-kvm</emulator>
			<disk type='file' device='disk'>
			  <driver name='qemu' type='qcow2' cache='none'/>
			  <source file='{{.os_path}}'/>
			  <target dev='vda' bus='virtio'/>
			  <boot order='1'/>
				 <iotune>
                	<read_bytes_sec>{{.os_read}}</read_bytes_sec>
			  	    <write_bytes_sec>{{.os_write}}</write_bytes_sec>
			   	    <read_iops_sec>{{.os_iops}}</read_iops_sec>
			  	    <write_iops_sec>{{.os_iops}}</write_iops_sec>
			    </iotune>
			</disk>
			<disk type='file' device='disk'>
			  <driver name='qemu' type='qcow2' cache='none'/>
			  <source file='{{.data_path}}'/>
			  <target dev='vdb' bus='virtio'/>
               <iotune>
			   	 <read_bytes_sec>{{.data_read}}</read_bytes_sec>
			     <write_bytes_sec>{{.data_write}}</write_bytes_sec>
			     <read_iops_sec>{{.data_iops}}</read_iops_sec>
			     <write_iops_sec>{{.data_iops}}</write_iops_sec>
             </iotune>
			</disk>
			<disk type='file' device='cdrom'>
			  <driver name='qemu' type='raw'/>
			  <source file='{{.cdrom}}'/>
			  <target dev='hdc' bus='ide'/>
			  <readonly/>
			  <boot order='2'/>
			  <address type='drive' controller='0' bus='1' target='0' unit='0'/>
			</disk>
			<controller type='usb' index='0' model='ich9-ehci1'>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x05' function='0x7'/>
			</controller>
			<controller type='usb' index='0' model='ich9-uhci1'>
			  <master startport='0'/>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x05' function='0x0' multifunction='on'/>
			</controller>
			<controller type='usb' index='0' model='ich9-uhci2'>
			  <master startport='2'/>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x05' function='0x1'/>
			</controller>
			<controller type='usb' index='0' model='ich9-uhci3'>
			  <master startport='4'/>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x05' function='0x2'/>
			</controller>
			<controller type='pci' index='0' model='pci-root'/>
			<controller type='ide' index='0'>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x01' function='0x1'/>
			</controller>
			<controller type='virtio-serial' index='0'>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x06' function='0x0'/>
			</controller>
			<interface type='bridge'>
			  <mac address='{{.mac}}'/>
			  <source bridge='br0'/>
              <target dev='{{.host_name}}'/>
			  <model type='virtio'/>
              <filterref filter='{{.host_name}}_public_firewall'>
                {{.filterip|unescaped}}
              </filterref>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x0'/>
              <bandwidth>
			     <inbound average='{{.bandwidth_in}}' peak='{{.bandwidth_in}}'   burst='{{.bandwidth_in}}'/>
			     <outbound average='{{.bandwidth_out}}' peak='{{.bandwidth_out}}' burst='{{.bandwidth_out}}'/>
               </bandwidth>
			</interface>
		   <interface type='bridge'>
              <mac address='{{.mac1}}'/>
			  <source bridge='br0'/>
			  <model type='virtio'/>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x04' function='0x0'/>
			</interface>

			<serial type='pty'>
			  <target type='isa-serial' port='0'>
				<model name='isa-serial'/>
			  </target>
			</serial>
			<console type='pty'>
			  <target type='serial' port='0'/>
			</console>
            <channel type='unix'>
               <source mode='bind'/>
               <target type='virtio' name='org.qemu.guest_agent.0' state='connected'/>
               <address type='virtio-serial' controller='0' bus='0' port='1'/>
            </channel>
			<input type='mouse' bus='ps2'/>
			<input type='keyboard' bus='ps2'/>
            <input type='tablet' bus='usb'/>
			<graphics type='vnc' port='{{.vnc_port}}' autoport='no' keymap='en-us' listen='0.0.0.0'  passwd='{{.vnc_password}}'>
			  <listen type="address" address="0.0.0.0"/>
			</graphics>
        	<sound model='ich6'>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x09' function='0x0'/>
			</sound>
			<video>
			  <model type='qxl' ram='65536' vram='65536' vgamem='16384' heads='1' primary='yes'/>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x02' function='0x0'/>
			</video>
			<redirdev bus='usb' type='spicevmc'>
			  <address type='usb' bus='0' port='1'/>
			</redirdev>
			<redirdev bus='usb' type='spicevmc'>
			  <address type='usb' bus='0' port='2'/>
			</redirdev>
			<memballoon model='virtio'>
			  <address type='pci' domain='0x0000' bus='0x00' slot='0x07' function='0x0'/>
			</memballoon>
		  </devices>
		</domain>`
	return xml
	//cpu model host-model  host-passthrough
	//限制ip策略模板
//<filterref filter='clean-traffic'>
//    <parameter name='IP' value='{{.Ip}}'/>
	//<parameter name='IP' value='{{.Ip}}'/>
	//<parameter name='IP' value='{{.Ip}}'/>
//</filterref>
}

func GetCreateVolumeXml() string {
	xmldesc := `
	<volume>
	 <name>{{.name}}</name>
	 <capacity unit="G">{{.capacity}}</capacity>
	 <target>
		<path>{{.path}}/{{.name}}</path>
		<format type="qcow2"/>
    </target>
	</volume>`
	//<allocation unit="bytes">{{.allocation}}</allocation> qcow2 raw
	return xmldesc
}

func GetCreatePoolXml() string {
	xmldesc := `
	<pool type='dir'>
		<name>{{.name}}</name>
		<target>
			<path>{{.path}}</path>
		</target>
	</pool>`
	return xmldesc
}

func GetSnapshotXml()string{
	xml:=` 
<domainsnapshot>
	<name>{{.name}}</name>
	 <memory snapshot='internal'/>
	  <disks>
		<disk name='vda' snapshot='internal'/>
	  </disks>
 </domainsnapshot>
`
	return  xml
}
