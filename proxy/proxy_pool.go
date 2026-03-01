package proxy

import (
	"ZShadowRotator/config"
	"ZShadowRotator/utils"
	"ZShadowRotator/validator"
	"fmt"
	"sync"
	"time"
)

type Proxy struct {
	Address   string
	Country   string
	Flag      string
	Code      string
	Protocol  string
	Myth      string
	Latency   time.Duration
	Alive     bool
	LastCheck time.Time
}

type ProxyPool struct {
	proxies      []Proxy
	mu           sync.RWMutex
	cfg          *config.Config
	healthTicker *time.Ticker
	stopChan     chan bool
}

var defaultProxies = []Proxy{
	// USA
	{"45.79.203.254:48388", "USA", "🇺🇸", "USA", "SOCKS5", "CIA Surveillance Van 🚐", 0, false, time.Time{}},
	{"104.219.236.127:1080", "USA", "🇺🇸", "USA", "SOCKS5", "Area 51 Secret Tunnel 🛸", 0, false, time.Time{}},
	{"165.22.110.253:1080", "USA", "🇺🇸", "USA", "SOCKS5", "Hollywood Green Screen 🎬", 0, false, time.Time{}},
	{"192.241.230.75:1080", "USA", "🇺🇸", "USA", "SOCKS5", "NYC Subway Node 🚇", 0, false, time.Time{}},
	
	// Germany
	{"185.133.239.244:16299", "Germany", "🇩🇪", "GER", "SOCKS5", "Bratwurst Security 🌭", 0, false, time.Time{}},
	{"185.194.217.97:1080", "Germany", "🇩🇪", "GER", "SOCKS5", "Octobersfest Hidden Lager 🍺", 0, false, time.Time{}},
	{"84.200.125.162:1080", "Germany", "🇩🇪", "GER", "SOCKS5", "AutoBahn High Speed 🏎️", 0, false, time.Time{}},
	{"46.4.53.115:1080", "Germany", "🇩🇪", "GER", "SOCKS5", "Black Forest Gateway 🌲", 0, false, time.Time{}},
	
	// Japan
	{"20.210.113.32:8123", "Japan", "🇯🇵", "JPN", "HTTP", "Akihabara Glitch 🤖", 0, false, time.Time{}},
	{"89.116.88.19:80", "Japan", "🇯🇵", "JPN", "HTTP", "Shibuya Crossing Ghost 👻", 0, false, time.Time{}},
	{"153.122.100.18:1080", "Japan", "🇯🇵", "JPN", "SOCKS5", "Mount Fuji Uplink 🏔️", 0, false, time.Time{}},
	{"45.125.44.118:1080", "Japan", "🇯🇵", "JPN", "SOCKS5", "Bullet Train Tunnel 🚄", 0, false, time.Time{}},
	
	// Brazil
	{"186.26.95.249:61445", "Brazil", "🇧🇷", "BRA", "SOCKS5", "Amazon Rain Forest Wi-Fi 🌳", 0, false, time.Time{}},
	{"187.17.201.203:38737", "Brazil", "🇧🇷", "BRA", "SOCKS5", "Maracana Stadium Node ⚽", 0, false, time.Time{}},
	{"177.136.124.47:56113", "Brazil", "🇧🇷", "BRA", "SOCKS5", "Rio Carnival Mask 🎭", 0, false, time.Time{}},
	{"191.252.62.147:1080", "Brazil", "🇧🇷", "BRA", "SOCKS5", "Samba Beat Rhythm 🥁", 0, false, time.Time{}},
	
	// India
	{"110.235.246.62:1080", "India", "🇮🇳", "IND", "SOCKS5", "Taj Mahal Mirror 🕌", 0, false, time.Time{}},
	{"64.227.131.240:1080", "India", "🇮🇳", "IND", "SOCKS5", "Bangalore Tech Spirit 🧘", 0, false, time.Time{}},
	{"139.59.24.173:1080", "India", "🇮🇳", "IND", "SOCKS5", "Curry Powered Server 🍛", 0, false, time.Time{}},
	{"103.149.162.194:1080", "India", "🇮🇳", "IND", "SOCKS5", "Bollywood Dance Number 💃", 0, false, time.Time{}},
	
	// Singapore
	{"165.22.80.17:1080", "Singapore", "🇸🇬", "SGP", "SOCKS5", "Merlion Water Cannon 🌊", 0, false, time.Time{}},
	{"167.172.112.65:1080", "Singapore", "🇸🇬", "SGP", "SOCKS5", "Marina Bay Sands Node 🏨", 0, false, time.Time{}},
	{"139.59.125.101:1080", "Singapore", "🇸🇬", "SGP", "SOCKS5", "Satay by the Bay BBQ 🍢", 0, false, time.Time{}},
	
	// Netherlands
	{"46.101.11.45:1080", "Netherlands", "🇳🇱", "NLD", "SOCKS5", "Tulip Field Server 🌷", 0, false, time.Time{}},
	{"188.166.98.210:1080", "Netherlands", "🇳🇱", "NLD", "SOCKS5", "Canal Boat Connection 🛶", 0, false, time.Time{}},
	{"95.179.175.62:1080", "Netherlands", "🇳🇱", "NLD", "SOCKS5", "Windmill Power ⚡", 0, false, time.Time{}},
	
	// Canada
	{"167.71.205.251:1080", "Canada", "🇨🇦", "CAN", "SOCKS5", "Maple Syrup Router 🍁", 0, false, time.Time{}},
	{"159.89.192.73:1080", "Canada", "🇨🇦", "CAN", "SOCKS5", "Poutine Protocol 🍟", 0, false, time.Time{}},
	{"138.197.199.102:1080", "Canada", "🇨🇦", "CAN", "SOCKS5", "Whistler Ski Lift ⛷️", 0, false, time.Time{}},
}

func NewProxyPool(cfg *config.Config) *ProxyPool {
	pool := &ProxyPool{
		proxies:   defaultProxies,
		cfg:       cfg,
		stopChan:  make(chan bool),
	}
	
	// Filter berdasarkan config
	pool.filterByConfig()
	
	// Start health checker
	go pool.healthChecker()
	
	return pool
}

func (p *ProxyPool) filterByConfig() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	var filtered []Proxy
	for _, proxy := range p.proxies {
		// Check if country is selected
		selected := false
		for _, c := range p.cfg.SelectedCountries {
			if proxy.Country == c {
				selected = true
				break
			}
		}
		
		// Check if country is excluded
		excluded := false
		for _, c := range p.cfg.ExcludedCountries {
			if proxy.Country == c {
				excluded = true
				break
			}
		}
		
		// Check protocol
		protocolOk := false
		for _, proto := range p.cfg.Protocols {
			if proxy.Protocol == proto {
				protocolOk = true
				break
			}
		}
		
		if selected && !excluded && protocolOk {
			filtered = append(filtered, proxy)
		}
	}
	
	p.proxies = filtered
}

func (p *ProxyPool) healthChecker() {
	p.healthTicker = time.NewTicker(time.Duration(p.cfg.HealthCheckInterval) * time.Second)
	
	for {
		select {
		case <-p.healthTicker.C:
			p.checkAllProxies()
		case <-p.stopChan:
			p.healthTicker.Stop()
			return
		}
	}
}

func (p *ProxyPool) checkAllProxies() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	var wg sync.WaitGroup
	for i := range p.proxies {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			alive, latency := validator.ValidateProxy(p.proxies[idx].Address, p.proxies[idx].Protocol)
			p.proxies[idx].Alive = alive
			p.proxies[idx].Latency = latency
			p.proxies[idx].LastCheck = time.Now()
		}(i)
	}
	wg.Wait()
	
	// Log hasil
	aliveCount := 0
	for _, pxy := range p.proxies {
		if pxy.Alive {
			aliveCount++
		}
	}
	utils.PrintDebug(fmt.Sprintf("Health check: %d/%d proxies alive", aliveCount, len(p.proxies)))
}

func (p *ProxyPool) GetAliveProxies() []Proxy {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var alive []Proxy
	for _, proxy := range p.proxies {
		if proxy.Alive {
			alive = append(alive, proxy)
		}
	}
	return alive
}

func (p *ProxyPool) GetRandomProxy() *Proxy {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var alive []Proxy
	for _, proxy := range p.proxies {
		if proxy.Alive {
			alive = append(alive, proxy)
		}
	}
	
	if len(alive) == 0 {
		return nil
	}
	
	return &alive[utils.RandomInt(0, len(alive)-1)]
}

func (p *ProxyPool) GetProxiesByCountry(country string) []Proxy {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var result []Proxy
	for _, proxy := range p.proxies {
		if proxy.Country == country && proxy.Alive {
			result = append(result, proxy)
		}
	}
	return result
}

func (p *ProxyPool) Stop() {
	p.stopChan <- true
}
