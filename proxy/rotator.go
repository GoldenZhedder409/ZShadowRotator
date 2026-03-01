package proxy

import (
	"ZShadowRotator/config"
	"ZShadowRotator/utils"
	"fmt"
	"sync"
	"time"
)

type Rotator struct {
	pool          *ProxyPool
	cfg           *config.Config
	chain         *Chain
	hopTimers     []*time.Timer
	hopIntervals  []time.Duration
	mu            sync.RWMutex
	stopChan      chan bool
	rotationCount int
}

func NewRotator(pool *ProxyPool, cfg *config.Config, chain *Chain) *Rotator {
	r := &Rotator{
		pool:         pool,
		cfg:          cfg,
		chain:        chain,
		hopTimers:    make([]*time.Timer, len(chain.Hops)),
		hopIntervals: make([]time.Duration, len(chain.Hops)),
		stopChan:     make(chan bool),
	}
	
	// Set random intervals for each hop
	for i := 0; i < len(chain.Hops); i++ {
		r.hopIntervals[i] = utils.RandomDuration(cfg.RotateMin, cfg.RotateMax)
	}
	
	// Start all hop rotators
	r.startAllRotators()
	
	return r
}

func (r *Rotator) startAllRotators() {
	for i := 0; i < len(r.chain.Hops); i++ {
		go r.rotateHop(i)
	}
}

func (r *Rotator) rotateHop(hopIndex int) {
	interval := r.hopIntervals[hopIndex]
	
	for {
		select {
		case <-time.After(interval):
			r.mu.Lock()
			
			// Get new proxy for this hop
			newProxy := r.pool.GetRandomProxy()
			if newProxy != nil {
				oldProxy := r.chain.Hops[hopIndex]
				r.chain.Hops[hopIndex] = newProxy
				r.rotationCount++
				
				// Log rotasi
				utils.PrintDebug(fmt.Sprintf("Hop %d rotated: %s %s → %s %s (interval: %v)",
					hopIndex+1,
					oldProxy.Flag, oldProxy.Country,
					newProxy.Flag, newProxy.Country,
					interval))
			}
			
			// Randomize next interval (add jitter)
			newInterval := utils.RandomDuration(r.cfg.RotateMin, r.cfg.RotateMax)
			interval = utils.RandomJitter(newInterval, 30) // 30% jitter
			r.hopIntervals[hopIndex] = interval
			
			r.mu.Unlock()
			
		case <-r.stopChan:
			return
		}
	}
}

func (r *Rotator) GetCurrentChain() *Chain {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.chain
}

func (r *Rotator) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_rotations"] = r.rotationCount
	stats["chain_length"] = len(r.chain.Hops)
	
	hopInfo := make([]map[string]interface{}, len(r.chain.Hops))
	for i, hop := range r.chain.Hops {
		hopInfo[i] = map[string]interface{}{
			"position": i + 1,
			"country":  hop.Country,
			"flag":     hop.Flag,
			"address":  hop.Address,
			"protocol": hop.Protocol,
			"interval": r.hopIntervals[i].String(),
		}
	}
	stats["hops"] = hopInfo
	
	return stats
}

func (r *Rotator) Stop() {
	r.stopChan <- true
}
