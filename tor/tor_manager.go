package tor

import (
	"ZShadowRotator/utils"
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type TorManager struct {
	cmd           *exec.Cmd
	controlConn   net.Conn
	socksPort     int
	controlPort   int
	mu            sync.Mutex
	isRunning     bool
	circuitID     string
	stopChan      chan bool
}

func NewTorManager(socksPort, controlPort int) *TorManager {
	return &TorManager{
		socksPort:   socksPort,
		controlPort: controlPort,
		stopChan:    make(chan bool),
	}
}

func (tm *TorManager) Start() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	// Check if Tor is already running
	if tm.isRunning {
		return nil
	}
	
	utils.PrintInfo("Starting Tor daemon...")
	
	// Kill any existing Tor process
	exec.Command("pkill", "tor").Run()
	time.Sleep(2 * time.Second)
	
	// Start Tor with config
	tm.cmd = exec.Command("tor",
		"--SocksPort", fmt.Sprintf("%d", tm.socksPort),
		"--ControlPort", fmt.Sprintf("%d", tm.controlPort),
		"--CookieAuthentication", "0",
		"--HashedControlPassword", "",
	)
	
	err := tm.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start Tor: %v", err)
	}
	
	// Wait for Tor to be ready
	time.Sleep(5 * time.Second)
	
	// Connect to control port
	err = tm.connectControl()
	if err != nil {
		return fmt.Errorf("failed to connect to Tor control: %v", err)
	}
	
	tm.isRunning = true
	
	// Start circuit monitor
	go tm.monitorCircuits()
	
	utils.PrintSuccess("Tor daemon started on port " + fmt.Sprintf("%d", tm.socksPort))
	return nil
}

func (tm *TorManager) connectControl() error {
	var err error
	tm.controlConn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", tm.controlPort))
	if err != nil {
		return err
	}
	
	// Authenticate (no password for simplicity)
	fmt.Fprintf(tm.controlConn, "AUTHENTICATE\r\n")
	
	reader := bufio.NewReader(tm.controlConn)
	resp, _ := reader.ReadString('\n')
	if !strings.HasPrefix(resp, "250") {
		return fmt.Errorf("authentication failed: %s", resp)
	}
	
	return nil
}

func (tm *TorManager) monitorCircuits() {
	ticker := time.NewTicker(10 * time.Minute)
	
	for {
		select {
		case <-ticker.C:
			// Request new circuit
			tm.requestNewCircuit()
		case <-tm.stopChan:
			ticker.Stop()
			return
		}
	}
}

func (tm *TorManager) requestNewCircuit() {
	if tm.controlConn == nil {
		return
	}
	
	// Signal Tor to use a new circuit
	fmt.Fprintf(tm.controlConn, "SIGNAL NEWNYM\r\n")
	
	reader := bufio.NewReader(tm.controlConn)
	resp, _ := reader.ReadString('\n')
	if strings.HasPrefix(resp, "250") {
		utils.PrintInfo("Tor circuit refreshed (NEWNYM signal sent)")
	}
}

func (tm *TorManager) GetSocksProxy() string {
	return fmt.Sprintf("127.0.0.1:%d", tm.socksPort)
}

func (tm *TorManager) IsRunning() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.isRunning
}

func (tm *TorManager) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	tm.stopChan <- true
	
	if tm.controlConn != nil {
		tm.controlConn.Close()
	}
	
	if tm.cmd != nil && tm.cmd.Process != nil {
		tm.cmd.Process.Kill()
	}
	
	tm.isRunning = false
	utils.PrintInfo("Tor daemon stopped")
}
