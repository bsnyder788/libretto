// Package iland allows managing iland VMs.
// Copyright 2016 iland Internet Solutions. All rights reserved.
package iland

import (
	"encoding/json"
	"net"

	"github.com/apcera/libretto/ssh"
	"github.com/apcera/libretto/util"
	iland "github.com/ilanddev/golang-sdk"
)

// Details contains details of an iland cloud VM.
type Details struct {
	UUID            string `json:"uuid"`
	Name            string `json:"name"`
	State           string `json:"status"`
	Description     string `json:"description"`
	CPUCount        int    `json:"cpus_number"`
	MemorySize      int64  `json:"memory_size"`
	OperatingSystem string `json:"os"`
	Deleted         bool   `json:"deleted"`
	VappUUID        string `json:"vapp_uuid"`
}

// Template to use when copying (provisioning) a new VM based off an existing VM.
type Template struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	IPAddressMode      string `json:"ip_address_mode"`
	NetworkUUID        string `json:"network_uuid"`
	VappTemplateUUID   string `json:"vapp_template_uuid"`
	VappTemplateName   string `json:"vapp_template_name"`
	VMTemplateUUID     string `json:"vm_template_uuid"`
	IPAddress          string `json:"ip_address"`
	StorageProfileUUID string `json:"storage_profile_uuid"`
}

// VM iland Virtual Machine.
type VM struct {
	Config iland.Config // iland client config

	Details     *Details
	Credentials ssh.Credentials // SSH credentials required to connect to machine

	Template []Template //template parameters to use when creating a new virtual machine
	Client   *iland.Client
}

// IP address
type IP struct {
	IP string `json:"ip_addr"`
}

// GetName of the VM.
func (vm *VM) GetName() string {
	err := vm.updateInfo()
	if err != nil {
		return ""
	}
	return vm.Details.Name
}

//Provision clones this VM.
func (vm *VM) Provision() error {
	err := vm.updateInfo()
	if err != nil {
		return err
	}
	template, err := json.Marshal(vm.Template)
	if err != nil {
		return err
	}
	_, err = vm.getIlandClient().Post("/vapp/"+vm.Details.VappUUID+"/vms", string(template))
	return err
}

// Destroy the VM.
func (vm *VM) Destroy() error {
	_, err := vm.getIlandClient().Delete("/vm/" + vm.Details.UUID)
	return err
}

// GetState gets the state of the VM through the iland API.
func (vm *VM) GetState() (string, error) {
	err := vm.updateInfo()
	if err != nil {
		return "", err
	}
	return vm.Details.State, nil
}

//GetIPs associated with a VM.
func (vm *VM) GetIPs() ([]net.IP, error) {
	vnics, err := vm.getIlandClient().Get("/vm/" + vm.Details.UUID + "/vnics")
	if err != nil {
		return nil, err
	}
	// Parse IPs from the VNIC array
	var ipList []IP
	var ips []net.IP
	err = json.Unmarshal([]byte(vnics), &ipList)
	if err != nil {
		return nil, err
	}
	for _, ip := range ipList {
		//try to parse IP into net.IP
		parsed := net.ParseIP(ip.IP)
		if parsed != nil {
			ips = append(ips, parsed)
		}
	}
	return ips, nil
}

//Suspend a VM.
func (vm *VM) Suspend() error {
	_, err := vm.getIlandClient().Post("/vm/"+vm.Details.UUID+"/suspend", "")
	return err
}

// Resume a VM.
func (vm *VM) Resume() error {
	return vm.Start()
}

// Halt the VM.
func (vm *VM) Halt() error {
	_, err := vm.getIlandClient().Post("/vm/"+vm.Details.UUID+"/poweroff", "")
	return err
}

// Start the VM.
func (vm *VM) Start() error {
	_, err := vm.getIlandClient().Post("/vm/"+vm.Details.UUID+"/poweron", "")
	return err
}

//GetSSH for a VM.
func (vm *VM) GetSSH(options ssh.Options) (ssh.Client, error) {
	ips, err := util.GetVMIPs(vm, options)
	if err != nil {
		return nil, err
	}
	client := ssh.SSHClient{Creds: &vm.Credentials, IP: ips[0], Port: 22, Options: options}
	return &client, nil
}
