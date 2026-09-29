package main

import (
	"slices"
	"sync"
	"time"
)

// demoMode is set by --demo. Every backend call must honor it so the
// TUI never reads /etc/wireguard or talks to a live helper.
var demoMode bool

func enableDemoMode() { demoMode = true }

var demoMu sync.Mutex

// Fake tunnel names only. Never copy names from a real /etc/wireguard.
var demoConfigs = []string{"homelab", "office", "travel"}

var demoActive = []string{"homelab"}

var demoInfos = map[string]ConfigInfo{
	"homelab": {
		Address:    "10.8.0.2/24",
		DNS:        "1.1.1.1",
		Endpoint:   "vpn.example.net:51820",
		PeerKey:    "DemoPeerKeyNotARealKeyAAAAAAA=",
		AllowedIPs: []string{"10.8.0.0/24"},
	},
	"office": {
		Address:    "10.10.0.4/24",
		DNS:        "9.9.9.9",
		Endpoint:   "office.example.net:51820",
		PeerKey:    "DemoOfficePeerNotARealKeyBBBB=",
		AllowedIPs: []string{"10.10.0.0/24"},
	},
	"travel": {
		Address:    "10.20.0.8/24",
		DNS:        "1.1.1.1",
		Endpoint:   "travel.example.net:51820",
		PeerKey:    "DemoTravelPeerNotARealKeyCCCC=",
		AllowedIPs: []string{"0.0.0.0/0", "::/0"},
	},
}

func demoListConfigs() []string {
	return slices.Clone(demoConfigs)
}

func demoActiveVPNs() []string {
	demoMu.Lock()
	defer demoMu.Unlock()
	return slices.Clone(demoActive)
}

func demoConnect(name string) {
	time.Sleep(450 * time.Millisecond)
	demoMu.Lock()
	defer demoMu.Unlock()
	if !slices.Contains(demoActive, name) {
		demoActive = append(demoActive, name)
	}
}

func demoDisconnect(name string) {
	demoMu.Lock()
	defer demoMu.Unlock()
	out := demoActive[:0]
	for _, n := range demoActive {
		if n != name {
			out = append(out, n)
		}
	}
	demoActive = out
}

func demoVPNStatus(name string) VPNStatus {
	info := demoInfos[name]
	return VPNStatus{
		Interface:  name,
		Endpoint:   info.Endpoint,
		TransferRx: "12.40 MiB received",
		TransferTx: "3.10 MiB sent",
		Handshake:  "1 minute ago",
	}
}

func demoConfigInfo(name string) ConfigInfo {
	if info, ok := demoInfos[name]; ok {
		return info
	}
	return ConfigInfo{}
}
