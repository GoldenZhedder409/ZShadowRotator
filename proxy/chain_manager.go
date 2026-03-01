package proxy

import (
	"ZShadowRotator/config"
	"ZShadowRotator/utils"
	"fmt"
	"sync"
	"time"
)

type Chain struct {
	Hops      []*Proxy
	CreatedAt time.Time
	ID        string
}

type ChainManager struct {
	pool        *ProxyPool
	cfg         *config.Config
	activeChain *Chain
	backupChain *Chain
	mu          sync.RWMutex
	stopChan    chan bool
}

func NewChainManager(pool *ProxyPool, cfg *config.Config) *ChainManager {
	cm := &ChainManager{
		pool:     pool,
		cfg:      cfg,
		stopChan: make(chan bool),
	}
	
	// Create initial chains
	cm.activeChain = cm.createNewChain()
	cm.backupChain = cm.createNewChain()
	
	// Start chain refresher
	go cm.chainRefresher()
	
	return cm
}

func (cm *ChainManager) createNewChain() *Chain {
	alive := cm.pool.GetAliveProxies()
	if len(alive) < cm.cfg.ChainLength {
		utils.PrintWarning("Not enough alive proxies for chain length")
		// Return whatever we have
		if len(alive) == 0 {
			return nil
		}
	}
	
	// Randomly select proxies for chain
	used := make(map[string]bool)
	var hops []*Proxy
	
	for len(hops) < cm.cfg.ChainLength {
		if len(alive) == 0 {
			break
		}
		
		// Pilih random proxy
		idx := utils.RandomInt(0, len(alive)-1)
		proxy := &alive[idx]
		
		// Avoid duplicate countries in chain (optional)
		countryUsed := false
		for _, h := range hops {
			if h.Country == proxy.Country {
				countryUsed = true
				break
			}
		}
		
		if !countryUsed && !used[proxy.Address] {
			hops = append(hops, proxy)
			used[proxy.Address] = true
		}
		
		// Remove from alive list to avoid infinite loop
		alive = append(alive[:idx], alive[idx+1:]...)
	}
	
	// If we couldn't get enough, pad with random proxies
	for len(hops) < cm.cfg.ChainLength {
		proxy := cm.pool.GetRandomProxy()
		if proxy != nil {
			hops = append(hops, proxy)
		}
	}
	
	// Convert ID from int to string
	chainID := fmt.Sprintf("%d", utils.RandomInt(1000, 9999))
	
	return &Chain{
		Hops:      hops,
		CreatedAt: time.Now(),
		ID:        chainID, // Sekarang string, bukan int
	}
}

func (cm *ChainManager) chainRefresher() {
	ticker := time.NewTicker(10 * time.Minute) // Refresh chain every 10 minutes
	
	for {
		select {
		case <-ticker.C:
			cm.mu.Lock()
			// Rotate chains: backup becomes active, create new backup
			if cm.backupChain != nil {
				cm.activeChain = cm.backupChain
			}
			cm.backupChain = cm.createNewChain()
			cm.mu.Unlock()
			
			if cm.activeChain != nil {
				utils.PrintInfo(fmt.Sprintf("Chain rotated: New active chain ID %s", cm.activeChain.ID))
			}
			
		case <-cm.stopChan:
			ticker.Stop()
			return
		}
	}
}

func (cm *ChainManager) GetActiveChain() *Chain {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.activeChain
}

func (cm *ChainManager) GetBackupChain() *Chain {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.backupChain
}

func (cm *ChainManager) Stop() {
	cm.stopChan <- true
}
