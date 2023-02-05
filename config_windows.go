package main

import "github.com/songgao/water"

func GetPlatformConfig(name string, persist bool, multiqueue bool, ip_address string, tuntaposx bool) (cfg water.Config) {
	cfg.ComponentID = "root\\tap0901"
	cfg.InterfaceName = name
	cfg.Network = ip_address

	return
}
