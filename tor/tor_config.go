package tor

import (
	"fmt"
	"io/ioutil"
	"os"
)

type TorConfig struct {
	SocksPort        int
	ControlPort      int
	DataDirectory    string
	LogLevel         string
	GeoIPFile        string
	ExitNodes        string
	EntryNodes       string
	ExcludeNodes     string
	StrictNodes      bool
}

func DefaultTorConfig() *TorConfig {
	return &TorConfig{
		SocksPort:     9050,
		ControlPort:   9051,
		DataDirectory: "/data/data/com.termux/files/home/.tor",
		LogLevel:      "notice",
		ExitNodes:     "", // Biarkan Tor pilih sendiri
		EntryNodes:    "", // Biarkan Tor pilih sendiri
		ExcludeNodes:  "{cn},{ru},{ir},{kp},{sy}",
		StrictNodes:   true,
	}
}

func (tc *TorConfig) GenerateTorrc() string {
	config := fmt.Sprintf(`SocksPort %d
ControlPort %d
DataDirectory %s
Log %s file %s/tor.log
CookieAuthentication 0

# Exclude certain countries
ExcludeNodes %s
StrictNodes %d

# Performance tuning
NumCPUs 2
NumDirectoryGuards 3
CircuitBuildTimeout 60
LearnCircuitBuildTimeout 0

# Avoid being an exit node (we're client only)
ExitRelay 0

# Use bridges if needed (uncomment if in censored country)
# UseBridges 1
# Bridge obfs4 ...

`, tc.SocksPort, tc.ControlPort, tc.DataDirectory, tc.LogLevel, tc.DataDirectory,
		tc.ExcludeNodes, boolToInt(tc.StrictNodes))
	
	return config
}

func (tc *TorConfig) SaveTorrc(path string) error {
	content := tc.GenerateTorrc()
	return ioutil.WriteFile(path, []byte(content), 0644)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func CreateTorDirectory() error {
	homeDir, _ := os.UserHomeDir()
	torDir := homeDir + "/.tor"
	return os.MkdirAll(torDir, 0700)
}
