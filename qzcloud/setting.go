package qzcloud

import "github.com/libvirt/libvirt-go"

var Conn *libvirt.Connect
var ConnError error
func Setup(){
	Conn, ConnError = libvirt.NewConnect("qemu:///system")
}
