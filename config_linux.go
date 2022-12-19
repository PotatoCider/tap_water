package main

import "github.com/songgao/water"

func GetConfig(name string, use_tun bool, persist bool, multiqueue bool, ip_address string) water.Config {
	cfg := water.Config{
		DeviceType: water.TAP,
	}
	if use_tun {
		cfg.DeviceType = water.TUN
	}
	cfg.Name = name
	cfg.MultiQueue = multiqueue
	cfg.Persist = persist

	return cfg
}
