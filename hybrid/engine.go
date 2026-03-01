package hybrid

import (
	"ZShadowRotator/config"
	"ZShadowRotator/network"
	"ZShadowRotator/proxy"
	"ZShadowRotator/tor"
	"ZShadowRotator/utils"
	"fmt"
	"sync"
	"time"
)

type HybridEngine struct {
	cfg           *config.Config
	torManager    *tor.TorManager
	proxyPool     *proxy.ProxyPool
	chainManager  *proxy.ChainManager
	rotator       *proxy.Rotator
	tunnel        *network.Tunnel
	listener      *network.Listener
	adaptiveNet   *network.AdaptiveNetwork
	env           *utils.Environment
	mu            sync.RWMutex
	isRunning     bool
	stopChan      chan bool
	stats         map[string]interface{}
}

func NewEngine(cfg *config.Config) *HybridEngine {
	// Detect environment
	env := utils.DetectEnvironment()
	
	return &HybridEngine{
		cfg:         cfg,
		env:         env,
		stopChan:    make(chan bool),
		stats:       make(map[string]interface{}),
		adaptiveNet: network.NewAdaptiveNetwork(env),
	}
}

func (he *HybridEngine) Start() {
	he.mu.Lock()
	defer he.mu.Unlock()
	
	if he.isRunning {
		return
	}
	
	utils.PrintHeader("🔥 HYBRID ENGINE INITIALIZING 🔥")
	
	// Tampilkan environment
	he.env.PrintStatus()
	fmt.Println()
	
	// 1. Setup adaptive network
	if err := he.adaptiveNet.SetupRouting(he.cfg.LocalPort); err != nil {
		utils.PrintWarning("Network setup issue: " + err.Error())
	}
	
	// 2. Initialize proxy pool
	utils.PrintInfo("Loading proxy pool...")
	he.proxyPool = proxy.NewProxyPool(he.cfg)
	
	// 3. Start Tor if enabled
	if he.cfg.TorEnabled {
		utils.PrintInfo("Starting Tor network...")
		he.torManager = tor.NewTorManager(he.cfg.TorPort, he.cfg.TorControlPort)
		err := he.torManager.Start()
		if err != nil {
			utils.PrintError("Failed to start Tor: " + err.Error())
			utils.PrintInfo("Tor is optional. Continuing with proxy chain only...")
			he.torManager = nil
		}
	}
	
	// 4. Create initial chain
	utils.PrintInfo("Building proxy chain...")
	he.chainManager = proxy.NewChainManager(he.proxyPool, he.cfg)
	initialChain := he.chainManager.GetActiveChain()
	
	if initialChain == nil {
		utils.PrintError("No proxies available! Check your config.")
		return
	}
	
	// 5. Start rotator
	utils.PrintInfo("Starting hop rotators...")
	he.rotator = proxy.NewRotator(he.proxyPool, he.cfg, initialChain)
	
	// 6. Create tunnel (Tor may be nil)
	if he.torManager != nil {
		utils.PrintInfo("Creating Tor→Proxy tunnel...")
		he.tunnel = network.NewTunnel(he.torManager, he.rotator)
	} else {
		utils.PrintInfo("Creating Direct→Proxy tunnel (no Tor)...")
		he.tunnel = network.NewTunnel(nil, he.rotator)
	}
	
	// 7. Start local listener
	utils.PrintInfo(fmt.Sprintf("Starting local proxy on port %d...", he.cfg.LocalPort))
	he.listener = network.NewListener(he.cfg.LocalPort, he.tunnel)
	err := he.listener.Start()
	if err != nil {
		utils.PrintError("Failed to start listener: " + err.Error())
		return
	}
	
	// 8. Start stats collector
	go he.collectStats()
	
	he.isRunning = true
	
	// Print success message with environment-specific tips
	he.printAdaptiveStatus()
}

func (he *HybridEngine) printAdaptiveStatus() {
	chain := he.rotator.GetCurrentChain()
	
	fmt.Println()
	utils.PrintSuccess("╔════════════════════════════════════════════════════╗")
	utils.PrintSuccess("║         ✅ HYBRID ENGINE IS RUNNING ✅           ║")
	utils.PrintSuccess("╠════════════════════════════════════════════════════╣")
	
	if he.torManager != nil {
		utils.PrintSuccess(fmt.Sprintf("║  🧅 Tor Network       : ACTIVE (3 hops) %-14s ║", ""))
	} else {
		utils.PrintSuccess(fmt.Sprintf("║  🧅 Tor Network       : DISABLED %-21s ║", ""))
	}
	
	utils.PrintSuccess(fmt.Sprintf("║  ⛓️  Proxy Chain       : %d hops rotating%-13s ║", 
		len(chain.Hops), ""))
	utils.PrintSuccess(fmt.Sprintf("║  🔄 Rotator           : %d-%ds per hop%-15s ║", 
		he.cfg.RotateMin, he.cfg.RotateMax, ""))
	utils.PrintSuccess(fmt.Sprintf("║  🌍 Countries in chain: %-30s ║", 
		he.formatCountries(chain)))
	utils.PrintSuccess(fmt.Sprintf("║  📡 Local Proxy       : 127.0.0.1:%d%-20s ║", 
		he.cfg.LocalPort, ""))
	utils.PrintSuccess(fmt.Sprintf("║  🛡️  Mode              : %-30s ║", 
		he.env.RecommendMode))
	utils.PrintSuccess("╚════════════════════════════════════════════════════╝")
	fmt.Println()
	
	// Tips based on mode
	if !he.env.IptablesAvail {
		utils.PrintInfo("📱 " + he.env.RecommendMode + " MODE TIPS:")
		utils.PrintInfo("   Set browser proxy manually to 127.0.0.1:" + fmt.Sprint(he.cfg.LocalPort))
		utils.PrintInfo("   Type: SOCKS5")
		utils.PrintInfo("   DNS: Proxy DNS when using SOCKS5")
	}
	
	if he.env.IsTermux && !he.env.HasRoot && !he.env.HasTsu {
		utils.PrintInfo("\n📱 Termux Non-Root Shortcut:")
		utils.PrintInfo("   export ALL_PROXY=socks5://127.0.0.1:" + fmt.Sprint(he.cfg.LocalPort))
		utils.PrintInfo("   curl ifconfig.me  # Cek IP berubah")
	}
}

func (he *HybridEngine) formatCountries(chain *proxy.Chain) string {
	if chain == nil {
		return "No chain"
	}
	
	var countries string
	for i, hop := range chain.Hops {
		if i > 0 {
			countries += " → "
		}
		if hop != nil {
			countries += hop.Flag + " " + hop.Country
		} else {
			countries += "❌"
		}
	}
	if len(countries) > 30 {
		return countries[:27] + "..."
	}
	return countries
}

func (he *HybridEngine) collectStats() {
	ticker := time.NewTicker(5 * time.Second)
	
	for {
		select {
		case <-ticker.C:
			if he.rotator != nil {
				he.mu.Lock()
				he.stats["rotations"] = he.rotator.GetStats()
				he.stats["active_chain"] = he.rotator.GetCurrentChain()
				he.mu.Unlock()
			}
		case <-he.stopChan:
			ticker.Stop()
			return
		}
	}
}

func (he *HybridEngine) GetStats() map[string]interface{} {
	he.mu.RLock()
	defer he.mu.RUnlock()
	return he.stats
}

func (he *HybridEngine) Stop() {
	he.mu.Lock()
	defer he.mu.Unlock()
	
	if !he.isRunning {
		return
	}
	
	utils.PrintInfo("Stopping Hybrid Engine...")
	
	he.stopChan <- true
	
	if he.listener != nil {
		he.listener.Stop()
	}
	
	if he.rotator != nil {
		he.rotator.Stop()
	}
	
	if he.chainManager != nil {
		he.chainManager.Stop()
	}
	
	if he.torManager != nil {
		he.torManager.Stop()
	}
	
	if he.proxyPool != nil {
		he.proxyPool.Stop()
	}
	
	// Cleanup network rules
	he.adaptiveNet.Cleanup()
	
	he.isRunning = false
	utils.PrintSuccess("Hybrid Engine stopped")
}

func (he *HybridEngine) IsRunning() bool {
	he.mu.RLock()
	defer he.mu.RUnlock()
	return he.isRunning
}
