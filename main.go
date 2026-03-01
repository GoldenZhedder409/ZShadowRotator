package main

import (
	"ZShadowRotator/config"
	"ZShadowRotator/hybrid"
	"ZShadowRotator/termux"
	"ZShadowRotator/utils"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Clear screen dan welcome
	utils.ClearScreen()
	utils.PrintLogo()
	
	// Cek environment
	if !termux.CheckTermux() {
		utils.PrintWarning("Running outside Termux, some features may not work")
	} else {
		utils.PrintSuccess("Termux detected! Running in optimal environment")
	}

	// Setup signal handling untuk clean exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.PrintError("Failed to load config: " + err.Error())
		cfg = config.DefaultConfig()
		config.SaveConfig(cfg)
	}

	// Main menu loop
	for {
		utils.PrintHeader("ZSHADOWROTATOR v1.0 - ONION HYBRID FORTRESS")
		fmt.Println("  [1] 🚀 START HYBRID ENGINE")
		fmt.Println("  [2] ⚙️  CONFIGURATION")
		fmt.Println("  [3] 🌍 PROXY POOL MANAGER")
		fmt.Println("  [4] 📊 STATUS & STATS")
		fmt.Println("  [5] 🔧 TERMUX SETUP")
		fmt.Println("  [6] 📖 DOCUMENTATION")
		fmt.Println("  [7] ❌ EXIT")
		fmt.Println()
		fmt.Print("👉 Choose: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			startHybridEngine(cfg)
		case 2:
			configMenu(cfg)
		case 3:
			proxyMenu(cfg)
		case 4:
			showStatus()
		case 5:
			termux.SetupMenu()
		case 6:
			showDocs()
		case 7:
			utils.PrintInfo("Shutting down...")
			cleanup()
			return
		default:
			utils.PrintError("Invalid choice!")
			time.Sleep(1 * time.Second)
		}
	}
}

func startHybridEngine(cfg *config.Config) {
	utils.PrintHeader("🚀 STARTING HYBRID ENGINE")
	
	// Cek dependencies
	if !termux.CheckDependencies() {
		utils.PrintError("Missing dependencies! Run Termux Setup first")
		time.Sleep(2 * time.Second)
		return
	}

	// Initialize hybrid engine
	engine := hybrid.NewEngine(cfg)
	
	// Start in background
	go engine.Start()
	
	// Wait for engine to be ready
	time.Sleep(3 * time.Second)
	
	utils.PrintSuccess("✅ Hybrid Engine is RUNNING!")
	utils.PrintInfo("📡 Local Proxy: 127.0.0.1:1080 (SOCKS5)")
	utils.PrintInfo("🧅 Tor Network: Active (3 hops)")
	utils.PrintInfo("⛓️  Proxy Chain: 5 hops (rotating)")
	utils.PrintInfo("🔄 Rotator: Active per hop (10-25s interval)")
	utils.PrintInfo("\nPress Ctrl+C to stop...\n")

	// Keep running until Ctrl+C
	<-sigChan
	utils.PrintInfo("\n\nShutting down Hybrid Engine...")
	engine.Stop()
	cleanup()
}

func configMenu(cfg *config.Config) {
	for {
		utils.ClearScreen()
		utils.PrintHeader("⚙️ CONFIGURATION")
		fmt.Printf("  Chain Length      : %d\n", cfg.ChainLength)
		fmt.Printf("  Rotate Min        : %ds\n", cfg.RotateMin)
		fmt.Printf("  Rotate Max        : %ds\n", cfg.RotateMax)
		fmt.Printf("  Tor Enabled       : %v\n", cfg.TorEnabled)
		fmt.Printf("  Tor Port          : %d\n", cfg.TorPort)
		fmt.Printf("  Local Proxy Port  : %d\n", cfg.LocalPort)
		fmt.Printf("  Health Check      : %ds\n", cfg.HealthCheckInterval)
		fmt.Println()
		fmt.Println("  [1] Edit Chain Length (3-10)")
		fmt.Println("  [2] Edit Rotate Interval")
		fmt.Println("  [3] Toggle Tor")
		fmt.Println("  [4] Edit Ports")
		fmt.Println("  [5] Save & Back")
		fmt.Println()
		fmt.Print("👉 Choose: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			fmt.Print("Enter chain length (3-10): ")
			fmt.Scanln(&cfg.ChainLength)
			if cfg.ChainLength < 3 {
				cfg.ChainLength = 3
			}
			if cfg.ChainLength > 10 {
				cfg.ChainLength = 10
			}
		case 2:
			fmt.Print("Enter min interval (5-30): ")
			fmt.Scanln(&cfg.RotateMin)
			fmt.Print("Enter max interval (10-60): ")
			fmt.Scanln(&cfg.RotateMax)
		case 3:
			cfg.TorEnabled = !cfg.TorEnabled
		case 4:
			fmt.Print("Enter Tor port (9050-9150): ")
			fmt.Scanln(&cfg.TorPort)
			fmt.Print("Enter local proxy port (1024-65535): ")
			fmt.Scanln(&cfg.LocalPort)
		case 5:
			config.SaveConfig(cfg)
			return
		}
		config.SaveConfig(cfg)
	}
}

func proxyMenu(cfg *config.Config) {
	// Implement proxy pool management
	utils.PrintInfo("Proxy Manager - Coming soon in v1.1")
	time.Sleep(2 * time.Second)
}

func showStatus() {
	utils.PrintInfo("Status - Coming soon in v1.1")
	time.Sleep(2 * time.Second)
}

func showDocs() {
	utils.PrintInfo("Documentation - Check docs/ folder")
	time.Sleep(2 * time.Second)
}

func cleanup() {
	// Kill Tor process
	exec.Command("pkill", "tor").Run()
	utils.PrintSuccess("Cleanup complete")
}

var sigChan chan os.Signal
