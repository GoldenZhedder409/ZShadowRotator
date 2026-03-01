package termux

import (
	"ZShadowRotator/utils"
	"os"
	"os/exec"
	"runtime"
)

type WifiLock struct {
	enabled bool
}

func NewWifiLock() *WifiLock {
	return &WifiLock{}
}

func (wl *WifiLock) Acquire() error {
	if runtime.GOOS != "android" {
		return nil
	}
	
	utils.PrintInfo("Acquiring WiFi lock...")
	
	// Try using Termux:API
	cmd := exec.Command("termux-wifi-enable", "true")
	if err := cmd.Run(); err == nil {
		wl.enabled = true
		utils.PrintSuccess("WiFi lock acquired")
		return nil
	}
	
	// Alternative: use wakelock
	cmd = exec.Command("termux-wake-lock")
	if err := cmd.Run(); err == nil {
		wl.enabled = true
		utils.PrintSuccess("Wake lock acquired")
		return nil
	}
	
	utils.PrintWarning("Could not acquire WiFi lock (install Termux:API)")
	return nil
}

func (wl *WifiLock) Release() {
	if !wl.enabled || runtime.GOOS != "android" {
		return
	}
	
	exec.Command("termux-wifi-enable", "false").Run()
	exec.Command("termux-wake-unlock").Run()
	utils.PrintInfo("WiFi lock released")
}

func CheckTermux() bool {
	if runtime.GOOS != "android" {
		return false
	}
	
	// Check if running in Termux
	cmd := exec.Command("uname", "-o")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return string(output) == "Android\n"
}

func CheckDependencies() bool {
	deps := []string{"tor", "iptables", "go"}
	
	for _, dep := range deps {
		cmd := exec.Command("which", dep)
		if err := cmd.Run(); err != nil {
			utils.PrintError("Missing dependency: " + dep)
			return false
		}
	}
	
	return true
}

func SetupMenu() {
	utils.PrintHeader("🔧 TERMUX SETUP")
	
	if !CheckTermux() {
		utils.PrintWarning("Not running in Termux")
		return
	}
	
	utils.PrintInfo("1. Install dependencies")
	utils.PrintInfo("2. Setup Tor")
	utils.PrintInfo("3. Setup iptables")
	utils.PrintInfo("4. Build project")
	utils.PrintInfo("5. Run with root")
	
	// Execute setup.sh
	cmd := exec.Command("bash", "termux/setup.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
