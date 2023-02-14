package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/gousb"
	"github.com/songgao/water"
)

func main() {
	var device_index int
	var iface_name string
	var ip_address string
	var gateway_address string
	var dns_address string
	var iface_up bool
	var persist bool
	var use_multiqueue bool
	var use_tun bool
	var wait_plug bool
	var dhcp bool
	var use_stream bool
	var tuntaposx bool
	flag.IntVar(&device_index, "i", 0, "device index (if there is a device vid:pid collision)")
	flag.StringVar(&iface_name, "I", "", "interface name (e.g. tap0)")
	flag.StringVar(&ip_address, "ip", "", "ip address in cidr notation (e.g. 192.168.10.2/24)")
	flag.StringVar(&gateway_address, "gw", "", "gateway address (e.g. 192.168.10.1)")
	flag.StringVar(&dns_address, "dns", "", "default DNS address (windows only) (e.g. 192.168.10.1)")
	flag.BoolVar(&iface_up, "up", false, "initial state of interface")
	flag.BoolVar(&persist, "persist", false, "Persist interface (linux only)")
	flag.BoolVar(&use_multiqueue, "multiqueue", false, "Multiqueue (linux only)")
	flag.BoolVar(&use_tun, "tun", false, "Use TUN instead of TAP adapter")
	flag.BoolVar(&wait_plug, "wait", false, "Wait for USB plug-in")
	flag.BoolVar(&dhcp, "dhcp", false, "Use DHCP")
	flag.BoolVar(&use_stream, "stream", false, "Use Buffering (doesn't work with plusb)")
	flag.BoolVar(&tuntaposx, "tuntaposx", false, "Use macos tuntaposx driver (required for TAP)")

	flag.Usage = func() {
		fmt.Printf("Usage: %s [flags...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// Initialize TAP virtual network interface
	cfg := GetPlatformConfig(iface_name, persist, use_multiqueue, ip_address, tuntaposx)
	if use_tun {
		cfg.DeviceType = water.TUN
	} else {
		cfg.DeviceType = water.TAP
	}

	tap_iface, err := water.New(cfg)
	if err != nil {
		log.Fatalf("water.New(%v): %s\n", cfg, err)
	}
	defer tap_iface.Close()
	log.Printf("Interface Name: %s\n", tap_iface.Name())

	i := 0
	for {
		// closure to defer cleanup
		func() {
			fmt.Printf("reset count %d\n", i)

			// Initialize PL25A1 Device
			ctx := gousb.NewContext()
			defer ctx.Close()

			devs := make([]*gousb.Device, 0)
			vid, pid := gousb.ID(0x067b), gousb.ID(0x25a1)

			fmt.Print("Waiting for USB local plug...")
			for len(devs) == 0 {
				devs, err = ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
					return desc.Vendor == vid && desc.Product == pid
				})
				if err != nil {
					log.Fatalf("OpenDevices(): %v", err)
				}
				time.Sleep(time.Second)
			}
			fmt.Println("Detected")

			for i, dev := range devs {
				defer dev.Close()
				fmt.Printf("%d. %s\n", i, dev)
			}

			dev := devs[device_index]

			if err := dev.SetAutoDetach(true); err != nil {
				log.Fatalf("dev.SetAutoDetach(true): %v", err)
			}

			var dev_iface *gousb.Interface
			if runtime.GOOS == "darwin" {
				cfg, err := dev.Config(1)
				if err != nil {
					log.Fatalf("%s.Config(1): %v", dev, err)
				}
				defer cfg.Close()
				dev_iface, err = cfg.Interface(0, 0)
				if err != nil {
					log.Fatalf("%s.Interface(0, 0): %v", cfg, err)
				}
				defer dev_iface.Close()
			} else {
				// Claim interface 0
				var done func()
				dev_iface, done, err = dev.DefaultInterface()
				if err != nil {
					log.Fatalf("%s.DefaultInterface(): %v", dev, err)
				}
				defer done()
			}

			// Wait for remote plugin
			// WaitForRemotePlug(dev)

			outEndpoint, err := dev_iface.OutEndpoint(0x02)
			if err != nil {
				log.Fatalf("%s.OutEndpoint(0x02): %v", dev_iface, err)
			}
			inEndpoint, err := dev_iface.InEndpoint(0x83)
			if err != nil {
				log.Fatalf("%s.InEndpoint(0x83): %v", dev_iface, err)
			}

			var usbReader io.Reader = inEndpoint
			var usbWriter io.Writer = outEndpoint

			if use_stream {
				writeStream, err := outEndpoint.NewStream(2048, 4)
				if err != nil {
					log.Fatalf("%s.NewStream(2048, 16): %v", outEndpoint, err)
				}
				defer writeStream.Close()

				readStream, err := inEndpoint.NewStream(2048, 4)
				if err != nil {
					log.Fatalf("%s.NewStream(2048, 16): %v", inEndpoint, err)
				}
				defer readStream.Close()

				usbReader = readStream
				usbWriter = writeStream
			}

			go func() {
				time.Sleep(time.Second)
				switch runtime.GOOS {
				case "linux":
					if ip_address != "" {
						if err := run_command("ip", "addr", "add", ip_address, "dev", tap_iface.Name()); err != nil {
							fmt.Printf("set ip: %s\n", err)
						}
					}
					if iface_up {
						if err := run_command("ip", "link", "set", tap_iface.Name(), "up"); err != nil {
							fmt.Printf("iface_up: %s\n", err)
						}
					}
					if gateway_address != "" {
						if err := run_command("ip", "route", "add", "default", "via", gateway_address, "dev", tap_iface.Name()); err != nil {
							fmt.Printf("set gateway: %s\n", err)
						}
					}
				case "windows":
					name := tap_iface.Name()
					if dhcp {
						if err := run_command("netsh", "interface", "ip", "set", "address", name, "dhcp"); err != nil {
							// panic(err)
							fmt.Printf("dhcp: %s\n", err)
						}
						go func() {
							time.Sleep(time.Second)
							// TODO: clean this nasty go-routine
							if err := run_command("ipconfig", "-renew"); err != nil {
								fmt.Printf("ipconfig -renew: %s\n", err)
							}
						}()
					}
					if ip_address != "" {
						if err := run_command("netsh", "interface", "ip", "set", "address", name, "static", ip_address, "gateway="+gateway_address); err != nil {
							fmt.Printf("set ip: %s\n", err)
						}
					}
					if iface_up {
						if err := run_command("netsh", "interface", "set", "interface", name, "enable"); err != nil {
							fmt.Printf("iface_up: %s\n", err)
						}
					}
					if dns_address != "" {
						if err := run_command("netsh", "interface", "ip", "set", "dnsservers", name, "static", dns_address, "primary"); err != nil {
							fmt.Printf("set dns: %s\n", err)
						}
					}
				case "darwin":
					if dhcp {
						if err := run_command("ipconfig", "set", tap_iface.Name(), "DHCP"); err != nil {
							fmt.Printf("dhcp: %s\n", err)
						}
					}
					if ip_address != "" {
						if err := run_command("ifconfig", tap_iface.Name(), "inet", ip_address); err != nil {
							fmt.Printf("set ip: %s\n", err)
						}
					}
				}
			}()

			var wg sync.WaitGroup
			finish := false
			wg.Add(1)

			go func() {
				for {
					if _, err := io.Copy(usbWriter, tap_iface); err != nil {
						log.Println("writeStream err:", err)
						if finish || err == io.ErrClosedPipe || err == gousb.ErrorIO || err == gousb.TransferNoDevice || err == gousb.ErrorNoDevice {
							break
						}
					}
				}
				wg.Done()
			}()

			go func() {
				for {
					if _, err := io.Copy(tap_iface, usbReader); err != nil {
						log.Println("readStream err:", err)
						if finish || err == io.ErrClosedPipe || err == gousb.ErrorIO || err == gousb.TransferNoDevice || err == gousb.ErrorNoDevice {
							break
						}
					}
				}
				wg.Done()
			}()

			wg.Wait()

			// wait for 2nd go-routine to finish
			wg.Add(1)
			fmt.Println("exitting...")
			dev.Close()
			finish = true
			wg.Wait()

			time.Sleep(time.Second)
			i++
		}()
	}
}

func run_command(name string, args ...string) error {
	fmt.Println(">", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
