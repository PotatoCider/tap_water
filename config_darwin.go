package main

import "github.com/songgao/water"

func GetPlatformConfig(name string, persist bool, multiqueue bool, ip_address string, tuntaposx bool) (cfg water.Config) {
	cfg.Name = name
	cfg.Driver = water.MacOSDriverSystem
	if tuntaposx {
		cfg.Driver = water.MacOSDriverTunTapOSX
	}

	return
}
