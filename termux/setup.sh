#!/data/data/com.termux/files/usr/bin/bash

echo "🔥 ZSHADOWROTATOR - TERMUX SETUP 🔥"
echo "===================================="

# Update packages
echo "📦 Updating packages..."
pkg update -y && pkg upgrade -y

# Install dependencies
echo "📦 Installing dependencies..."
pkg install -y golang tor iptables root-repo tsu
pkg install -y nano git wget curl

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go get -u github.com/gorilla/websocket
go get -u github.com/shadowsocks/go-shadowsocks2
go get -u github.com/xtaci/kcptun

# Setup Tor directory
echo "🔧 Configuring Tor..."
mkdir -p ~/.tor
chmod 700 ~/.tor

# Create config directory
echo "🔧 Creating config..."
mkdir -p ~/.config/zshadowrotator

# Setup iptables for non-root (using REDIRECT)
echo "🔧 Setting up network rules..."
if command -v iptables &> /dev/null; then
    # Try with tsu if available
    if command -v tsu &> /dev/null; then
        tsu iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 9040
        tsu iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports 9040
        echo "✅ iptables rules set (with root)"
    else
        echo "⚠️  iptables requires root. Run with 'tsu' for full functionality"
    fi
fi

# Build the project
echo "🔨 Building ZShadowRotator..."
cd ~/ZShadowRotator
go mod init ZShadowRotator
go mod tidy
go build -o zshadowrotator main.go

echo ""
echo "✅ SETUP COMPLETE!"
echo ""
echo "To run: ./zshadowrotator"
echo "Or: go run main.go"
echo ""
echo "⚠️  For best results, run with 'tsu' for iptables support"
