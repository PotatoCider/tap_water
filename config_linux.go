package main

import "github.com/songgao/water"

func GetPlatformConfig(name string, persist bool, multiqueue bool, ip_address string, tuntaposx bool) (cfg water.Config) {
	cfg.Name = name
	cfg.MultiQueue = multiqueue
	cfg.Persist = persist

	return
}
