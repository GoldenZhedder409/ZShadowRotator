package utils

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Environment struct {
	OS           string
	IsTermux     bool
	HasRoot      bool
	HasTsu       bool
	HasSudo      bool
	IptablesAvail bool
	RecommendMode string
}

func DetectEnvironment() *Environment {
	env := &Environment{
		OS: runtime.GOOS,
	}
	
	// Deteksi Termux
	env.IsTermux = checkTermux()
	
	// Deteksi Root Access
	env.HasRoot = checkRoot()
	env.HasTsu = checkTsu()
	env.HasSudo = checkSudo()
	
	// Deteksi iptables
	env.IptablesAvail = checkIptables()
	
	// Rekomendasi mode
	env.RecommendMode = env.getRecommendMode()
	
	return env
}

func checkTermux() bool {
	// Cek file khas Termux
	if _, err := os.Stat("/data/data/com.termux"); err == nil {
		return true
	}
	
	// Cek environment variable
	if os.Getenv("TERMUX_VERSION") != "" {
		return true
	}
	
	// Cek uname
	cmd := exec.Command("uname", "-o")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output)) == "Android"
	}
	
	return false
}

func checkRoot() bool {
	// Coba akses file system root
	cmd := exec.Command("su", "-c", "echo test")
	err := cmd.Run()
	return err == nil
}

func checkTsu() bool {
	cmd := exec.Command("which", "tsu")
	err := cmd.Run()
	return err == nil
}

func checkSudo() bool {
	cmd := exec.Command("which", "sudo")
	if err := cmd.Run(); err != nil {
		return false
	}
	
	// Test sudo tanpa password (kadang bisa)
	cmd = exec.Command("sudo", "-n", "true")
	err := cmd.Run()
	return err == nil
}

func checkIptables() bool {
	cmd := exec.Command("which", "iptables")
	if err := cmd.Run(); err != nil {
		return false
	}
	
	// Test iptables (butuh root biasanya)
	if checkRoot() || checkTsu() || checkSudo() {
		return true
	}
	
	return false
}

func (e *Environment) getRecommendMode() string {
	switch {
	case e.IsTermux && (e.HasRoot || e.HasTsu):
		return "FULL (Termux + Root)"
	case e.IsTermux:
		return "SOCKS5 (Termux Non-Root)"
	case e.OS == "linux" && (e.HasRoot || e.HasSudo):
		return "FULL (Linux + Sudo)"
	case e.OS == "linux":
		return "SOCKS5 (Linux Non-Root)"
	case e.OS == "darwin" && e.HasSudo:
		return "FULL (macOS + PFCTL)"
	case e.OS == "darwin":
		return "SOCKS5 (macOS)"
	case e.OS == "windows":
		return "SOCKS5 (Windows)"
	default:
		return "SOCKS5 (Default)"
	}
}

func (e *Environment) PrintStatus() {
	PrintHeader("🌍 ENVIRONMENT DETECTION")
	
	PrintInfo("OS: " + e.OS)
	PrintInfo("Termux: " + boolToEmoji(e.IsTermux))
	PrintInfo("Root Access: " + boolToEmoji(e.HasRoot))
	PrintInfo("Tsu Available: " + boolToEmoji(e.HasTsu))
	PrintInfo("Sudo Available: " + boolToEmoji(e.HasSudo))
	PrintInfo("iptables: " + boolToEmoji(e.IptablesAvail))
	PrintSuccess("Recommended Mode: " + e.RecommendMode)
	
	if !e.IptablesAvail {
		PrintWarning("Running in SOCKS5-only mode")
		PrintInfo("Set browser proxy manually to 127.0.0.1:1080")
	}
	
	if e.IsTermux && !e.HasRoot && !e.HasTsu {
		PrintInfo("\n📱 Termux Non-Root Tips:")
		PrintInfo("  • Install tsu for root: pkg install tsu")
		PrintInfo("  • Or use manual proxy setting in browser")
	}
	
	if e.OS == "linux" && !e.HasSudo {
		PrintInfo("\n🐧 Linux Tips:")
		PrintInfo("  • Add user to sudoers for full mode")
		PrintInfo("  • Or use manual proxy setting")
	}
}

func boolToEmoji(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}
