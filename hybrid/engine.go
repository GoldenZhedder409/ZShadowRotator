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
	cfg          *config.Config
	torManager   *tor.TorManager
	proxyPool    *proxy.ProxyPool
	chainManager *proxy.ChainManager
	rotator      *proxy.Rotator
	tunnel       *network.Tunnel
	listener     *network.Listener
	mu           sync.RWMutex
	isRunning    bool
	stopChan     chan bool
	stats        map[string]interface{}
}

func NewEngine(cfg *config.Config) *HybridEngine {
	return &HybridEngine{
		cfg:      cfg,
		stopChan: make(chan bool),
		stats:    make(map[string]interface{}),
	}
}

func (he *HybridEngine) Start() {
	he.mu.Lock()
	defer he.mu.Unlock()
	
	if he.isRunning {
		return
	}
	
	utils.PrintHeader("🔥 HYBRID ENGINE INITIALIZING 🔥")
	
	// 1. Initialize proxy pool
	utils.PrintInfo("Loading proxy pool...")
	he.proxyPool = proxy.NewProxyPool(he.cfg)
	
	// 2. Start Tor if enabled
	if he.cfg.TorEnabled {
		utils.PrintInfo("Starting Tor network...")
		he.torManager = tor.NewTorManager(he.cfg.TorPort, he.cfg.TorControlPort)
		err := he.torManager.Start()
		if err != nil {
			utils.PrintError("Failed to start Tor: " + err.Error())
			return
		}
	}
	
	// 3. Create initial chain
	utils.PrintInfo("Building proxy chain...")
	he.chainManager = proxy.NewChainManager(he.proxyPool, he.cfg)
	initialChain := he.chainManager.GetActiveChain()
	
	// 4. Start rotator
	utils.PrintInfo("Starting hop rotators...")
	he.rotator = proxy.NewRotator(he.proxyPool, he.cfg, initialChain)
	
	// 5. Create tunnel
	utils.PrintInfo("Creating Tor→Proxy tunnel...")
	he.tunnel = network.NewTunnel(he.torManager, he.rotator)
	
	// 6. Start local listener
	utils.PrintInfo(fmt.Sprintf("Starting local proxy on port %d...", he.cfg.LocalPort))
	he.listener = network.NewListener(he.cfg.LocalPort, he.tunnel)
	err := he.listener.Start()
	if err != nil {
		utils.PrintError("Failed to start listener: " + err.Error())
		return
	}
	
	// 7. Start stats collector
	go he.collectStats()
	
	he.isRunning = true
	
	// Print success message
	he.printStatus()
}

func (he *HybridEngine) printStatus() {
	chain := he.rotator.GetCurrentChain()
	
	fmt.Println()
	fmt.Println(utils.ColorGreen + "╔════════════════════════════════════════════════════╗")
	fmt.Println("║         ✅ HYBRID ENGINE IS RUNNING ✅         ║")
	fmt.Println("╠════════════════════════════════════════════════════╣")
	fmt.Printf("║  🧅 Tor Network       : %-24s ║\n", "ACTIVE (3 hops)")
	fmt.Printf("║  ⛓️  Proxy Chain       : %d hops rotating%-11s ║\n", len(chain.Hops), "")
	fmt.Printf("║  🔄 Rotator           : %d-%ds per hop%-15s ║\n", he.cfg.RotateMin, he.cfg.RotateMax, "")
	fmt.Printf("║  🌍 Countries in chain: %-24s ║\n", he.formatCountries(chain))
	fmt.Printf("║  📡 Local Proxy       : 127.0.0.1:%d%-20s ║\n", he.cfg.LocalPort, "")
	fmt.Printf("║  🛡️  Status            : BERSARANG DI TENGAH TOR  ║")
	fmt.Println(utils.ColorReset + "\n╚════════════════════════════════════════════════════╝")
	fmt.Println()
}

func (he *HybridEngine) formatCountries(chain *proxy.Chain) string {
	var countries string
	for i, hop := range chain.Hops {
		if i > 0 {
			countries += " → "
		}
		countries += hop.Flag + " " + hop.Country
	}
	if len(countries) > 24 {
		return countries[:21] + "..."
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
	
	he.isRunning = false
	utils.PrintSuccess("Hybrid Engine stopped")
}

func (he *HybridEngine) IsRunning() bool {
	he.mu.RLock()
	defer he.mu.RUnlock()
	return he.isRunning
}
