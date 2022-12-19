package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/google/gousb"
	"github.com/songgao/water"
)

func main() {
	var iface_name string
	var ip_address string
	var gateway_address string
	var dns_address string
	var iface_up bool
	var persist bool
	var use_multiqueue bool
	var use_tun bool
	flag.StringVar(&iface_name, "I", "", "interface name (e.g. tap0)")
	flag.StringVar(&ip_address, "ip", "", "ip address in cidr notation (e.g. 192.168.10.2/24)")
	flag.StringVar(&gateway_address, "gw", "", "gateway address (e.g. 192.168.10.1)")
	flag.StringVar(&dns_address, "dns", "", "default DNS address (windows only) (e.g. 192.168.10.1)")
	flag.BoolVar(&iface_up, "up", false, "initial state of interface")
	flag.BoolVar(&persist, "persist", false, "Persist interface (linux only)")
	flag.BoolVar(&use_multiqueue, "multiqueue", false, "Multiqueue (linux only)")
	flag.BoolVar(&use_tun, "tun", false, "Use TUN instead of TAP adapter")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [flags...] <index>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}
	index, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		log.Fatalf("Index parse error: %v", err)
	}

	// Initialize TAP virtual network interface
	cfg := GetConfig(iface_name, use_tun, persist, use_multiqueue, ip_address)
	tap_iface, err := water.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer tap_iface.Close()

	log.Printf("Interface Name: %s\n", tap_iface.Name())

	// Initialize PL25A1 Device
	ctx := gousb.NewContext()
	defer ctx.Close()

	vid, pid := gousb.ID(0x067b), gousb.ID(0x25a1)

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// this function is called for every device present.
		// Returning true means the device should be opened.
		return desc.Vendor == vid && desc.Product == pid
	})
	for _, d := range devs {
		defer d.Close()
	}
	if err != nil {
		log.Fatalf("OpenDevices(): %v", err)
	}
	if len(devs) == 0 {
		log.Fatalf("no devices found matching VID %s and PID %s", vid, pid)
	}

	for i, dev := range devs {
		fmt.Printf("%d. %s\n", i, dev)
	}

	dev := devs[index]

	if err := dev.SetAutoDetach(true); err != nil {
		log.Fatalf("dev.SetAutoDetach(true): %v", err)
	}

	dev_iface, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer done()

	outEP, err := dev_iface.OutEndpoint(0x02)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(0x02): %v", dev_iface, err)
	}
	inEP, err := dev_iface.InEndpoint(0x83)
	if err != nil {
		log.Fatalf("%s.InEndpoint(0x83): %v", dev_iface, err)
	}

	writeStream, err := outEP.NewStream(2048, 16)
	if err != nil {
		log.Fatalf("%s.NewStream(2048, 16): %v", outEP, err)
	}
	defer writeStream.Close()
	readStream, err := inEP.NewStream(2048, 16)
	if err != nil {
		log.Fatalf("%s.NewStream(2048, 16): %v", inEP, err)
	}
	defer readStream.Close()

	go func() {
		time.Sleep(time.Second)
		switch runtime.GOOS {
		case "linux":
			if ip_address != "" {
				if err := exec.Command("ip", "addr", "add", ip_address, "dev", tap_iface.Name()).Run(); err != nil {
					panic(err)
				}
			}
			if iface_up {
				if err := exec.Command("ip", "link", "set", tap_iface.Name(), "up").Run(); err != nil {
					panic(err)
				}
			}
			if gateway_address != "" {
				if err := exec.Command("ip", "route", "add", "default", "via", gateway_address, "dev", tap_iface.Name()).Run(); err != nil {
					panic(err)
				}
			}
		case "windows":
			name := tap_iface.Name()
			if ip_address != "" {
				if err := exec.Command("netsh", "interface", "ip", "set", "address", name, "static", ip_address, "gateway="+gateway_address).Run(); err != nil {
					panic(err)
				}
			}
			if iface_up {
				if err := exec.Command("netsh", "interface", "set", "interface", name, "enable").Run(); err != nil {
					panic(err)
				}
			}
			if dns_address != "" {
				if err := exec.Command("netsh", "interface", "ip", "set", "dnsservers", name, "static", dns_address, "primary").Run(); err != nil {
					panic(err)
				}
			}
		}
	}()

	go func() {
		for {
			if _, err := io.Copy(writeStream, tap_iface); err != nil {
				log.Println("writeStream err:", err)
			}
		}
	}()

	go func() {
		for {
			if _, err := io.Copy(tap_iface, readStream); err != nil {
				log.Println("readStream err:", err)
			}
		}
	}()

	select {}
}
