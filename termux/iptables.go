package termux

import (
	"ZShadowRotator/utils"
	"os/exec"
	"runtime"
)

func SetupIptables(port int) error {
	if runtime.GOOS != "android" && runtime.GOOS != "linux" {
		return nil // Skip on non-Linux
	}
	
	utils.PrintInfo("Setting up iptables rules...")
	
	// Try with tsu first (Termux root)
	commands := [][]string{
		{"tsu", "iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "--to-ports", "9040"},
		{"tsu", "iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "9040"},
	}
	
	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			// Try without tsu
			if err2 := exec.Command("iptables", cmd[2:]...).Run(); err2 != nil {
				utils.PrintWarning("Failed to set iptables rules (may require root)")
			}
		}
	}
	
	return nil
}

func ClearIptables() {
	if runtime.GOOS != "android" && runtime.GOOS != "linux" {
		return
	}
	
	commands := [][]string{
		{"tsu", "iptables", "-t", "nat", "-D", "OUTPUT", "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "--to-ports", "9040"},
		{"tsu", "iptables", "-t", "nat", "-D", "OUTPUT", "-p", "tcp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "9040"},
	}
	
	for _, cmd := range commands {
		exec.Command(cmd[0], cmd[1:]...).Run()
	}
}
