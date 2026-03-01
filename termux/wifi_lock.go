package termux

import (
	"ZShadowRotator/utils"
        "fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
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
	
	return strings.TrimSpace(string(output)) == "Android"
}

func CheckDependencies() bool {
	// Deteksi environment dulu
	env := utils.DetectEnvironment()
	env.PrintStatus()
	
	// Dependencies yang WAJIB (Tor + Go)
	essential := []string{"tor", "go"}
	
	// Dependencies yang OPTIONAL (tergantung mode)
	optional := []string{"iptables", "tsu", "sudo", "git"}
	
	allEssentialOK := true
	
	// Check essential
	for _, dep := range essential {
		cmd := exec.Command("which", dep)
		if err := cmd.Run(); err != nil {
			utils.PrintError("Missing ESSENTIAL dependency: " + dep)
			utils.PrintInfo("Install with: pkg install " + dep)
			allEssentialOK = false
		} else {
			utils.PrintSuccess("Found: " + dep)
		}
	}
	
	// Check optional dengan info sesuai mode
	for _, dep := range optional {
		cmd := exec.Command("which", dep)
		if err := cmd.Run(); err != nil {
			switch dep {
			case "iptables":
				if env.RecommendMode == "FULL (Termux + Root)" || 
				   env.RecommendMode == "FULL (Linux + Sudo)" {
					utils.PrintError("Missing REQUIRED for FULL mode: " + dep)
				} else {
					utils.PrintWarning("Optional dependency missing: " + dep + " (not needed in current mode)")
				}
			case "tsu":
				if env.IsTermux && env.RecommendMode == "FULL (Termux + Root)" {
					utils.PrintWarning("Install tsu for better root: pkg install tsu")
				}
			default:
				utils.PrintDebug("Optional: " + dep + " not found")
			}
		} else {
			utils.PrintDebug("Found optional: " + dep)
		}
	}
	
	return allEssentialOK
}

func SetupMenu() {
	utils.PrintHeader("🔧 TERMUX SETUP")
	
	env := utils.DetectEnvironment()
	env.PrintStatus()
	
	fmt.Println()
	utils.PrintInfo("1. Install essential dependencies")
	utils.PrintInfo("2. Setup Tor configuration")
	utils.PrintInfo("3. Setup network (adaptive)")
	utils.PrintInfo("4. Build project")
	utils.PrintInfo("5. Run with optimal mode")
	
	fmt.Print("\n👉 Choose: ")
	var choice int
	fmt.Scanln(&choice)
	
	switch choice {
	case 1:
		exec.Command("pkg", "install", "-y", "tor", "golang", "git").Run()
		if env.IsTermux && env.HasRoot {
			exec.Command("pkg", "install", "-y", "iptables").Run()
		}
	case 2:
		exec.Command("pkg", "install", "-y", "tor").Run()
		os.MkdirAll("/data/data/com.termux/files/home/.tor", 0700)
	case 3:
		utils.PrintInfo("Network will be configured adaptively when starting engine")
	case 4:
		exec.Command("go", "mod", "tidy").Run()
		exec.Command("go", "build", "-o", "zshadow", "main.go").Run()
	case 5:
		utils.PrintInfo("Run with: ./zshadow")
		utils.PrintInfo("Then choose option 1")
	}
}
