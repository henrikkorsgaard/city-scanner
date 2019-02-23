package config

import (
	"fmt"

	"../../../wifi"
)

func init() {
	fmt.Println("initalising config")
}

var (
	client *wifi.Client
)

func ConfigureInterfaces() (err error) {
	client, err = wifi.New()
	if err != nil {
		return
	}

	phys, err := client.PHYs()

	if err != nil {
		return
	}

	//We need two devices -- on for scanning (because it will cycle through the channels) and one for reporting signals to server
	if len(phys) < 2 {
		return fmt.Errorf("Could only detect %d physical wifi devices. Need at least 2 physical wifi devices to operate.", len(phys))
	}

	ifaces, err := client.Interfaces()
	if err != nil {
		return
	}

	connectedPhyIndex := -1
	scannerPhyIndex := -1
	var bss *wifi.BSS
	for _, iface := range ifaces {

		bss, _ = client.BSS(iface)
		if bss != nil {
			connectedPhyIndex = iface.PHY
			break
		}
	}

	for _, p := range phys {
		if p.Index != connectedPhyIndex && contains(p.SupportedIftypes, "monitor") {
			scannerPhyIndex = p.Index
		}

		if connectedPhyIndex == -1 && scannerPhyIndex != p.Index {
			connectedPhyIndex = p.Index
		}
	}

	if scannerPhyIndex == -1 {
		return fmt.Errorf("Unable to find interface supporting monitor mode")
	}

	if connectedPhyIndex == -1 {
		setupWifiConnection()
	}

	setupMonitorInterface(scannerPhyIndex)

	return
}

func setupWifiConnection() {
	fmt.Println("need to setup wifi")
}

func setupMonitorInterface(PHY int) {
	fmt.Println(PHY)

	err := client.CreateNewInterface(PHY, wifi.InterfaceTypeMonitor, "go-mon-hk")
	fmt.Println("create error")
	fmt.Println(err)
	//we need to remove all interfaces on
	//we need to setup a monitor interface ->
}

func contains(types []wifi.InterfaceType, t string) bool {
	for _, a := range types {
		if a.String() == t {
			return true
		}
	}
	return false
}
