// Package iland allows managing iland VMs.
// Copyright 2016 iland internet Solutions. All rights reserved.
package iland

import (
	"encoding/json"
	"fmt"

	iland "github.com/ilanddev/golang-sdk"
)

func (vm *VM) getIlandClient() *iland.Client {
	if vm.Client != nil {
		return vm.Client
	}
	vm.Client = iland.NewClient(&vm.Config)
	return vm.Client
}

func (vm *VM) updateInfo() error {
	if vm.Details.UUID == "" {
		return fmt.Errorf("Need a valid UUID to retrieve virtual machine info.")
	}
	updated, err := vm.getIlandClient().Get("/vm/" + vm.Details.UUID)
	if err != nil {
		return err
	}
	var d Details
	err = json.Unmarshal([]byte(updated), &d)
	if err != nil {
		return fmt.Errorf("Updating virtual machine %s: %s", vm.Details.UUID, err)
	}
	vm.Details = &d
	return nil
}
