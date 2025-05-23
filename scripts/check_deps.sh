#!/bin/sh

echo "=== Checking system dependencies for Sylve ==="

if [ "$(uname)" != "FreeBSD" ]; then
    echo "❌ Error: This script must be run on FreeBSD."
    exit 1
fi
echo "✅ OS Check: Running on FreeBSD."

RELEASE=$(freebsd-version | cut -d '.' -f 1)
if [ "$RELEASE" -lt 14 ]; then
    echo "⚠️ Error: This script requires FreeBSD 14.0 or newer. Detected version: $(freebsd-version)"
    exit 1
else
    echo "✅ FreeBSD version: $(freebsd-version)"
fi

check_command() {
    if command -v "$1" >/dev/null 2>&1; then
        echo "✅ $1 found: $($2)"
    else
        echo "❌ Error: $1 is required but not found. Install using '$3'"
        exit 1
    fi
}

check_command node "node -v" "pkg install node20"
check_command npm "npm -v" "pkg install npm-node20"
check_command go "go version" "pkg install go"
check_command tmux "tmux -V" "pkg install tmux"
check_command virsh "virsh --version" "pkg install libvirt"

RC_CONF="/etc/rc.conf"
check_rcconf() {
    if grep -q "^$1=\"YES\"" "$RC_CONF"; then
        echo "✅ $1 is enabled in rc.conf"
    else
        echo "❌ Error: $1 is not enabled in rc.conf. Add '$1=\"YES\"' to enable it."
        exit 1
    fi
}

check_rcconf smartd_enable
check_rcconf linux_enable
check_rcconf libvirtd_enable
check_rcconf vmm_load
check_rcconf if_bridge_load
check_rcconf nmdm_load
check_rcconf gateway_enable
check_rcconf pf_enable

echo "=== Dependency check completed ==="
echo
exit 0
