package main

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

type Status struct {
	LocalSuspend    bool
	LocalUnplugged  bool
	RemoteSuspend   bool
	RemoteUnplugged bool
	LeftSide        bool
	RightSide       bool
}

func WaitForRemotePlug(dev *gousb.Device) {
	fmt.Print("Waiting for remote plug...")

	status := Status{RemoteUnplugged: true}
	var err error

	for status.RemoteUnplugged {
		status, err = DeviceStatus(dev)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second)
	}
	fmt.Println("Detected remote plug")
}

func DeviceStatus(dev *gousb.Device) (Status, error) {
	flags := make([]byte, 2)

	n, err := dev.Control(0b1100_0000, 0b1111_1011, 0, 0, flags)
	if err != nil {
		return Status{}, err
	}
	if n != 2 {
		return Status{}, fmt.Errorf("control request expected 2, got %d", n)
	}
	fmt.Printf("0b%08b, 0b%08b\n", flags[0], flags[1])
	return Status{
		LocalSuspend:    flags[0]&0x01 != 0,
		LocalUnplugged:  flags[0]&0x02 != 0,
		RemoteSuspend:   flags[1]&0x01 != 0,
		RemoteUnplugged: flags[1]&0x02 != 0,
		LeftSide:        flags[0]&0x08 != 0,
		RightSide:       flags[1]&0x08 != 0,
	}, nil
}
