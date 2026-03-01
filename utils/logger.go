package utils

import (
	"fmt"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func PrintLogo() {
	fmt.Println(ColorCyan + `
╔════════════════════════════════════════════════════════════╗
║                                                           ║
║              🔥 ONION HYBRID FORTRESS v1.0 🔥             ║
║      "Based within ToR, stealth mode is in operation"     ║
║                                                           ║
╚════════════════════════════════════════════════════════════╝` + ColorReset + "\n")
}

func PrintHeader(title string) {
	fmt.Println(ColorPurple + "\n╔════════════════════════════════════════════════════╗")
	fmt.Printf("║  %-46s ║\n", title)
	fmt.Println("╚════════════════════════════════════════════════════╝" + ColorReset)
}

func PrintSuccess(msg string) {
	fmt.Println(ColorGreen + "✅ " + msg + ColorReset)
}

func PrintError(msg string) {
	fmt.Println(ColorRed + "❌ " + msg + ColorReset)
}

func PrintInfo(msg string) {
	fmt.Println(ColorBlue + "ℹ️  " + msg + ColorReset)
}

func PrintWarning(msg string) {
	fmt.Println(ColorYellow + "⚠️  " + msg + ColorReset)
}

func PrintDebug(msg string) {
	fmt.Println(ColorCyan + "🔧 " + msg + ColorReset)
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func Log(level, msg string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] [%s] %s\n", timestamp, level, msg)
}
