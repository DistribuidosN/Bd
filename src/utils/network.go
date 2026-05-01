package utils

import (
	"net"
	"strings"
)

// GetLocalIP returns the local IP of this machine, ignoring virtual adapters
func GetLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	var fallbackIP string

	for _, iface := range interfaces {
		// Ignorar interfaces apagadas o loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Ignorar adaptadores virtuales comunes (Docker, WSL, Hyper-V, VMware, VirtualBox)
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "virtual") ||
			strings.Contains(name, "vethernet") ||
			strings.Contains(name, "wsl") ||
			strings.Contains(name, "vmware") ||
			strings.Contains(name, "vbox") ||
			strings.Contains(name, "docker") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip := ipnet.IP.String()
					// Si encontramos una IP local típica (192.168.x.x, 10.x.x.x, 172.16-31.x.x) retornamos esta
					if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
						return ip
					}
					// Guardamos la primera IPv4 que veamos como respaldo
					if fallbackIP == "" {
						fallbackIP = ip
					}
				}
			}
		}
	}

	// Si no se encontró una IP de red privada pero sí hay otra IPv4 válida
	if fallbackIP != "" {
		return fallbackIP
	}

	return "127.0.0.1"
}
