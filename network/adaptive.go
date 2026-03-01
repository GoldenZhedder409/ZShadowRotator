package network

import (
	"ZShadowRotator/utils"
	"fmt"
	"os/exec"
	"strings"
)

type AdaptiveNetwork struct {
	env *utils.Environment
}

func NewAdaptiveNetwork(env *utils.Environment) *AdaptiveNetwork {
	return &AdaptiveNetwork{env: env}
}

func (an *AdaptiveNetwork) SetupRouting(port int) error {
	utils.PrintInfo("Setting up adaptive routing for " + an.env.OS)
	
	switch {
	case an.env.IsTermux:
		return an.setupTermux(port)
	case an.env.OS == "linux":
		return an.setupLinux(port)
	case an.env.OS == "darwin":
		return an.setupMacOS(port)
	case an.env.OS == "windows":
		return an.setupWindows(port)
	default:
		utils.PrintWarning("Unknown OS, falling back to SOCKS5 only")
		return nil
	}
}

func (an *AdaptiveNetwork) setupTermux(port int) error {
	utils.PrintInfo("📱 Termux detected")
	
	// Cek root/tsu
	if an.env.HasTsu {
		utils.PrintInfo("tsu available! Trying root mode...")
		return an.setupTermuxWithTsu(port)
	}
	
	if an.env.HasRoot {
		utils.PrintInfo("Root access available!")
		return an.setupTermuxWithRoot(port)
	}
	
	// Non-root mode
	utils.PrintWarning("Running Termux without root")
	utils.PrintInfo("SOCKS5 proxy akan berjalan di 127.0.0.1:" + fmt.Sprint(port))
	utils.PrintInfo("Setting manual di browser/app:")
	utils.PrintInfo("  • Host: 127.0.0.1")
	utils.PrintInfo("  • Port: " + fmt.Sprint(port))
	utils.PrintInfo("  • Type: SOCKS5")
	return nil
}

func (an *AdaptiveNetwork) setupTermuxWithTsu(port int) error {
	utils.PrintSuccess("tsu detected! Setting up iptables with tsu...")
	
	commands := [][]string{
		{"tsu", "sysctl", "-w", "net.ipv4.ip_forward=1"},
		{"tsu", "iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "--to-ports", "9040"},
		{"tsu", "iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "9040"},
	}
	
	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			utils.PrintWarning("Command failed: " + strings.Join(cmd, " "))
		} else {
			utils.PrintDebug("Success: " + strings.Join(cmd, " "))
		}
	}
	
	utils.PrintSuccess("✅ Full routing mode active (with tsu)")
	return nil
}

func (an *AdaptiveNetwork) setupTermuxWithRoot(port int) error {
	utils.PrintSuccess("Root detected! Setting up iptables...")
	
	commands := [][]string{
		{"su", "-c", "sysctl -w net.ipv4.ip_forward=1"},
		{"su", "-c", "iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 9040"},
		{"su", "-c", "iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports 9040"},
	}
	
	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			utils.PrintWarning("Command failed: " + strings.Join(cmd, " "))
		}
	}
	
	utils.PrintSuccess("✅ Full routing mode active (with root)")
	return nil
}

func (an *AdaptiveNetwork) setupLinux(port int) error {
	utils.PrintInfo("🐧 Linux detected")
	
	if an.env.HasSudo {
		utils.PrintInfo("sudo available! Setting up iptables...")
		
		commands := [][]string{
			{"sudo", "sysctl", "-w", "net.ipv4.ip_forward=1"},
			{"sudo", "iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "--to-ports", "9040"},
			{"sudo", "iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "9040"},
		}
		
		for _, cmd := range commands {
			exec.Command(cmd[0], cmd[1:]...).Run()
		}
		
		utils.PrintSuccess("✅ Full routing mode active (with sudo)")
		return nil
	}
	
	utils.PrintWarning("No sudo access, falling back to SOCKS5 only")
	utils.PrintInfo("Set browser proxy manually to 127.0.0.1:" + fmt.Sprint(port))
	return nil
}

func (an *AdaptiveNetwork) setupMacOS(port int) error {
	utils.PrintInfo("🍎 macOS detected")
	
	if an.env.HasSudo {
		utils.PrintInfo("sudo available! Setting up pfctl...")
		
		// macOS pakai pf (packet filter) instead of iptables
		pfRules := fmt.Sprintf(`
rdr pass on lo0 inet proto tcp from any to any port 80 -> 127.0.0.1 port %d
rdr pass on lo0 inet proto tcp from any to any port 443 -> 127.0.0.1 port %d
`, port, port)
		
		cmd := exec.Command("sudo", "pfctl", "-f", "-")
		cmd.Stdin = strings.NewReader(pfRules)
		cmd.Run()
		
		exec.Command("sudo", "pfctl", "-e").Run()
		
		utils.PrintSuccess("✅ Full routing mode active (with pfctl)")
		return nil
	}
	
	utils.PrintWarning("No sudo access, falling back to SOCKS5 only")
	utils.PrintInfo("Set browser proxy manually to 127.0.0.1:" + fmt.Sprint(port))
	return nil
}

func (an *AdaptiveNetwork) setupWindows(port int) error {
	utils.PrintInfo("🪟 Windows detected")
	utils.PrintWarning("Windows doesn't support iptables")
	utils.PrintInfo("Using SOCKS5 proxy only")
	utils.PrintInfo("Set browser proxy manually to 127.0.0.1:" + fmt.Sprint(port))
	utils.PrintInfo("Or use: netsh winhttp set proxy 127.0.0.1:" + fmt.Sprint(port))
	return nil
}

func (an *AdaptiveNetwork) Cleanup() {
	utils.PrintInfo("Cleaning up network rules...")
	
	switch {
	case an.env.IsTermux && an.env.HasTsu:
		exec.Command("tsu", "iptables", "-t", "nat", "-F").Run()
	case an.env.IsTermux && an.env.HasRoot:
		exec.Command("su", "-c", "iptables -t nat -F").Run()
	case an.env.OS == "linux" && an.env.HasSudo:
		exec.Command("sudo", "iptables", "-t", "nat", "-F").Run()
	case an.env.OS == "darwin" && an.env.HasSudo:
		exec.Command("sudo", "pfctl", "-f", "/etc/pf.conf").Run()
		exec.Command("sudo", "pfctl", "-d").Run()
	}
}
