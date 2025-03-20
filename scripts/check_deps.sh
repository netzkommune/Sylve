#!/bin/sh

echo "=== Checking system dependencies for Sylve ==="

OS=$(uname)
if [ "$OS" != "FreeBSD" ]; then
    echo "❌ Error: This script must be run on FreeBSD."
    exit 1
fi
echo "✅ OS Check: Running on FreeBSD."

RELEASE=$(freebsd-version | cut -d '.' -f 1)
if [ "$RELEASE" != "14" ]; then
    echo "⚠️ Error: This script is intended for FreeBSD 14.0. Detected version: $(freebsd-version)"
else
    echo "✅ FreeBSD version: $(freebsd-version)"
fi

# Check for Node.js and npm
if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then
    NODE_VERSION=$(node -v)
    NPM_VERSION=$(npm -v)
    echo "✅ Node.js found: $NODE_VERSION"
    echo "✅ npm found: $NPM_VERSION"
else
    echo "❌ Error: Node.js and npm are required but not found. Install using 'pkg install npm-20'"
    exit 1
fi

if command -v go >/dev/null 2>&1; then
    GO_VERSION=$(go version)
    echo "✅ Go found: $GO_VERSION"
else
    echo "❌ Error: Golang is required but not found. Install using 'pkg install go'"
    exit 1
fi

if command -v tmux >/dev/null 2>&1; then
    TMUX_VERSION=$(tmux -V)
    echo "✅ tmux found: $TMUX_VERSION"
else
    echo "❌ Error: tmux is required but not found. Install using 'pkg install tmux'"
    exit 1
fi

RC_CONF="/etc/rc.conf"

if grep -q '^smartd_enable="YES"' "$RC_CONF"; then
    echo "✅ smartd is enabled in rc.conf"
else
    echo "❌ Error: smartd is not enabled in rc.conf. Add 'smartd_enable=\"YES\"' to enable it."
    exit 1
fi

if grep -q '^linux_enable="YES"' "$RC_CONF"; then
    echo "✅ Linux compatibility mode is enabled in rc.conf"
else
    echo "❌ Error: Linux compatibility mode is not enabled in rc.conf. Add 'linux_enable=\"YES\"' to enable it."
    exit 1
fi

echo "=== Dependency check completed ==="
echo
exit 0