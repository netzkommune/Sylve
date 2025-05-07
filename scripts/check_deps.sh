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

if command -v virsh >/dev/null 2>&1; then
    VIRSH_VERSION=$(virsh --version)
    echo "✅ virsh found: $VIRSH_VERSION"
else
    echo "❌ Error: virsh is required but not found. Install using 'pkg install libvirt'"
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

if grep -q '^libvirtd_enable="YES"' "$RC_CONF"; then
    echo "✅ libvirtd is enabled in rc.conf"
else
    echo "❌ Error: libvirtd is not enabled in rc.conf. Add 'libvirtd_enable=\"YES\"' to enable it."
    exit 1
fi

if grep -q '^vmm_load="YES"' "$RC_CONF"; then
    echo "✅ vmm is enabled in rc.conf"
else
    echo "❌ Error: vmm is not enabled in rc.conf. Add 'vmm_load=\"YES\"' to enable it."
    exit 1
fi

if grep -q '^if_bridge_load="YES"' "$RC_CONF"; then
    echo "✅ if_bridge is enabled in rc.conf"
else
    echo "❌ Error: if_bridge is not enabled in rc.conf. Add 'if_bridge_load=\"YES\"' to enable it."
    exit 1
fi

if grep -q '^nmdm_load="YES"' "$RC_CONF"; then
    echo "✅ nmdm is enabled in rc.conf"
else
    echo "❌ Error: nmdm is not enabled in rc.conf. Add 'nmdm_load=\"YES\"' to enable it."
    exit 1
fi

echo "=== Dependency check completed ==="
echo
exit 0