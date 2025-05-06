// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

//go:build freebsd

package iface

/*
#include <ifaddrs.h>
#include <net/if.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdlib.h>
#include <sys/ioctl.h>
#include <sys/sysctl.h>
#include <unistd.h>
#include <string.h>
#include <net/if_types.h>
#include <net/if_dl.h>
#include <net/if_var.h>
#include <netinet6/in6_var.h>
#include <net/if_media.h>
#include <netinet6/nd6.h>
#include <sys/sockio.h>
#include <errno.h>
#include <net/if_bridgevar.h>

static int
get_stp_op_params(int fd, const char *ifname, struct ifbropreq *opr)
{
    struct ifdrv dr;

    memset(&dr, 0, sizeof(dr));
    strncpy(dr.ifd_name, ifname, IFNAMSIZ-1);
    dr.ifd_cmd  = BRDGPARAM;
    dr.ifd_len  = sizeof(*opr);
    dr.ifd_data = opr;

    return ioctl(fd, SIOCGDRVSPEC, &dr);
}

static int
get_bridge_param(int fd, const char *ifname, int cmd, struct ifbrparam *bp)
{
    struct ifdrv dr;
    memset(&dr, 0, sizeof(dr));
    strncpy(dr.ifd_name, ifname, IFNAMSIZ-1);
    dr.ifd_cmd  = cmd;
    dr.ifd_len  = sizeof(*bp);
    dr.ifd_data = bp;
    return ioctl(fd, SIOCGDRVSPEC, &dr);
}

static void set_ifname(char *dst, const char *src) {
	strncpy(dst, src, IFNAMSIZ);
	dst[IFNAMSIZ - 1] = '\0';
}

static int get_nd6_flags(int fd, const char *name) {
	struct in6_ndireq nd;
	memset(&nd, 0, sizeof(nd));
	strncpy(nd.ifname, name, IFNAMSIZ - 1);
	if (ioctl(fd, SIOCGIFINFO_IN6, &nd) < 0)
		return -1;
	return nd.ndi.flags;
}

static int get_in6_flags(int fd, const char *name, struct in6_addr addr, uint32_t scope_id) {
	struct in6_ifreq ifr6;
	memset(&ifr6, 0, sizeof(ifr6));
	set_ifname(ifr6.ifr_name, name);
	ifr6.ifr_addr.sin6_family = AF_INET6;
	ifr6.ifr_addr.sin6_len = sizeof(struct sockaddr_in6);
	ifr6.ifr_addr.sin6_scope_id = scope_id;
	memcpy(&ifr6.ifr_addr.sin6_addr, &addr, sizeof(struct in6_addr));

	if (ioctl(fd, SIOCGIFAFLAG_IN6, &ifr6) < 0)
		return -1;

	return ifr6.ifr_ifru.ifru_flags6;
}

static void get_lifetimes(int fd, const char *name, struct in6_addr addr, uint32_t scope_id, uint32_t *pltime, uint32_t *vltime) {
	struct in6_ifreq ifr6;
	memset(&ifr6, 0, sizeof(ifr6));
	set_ifname(ifr6.ifr_name, name);
	ifr6.ifr_addr.sin6_family = AF_INET6;
	ifr6.ifr_addr.sin6_len = sizeof(struct sockaddr_in6);
	ifr6.ifr_addr.sin6_scope_id = scope_id;
	memcpy(&ifr6.ifr_addr.sin6_addr, &addr, sizeof(struct in6_addr));

	if (ioctl(fd, SIOCGIFALIFETIME_IN6, &ifr6) < 0) {
		*pltime = 0;
		*vltime = 0;
		return;
	}
	*pltime = ifr6.ifr_ifru.ifru_lifetime.ia6t_pltime;
	*vltime = ifr6.ifr_ifru.ifru_lifetime.ia6t_vltime;
}

static uint32_t get_flagshigh(const struct ifreq *req) {
	union {
		const struct sockaddr *sa;
		const uint32_t *u32;
	} pun;

	pun.sa = &req->ifr_addr;
	return *pun.u32;
}

static void get_capabilities(int fd, const char* name, uint32_t *enabled, uint32_t *supported) {
	struct ifreq req;
	memset(&req, 0, sizeof(req));
	strncpy(req.ifr_name, name, IFNAMSIZ - 1);

	if (ioctl(fd, SIOCGIFCAPNV, &req) < 0) {
		*enabled = 0;
		*supported = 0;
		return;
	}

	*enabled = req.ifr_curcap;
	*supported = req.ifr_reqcap;
}

static int ioctl_wrap(int fd, unsigned long req, void *arg) {
    return ioctl(fd, req, arg);
}

static uint64_t
get_combined_flags(int fd, const char *name)
{
    struct ifreq ifr;
    uint32_t low = 0, high = 0;

    memset(&ifr, 0, sizeof(ifr));
    strlcpy(ifr.ifr_name, name, sizeof(ifr.ifr_name));

    if (ioctl(fd, SIOCGIFFLAGS, &ifr) < 0) {
		return 0;
	}

    low = ifr.ifr_flags;

    return ((uint64_t)high << 32) | low;
}


static int get_media_info(int fd, const char *name, struct ifmediareq *ifmr) {
	memset(ifmr, 0, sizeof(struct ifmediareq));
	strncpy(ifmr->ifm_name, name, IFNAMSIZ - 1);
	return ioctl(fd, SIOCGIFMEDIA, ifmr);
}

static void* get_broadaddr(struct ifaddrs* a) {
	return a->ifa_broadaddr;
}

int is_bridge(const char *ifname) {
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock < 0) {
        return -1;
    }

    struct ifreq ifr;
    struct if_data ifd;

    memset(&ifr, 0, sizeof(ifr));
    strlcpy(ifr.ifr_name, ifname, sizeof(ifr.ifr_name));

    ifr.ifr_data = (caddr_t)&ifd;

    if (ioctl(sock, SIOCGIFDATA, &ifr) < 0) {
        close(sock);
        return -1;
    }

    close(sock);

    return (ifd.ifi_type == IFT_BRIDGE) ? 1 : 0;
}

static int
get_bridge_members(int fd, const char *ifname, struct ifbifconf *bifc)
{
    struct ifdrv dr;

    memset(bifc, 0, sizeof(*bifc));
    memset(&dr, 0, sizeof(dr));


    strlcpy(dr.ifd_name, ifname, IFNAMSIZ);
    dr.ifd_cmd  = BRDGGIFS;
    dr.ifd_len  = sizeof(*bifc);
    dr.ifd_data = bifc;
    if (ioctl(fd, SIOCGDRVSPEC, &dr) < 0)
        return -1;


    bifc->ifbic_buf = malloc(bifc->ifbic_len);
    if (bifc->ifbic_buf == NULL)
        return -1;

    dr.ifd_len  = sizeof(*bifc);
    dr.ifd_data = bifc;
    if (ioctl(fd, SIOCGDRVSPEC, &dr) < 0) {
        free(bifc->ifbic_buf);
        return -1;
    }

    return 0;
}



static uint32_t get_mtu(const struct ifreq *ifr)    { return ifr->ifr_mtu; }
static uint32_t get_metric(const struct ifreq *ifr) { return ifr->ifr_metric; }

static int get_errno(void) { return errno; }
*/
import "C"

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"sylve/pkg/utils/sysctl"
	"syscall"
	"unsafe"
)

type IPv4 struct {
	IP        net.IP `json:"ip"`
	Netmask   string `json:"netmask"`
	Broadcast net.IP `json:"broadcast"`
}

type IPv6 struct {
	IP           net.IP `json:"ip"`
	PrefixLength int    `json:"prefixLength"`
	ScopeID      uint32 `json:"scopeId"`
	AutoConf     bool   `json:"autoConf"`
	Detached     bool   `json:"detached"`
	Deprecated   bool   `json:"deprecated"`
	LifeTimes    struct {
		Preferred uint32 `json:"preferred"`
		Valid     uint32 `json:"valid"`
	} `json:"lifeTimes"`
}

type Flags struct {
	Raw  uint32   `json:"raw"`
	Desc []string `json:"desc"`
}

type Capabilities struct {
	Enabled   Flags `json:"enabled"`
	Supported Flags `json:"supported"`
}

type Media struct {
	Type       string   `json:"type"`
	Subtype    string   `json:"subtype"`
	Options    []string `json:"options"`
	Mode       string   `json:"mode"`
	RawCurrent int      `json:"rawCurrent"`
	RawActive  int      `json:"rawActive"`
	Status     string   `json:"status"`
}

type ND6 struct {
	Flags
}

type STP struct {
	Priority     int    `json:"priority"`
	HelloTime    int    `json:"hellotime"`
	FwdDelay     int    `json:"fwddelay"`
	MaxAge       int    `json:"maxage"`
	HoldCnt      int    `json:"holdcnt"`
	Proto        string `json:"proto"`
	RootID       string `json:"rootId"`
	RootPriority int    `json:"rootPriority"`
	RootPathCost int    `json:"ifcost"`
	RootPort     int    `json:"port"`
}

type BridgeMember struct {
	Name      string `json:"name"`
	Flags     Flags  `json:"flags"`
	IfMaxAddr int    `json:"ifmaxaddr"`
	State     int    `json:"state"`
	Priority  int    `json:"priority"`
	Port      int    `json:"port"`
	PathCost  int    `json:"pathCost"`
}

type Interface struct {
	Name          string         `json:"name"`
	Ether         string         `json:"ether"`
	Flags         Flags          `json:"flags"`
	MTU           int            `json:"mtu"`
	Metric        int            `json:"metric"`
	Capabilities  Capabilities   `json:"capabilities"`
	Driver        string         `json:"driver"`
	Description   string         `json:"description"`
	BridgeID      string         `json:"bridgeId"`
	STP           *STP           `json:"stp"`
	MaxAddr       int            `json:"maxaddr"`
	Timeout       int            `json:"timeout"`
	BridgeMembers []BridgeMember `json:"bridgeMembers"`

	IPv4 []IPv4 `json:"ipv4"`
	IPv6 []IPv6 `json:"ipv6"`

	Media *Media `json:"media"`

	ND6 ND6 `json:"nd6"`
}

type FlagDescriptor struct {
	Mask uint32
	Name string
}

func parseFlags(flags uint32, descriptors []FlagDescriptor) ([]string, uint32) {
	var descriptions []string
	remaining := flags

	for _, desc := range descriptors {
		if flags&desc.Mask != 0 {
			descriptions = append(descriptions, desc.Name)
			remaining &^= desc.Mask
		}
	}

	if remaining != 0 {
		descriptions = append(descriptions, fmt.Sprintf("UNKNOWN_0x%x", remaining))
	}

	return descriptions, remaining
}

func (iface *Interface) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s: flags=%x<%s> metric %d mtu %d\n",
		iface.Name,
		iface.Flags.Raw,
		strings.Join(iface.Flags.Desc, ","),
		iface.Metric,
		iface.MTU,
	))

	if iface.Ether != "" {
		sb.WriteString(fmt.Sprintf("\tether %s\n", iface.Ether))
	}

	if iface.Capabilities.Enabled.Raw != 0 {
		sb.WriteString(fmt.Sprintf("\toptions=%x<%s>\n", iface.Capabilities.Enabled.Raw, strings.Join(iface.Capabilities.Enabled.Desc, ",")))
	}

	if iface.Capabilities.Supported.Raw != 0 {
		sb.WriteString(fmt.Sprintf("\tcapabilities=%x<%s>\n", iface.Capabilities.Supported.Raw, strings.Join(iface.Capabilities.Supported.Desc, ",")))
	}

	for _, a := range iface.IPv4 {
		sb.WriteString(fmt.Sprintf("\tinet %s", a.IP))
		if a.Netmask != "" {
			sb.WriteString(fmt.Sprintf(" netmask %s", a.Netmask))
		}
		if a.Broadcast != nil {
			sb.WriteString(fmt.Sprintf(" broadcast %s", a.Broadcast.String()))
		}
		sb.WriteString("\n")
	}

	for _, a := range iface.IPv6 {
		sb.WriteString(fmt.Sprintf("\tinet6 %s prefixlen %d", a.IP, a.PrefixLength))
		if a.ScopeID != 0 {
			sb.WriteString(fmt.Sprintf(" scopeid 0x%x", a.ScopeID))
		}

		if a.AutoConf {
			sb.WriteString(" autoconf")
		}

		if a.Detached {
			sb.WriteString(" detached")
		}

		if a.Deprecated {
			sb.WriteString(" deprecated")
		}

		if a.LifeTimes.Preferred > 0 && a.LifeTimes.Preferred != 0xffffffff {
			sb.WriteString(fmt.Sprintf(" pltime %d", a.LifeTimes.Preferred))
		}

		if a.LifeTimes.Valid > 0 && a.LifeTimes.Valid != 0xffffffff {
			sb.WriteString(fmt.Sprintf(" vltime %d", a.LifeTimes.Valid))
		}

		sb.WriteString("\n")
	}

	if iface.Media != nil {
		sb.WriteString(fmt.Sprintf("\tmedia: %s %s", iface.Media.Type, iface.Media.Subtype))
		if iface.Media.Mode != "" {
			sb.WriteString(fmt.Sprintf(" %s", iface.Media.Mode))
		}
		if len(iface.Media.Options) > 0 {
			sb.WriteString(fmt.Sprintf(" <%s>", strings.Join(iface.Media.Options, ",")))
		}
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("\tstatus: %s\n", iface.Media.Status))
	}

	if iface.ND6.Raw != 0 {
		sb.WriteString(fmt.Sprintf("\tnd6 options=%x<%s>\n",
			iface.ND6.Raw,
			strings.Join(iface.ND6.Desc, ","),
		))
	}

	if iface.STP != nil {
		sb.WriteString(fmt.Sprintf("\tid %s priority %d hellotime %d fwddelay %d\n", iface.BridgeID, iface.STP.Priority, iface.STP.HelloTime, iface.STP.FwdDelay))
		sb.WriteString(fmt.Sprintf("\tmaxage %d holdcnt %d proto %s maxaddr %d timeout %d\n", iface.STP.MaxAge, iface.STP.HoldCnt, iface.STP.Proto, iface.MaxAddr, iface.Timeout))
		sb.WriteString(fmt.Sprintf("\troot id %s priority %d ifcost %d port %d\n", iface.STP.RootID, iface.STP.RootPriority, iface.STP.RootPathCost, iface.STP.RootPort))
	}

	if len(iface.BridgeMembers) > 0 {
		for _, member := range iface.BridgeMembers {
			sb.WriteString(fmt.Sprintf("\tmember: %s flags=%x<%s>\n", member.Name, member.Flags.Raw, strings.Join(member.Flags.Desc, ",")))
			sb.WriteString(fmt.Sprintf("\t\tifmaxaddr %d port %d priority %d path cost %d\n", member.IfMaxAddr, member.Port, member.Priority, member.PathCost))
		}
	} else {
		sb.WriteString("\tno members\n")
	}

	return sb.String()
}

func getSysctlProperty(iface, prop string) string {
	if strings.HasPrefix(iface, "lo") {
		if prop == "driver" {
			return "lo"
		}
		return "Loopback"
	}

	re := regexp.MustCompile(`^([a-zA-Z]+)(\d+)$`)
	matches := re.FindStringSubmatch(iface)
	if len(matches) != 3 {
		return ""
	}

	key := fmt.Sprintf("dev.%s.%s.%%%s", matches[1], matches[2], prop)
	value, err := sysctl.GetString(key)
	if err != nil {
		return ""
	}

	return strings.TrimRight(value, "\x00")
}

func parseFlagsDesc(fl uint32) []string {
	descriptors := []FlagDescriptor{
		{Mask: 0x1, Name: "UP"},
		{Mask: 0x2, Name: "BROADCAST"},
		{Mask: 0x8, Name: "LOOPBACK"},
		{Mask: 0x10, Name: "POINTOPOINT"},
		{Mask: 0x40, Name: "RUNNING"},
		{Mask: 0x80, Name: "NOARP"},
		{Mask: 0x100, Name: "PROMISC"},
		{Mask: 0x200, Name: "ALLMULTI"},
		{Mask: 0x400, Name: "OACTIVE"},
		{Mask: 0x800, Name: "SIMPLEX"},
		{Mask: 0x1000, Name: "LINK0"},
		{Mask: 0x2000, Name: "LINK1"},
		{Mask: 0x4000, Name: "LINK2"},
		{Mask: 0x8000, Name: "MULTICAST"},
		{Mask: 0x01000000, Name: "LOWER_UP"},
	}

	descriptions, _ := parseFlags(fl, descriptors)
	return descriptions
}

func parseCapabilitiesDesc(caps uint32) []string {
	descriptors := []FlagDescriptor{
		{Mask: 1 << 0, Name: "RXCSUM"},
		{Mask: 1 << 1, Name: "TXCSUM"},
		{Mask: 1 << 2, Name: "NETCONS"},
		{Mask: 1 << 3, Name: "VLAN_MTU"},
		{Mask: 1 << 4, Name: "VLAN_HWTAGGING"},
		{Mask: 1 << 5, Name: "JUMBO_MTU"},
		{Mask: 1 << 6, Name: "POLLING"},
		{Mask: 1 << 7, Name: "VLAN_HWCSUM"},
		{Mask: 1 << 8, Name: "TSO4"},
		{Mask: 1 << 9, Name: "TSO6"},
		{Mask: 1 << 10, Name: "LRO"},
		{Mask: 1 << 11, Name: "WOL_UCAST"},
		{Mask: 1 << 12, Name: "WOL_MCAST"},
		{Mask: 1 << 13, Name: "WOL_MAGIC"},
		{Mask: 1 << 14, Name: "TOE4"},
		{Mask: 1 << 15, Name: "TOE6"},
		{Mask: 1 << 16, Name: "VLAN_HWFILTER"},
		{Mask: 1 << 17, Name: "NV"},
		{Mask: 1 << 18, Name: "VLAN_HWTSO"},
		{Mask: 1 << 19, Name: "LINKSTATE"},
		{Mask: 1 << 20, Name: "NETMAP"},
		{Mask: 1 << 21, Name: "RXCSUM_IPV6"},
		{Mask: 1 << 22, Name: "TXCSUM_IPV6"},
		{Mask: 1 << 23, Name: "HWSTATS"},
		{Mask: 1 << 24, Name: "TXRTLMT"},
		{Mask: 1 << 25, Name: "HWRXTSTMP"},
		{Mask: 1 << 26, Name: "MEXTPG"},
		{Mask: 1 << 27, Name: "TXTLS4"},
		{Mask: 1 << 28, Name: "TXTLS6"},
		{Mask: 1 << 29, Name: "VXLAN_HWCSUM"},
		{Mask: 1 << 30, Name: "VXLAN_HWTSO"},
		{Mask: 1 << 31, Name: "TXTLS_RTLMT"},
	}

	descriptions, _ := parseFlags(caps, descriptors)
	return descriptions
}

func knownFlagMask() uint32 {
	return 0x1 | 0x2 | 0x8 | 0x10 | 0x40 | 0x80 | 0x100 | 0x200 | 0x400 |
		0x800 | 0x1000 | 0x2000 | 0x4000 | 0x8000 | 0x01000000
}

func parseSTPProto(p uint8) string {
	switch p {
	case 0:
		return "stp"
	case 1:
		return "-"
	case 2:
		return "rstp"
	default:
		return fmt.Sprintf("%d", p)
	}
}

func parseBridgeFlags(f uint32) []string {
	var bridgeFlagDesc = []FlagDescriptor{
		{Mask: C.IFBIF_LEARNING, Name: "LEARNING"},
		{Mask: C.IFBIF_DISCOVER, Name: "DISCOVER"},
		{Mask: C.IFBIF_STP, Name: "STP"},
		{Mask: C.IFBIF_SPAN, Name: "SPAN"},
		{Mask: C.IFBIF_STICKY, Name: "STICKY"},
		{Mask: C.IFBIF_BSTP_AUTOEDGE, Name: "AUTOEDGE"},
		{Mask: C.IFBIF_BSTP_AUTOPTP, Name: "AUTOPTP"},
	}

	desc, _ := parseFlags(f, bridgeFlagDesc)
	return desc
}

func getInterfaceInfo(name string) (*Interface, error) {
	fd := C.socket(C.AF_INET, C.SOCK_DGRAM, 0)
	if fd < 0 {
		return nil, fmt.Errorf("socket failed")
	}
	defer C.close(fd)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var flags Flags

	raw := uint32(C.get_combined_flags(fd, cname))
	flags.Raw = raw & knownFlagMask()

	if flags.Raw == 0 {
		return nil, fmt.Errorf("get_combined_flags failed for %s", name)
	}

	flags.Desc = parseFlagsDesc(flags.Raw)

	var req C.struct_ifreq
	C.strncpy(&req.ifr_name[0], cname, C.IFNAMSIZ-1)

	mtu := 0
	metric := 0

	if C.ioctl_wrap(fd, C.SIOCGIFMTU, unsafe.Pointer(&req)) >= 0 {
		mtu = int(C.get_mtu(&req))
	}

	if C.ioctl_wrap(fd, C.SIOCGIFMETRIC, unsafe.Pointer(&req)) >= 0 {
		metric = int(C.get_metric(&req))
	}

	var capEnabled, capSupported C.uint32_t
	C.get_capabilities(fd, cname, &capEnabled, &capSupported)

	var capabilities Capabilities
	capabilities.Enabled.Raw = uint32(capEnabled)
	capabilities.Enabled.Desc = parseCapabilitiesDesc(capabilities.Enabled.Raw)

	capabilities.Supported.Raw = uint32(capSupported)
	capabilities.Supported.Desc = parseCapabilitiesDesc(capabilities.Supported.Raw)

	iface := &Interface{
		Name:         name,
		Flags:        flags,
		MTU:          mtu,
		Metric:       metric,
		Capabilities: capabilities,
	}

	iface.Media = getMediaInfo(fd, name)

	if C.is_bridge(cname) != 0 {
		var opr C.struct_ifbropreq

		if C.get_stp_op_params(fd, cname, &opr) >= 0 {
			rawBridge := uint64(opr.ifbop_bridgeid)
			var baddr [6]byte
			for i := 0; i < 6; i++ {
				shift := uint((5 - i) * 8)
				baddr[i] = byte((rawBridge >> shift) & 0xff)
			}
			iface.BridgeID = net.HardwareAddr(baddr[:]).String()

			iface.STP = &STP{
				Priority:  int(opr.ifbop_priority),
				HelloTime: int(opr.ifbop_hellotime),
				FwdDelay:  int(opr.ifbop_fwddelay),
				MaxAge:    int(opr.ifbop_maxage),
				HoldCnt:   int(opr.ifbop_holdcount),
				Proto:     parseSTPProto(uint8(opr.ifbop_protocol)),
			}

			rawRoot := uint64(opr.ifbop_designated_root)
			var raddr [6]byte
			for i := 0; i < 6; i++ {
				shift := uint((5 - i) * 8)
				raddr[i] = byte((rawRoot >> shift) & 0xff)
			}

			iface.STP.RootID = net.HardwareAddr(raddr[:]).String()
			iface.STP.RootPriority = int(rawRoot >> 48)
			iface.STP.RootPathCost = int(opr.ifbop_root_path_cost)
			iface.STP.RootPort = int(opr.ifbop_root_port) & 0xfff
		}

		var bp C.struct_ifbrparam
		if C.get_bridge_param(fd, cname, C.BRDGGCACHE, &bp) >= 0 {
			raw := *(*C.uint32_t)(unsafe.Pointer(&bp.ifbrp_ifbrpu[0]))
			iface.MaxAddr = int(raw)
		}
		if C.get_bridge_param(fd, cname, C.BRDGGTO, &bp) >= 0 {
			raw := *(*C.uint32_t)(unsafe.Pointer(&bp.ifbrp_ifbrpu[0]))
			iface.Timeout = int(raw)
		}

		// --- fetch bridge members ---
		var bifc C.struct_ifbifconf
		ret := C.get_bridge_members(fd, cname, &bifc)

		if ret < 0 {
			// grab the C errno
			errno := syscall.Errno(C.get_errno())
			// print it for debugging
			fmt.Printf("get_bridge_members returned %d, errno=%d (%s)\n",
				ret, errno, errno)
			return nil, fmt.Errorf("get_bridge_members failed for %s: %v", name, errno)
		}

		// Extract the allocated buffer pointer from the union
		unionPtr := unsafe.Pointer(&bifc.ifbic_ifbicu[0])
		bufPtr := *(*unsafe.Pointer)(unionPtr)
		defer C.free(bufPtr)

		// Compute how many struct ifbreq entries we got
		elemSize := unsafe.Sizeof(C.struct_ifbreq{})
		count := int(bifc.ifbic_len) / int(elemSize)

		// Pull out the pointer-to-array of C.struct_ifbreq
		reqPtr := *(**C.struct_ifbreq)(unionPtr)
		// Build a Go slice header backed by that C array
		entries := (*[1 << 28]C.struct_ifbreq)(unsafe.Pointer(reqPtr))[:count:count]

		for _, entry := range entries {
			raw := uint32(entry.ifbr_ifsflags)
			member := BridgeMember{
				Name:      C.GoString(&entry.ifbr_ifsname[0]),
				Flags:     Flags{Raw: raw, Desc: parseBridgeFlags(raw)},
				IfMaxAddr: int(entry.ifbr_addrmax),
				State:     int(entry.ifbr_state),
				Priority:  int(entry.ifbr_priority),
				Port:      int(entry.ifbr_portno),
				PathCost:  int(entry.ifbr_path_cost),
			}
			iface.BridgeMembers = append(iface.BridgeMembers, member)
		}

	}

	return iface, nil
}

func parseMediaTypeBase(active int) string {
	switch active & 0xe0 {
	case 0x20:
		return "Ethernet"
	case 0x40:
		return "Token Ring"
	case 0x60:
		return "FDDI"
	case 0x80:
		return "Wi-Fi"
	case 0xa0:
		return "ATM"
	default:
		return fmt.Sprintf("Unknown (0x%x)", active&0xe0)
	}
}

func parseMediaSubtype(active int) string {
	subtypeMap := map[int]string{
		3: "10baseT/UTP", 6: "100baseTX", 16: "1000baseT", 26: "10Gbase-T",
		27: "40Gbase-CR4", 28: "40Gbase-SR4", 29: "40Gbase-LR4",
	}
	return subtypeMap[active&0x1f]
}

func parseMediaOptions(active int) []string {
	var opts []string
	if active&0x00100000 != 0 {
		opts = append(opts, "full-duplex")
	}
	if active&0x00200000 != 0 {
		opts = append(opts, "half-duplex")
	}
	return opts
}

func parseMediaMode(current int) string {
	switch current & 0x1f {
	case 0:
		return "autoselect"
	case 1:
		return "manual"
	case 2:
		return "none"
	default:
		return fmt.Sprintf("mode-0x%x", current&0x1f)
	}
}

func getMediaInfo(fd C.int, name string) *Media {
	var ifmr C.struct_ifmediareq
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if C.get_media_info(fd, cname, &ifmr) < 0 {
		return nil
	}

	active := int(ifmr.ifm_active)
	current := int(ifmr.ifm_current)

	media := &Media{
		Type:       parseMediaTypeBase(active),
		Subtype:    parseMediaSubtype(active),
		Options:    parseMediaOptions(active),
		Mode:       parseMediaMode(current),
		Status:     parseMediaStatus(ifmr.ifm_status, active&0x000000e0),
		RawCurrent: current,
		RawActive:  active,
	}

	return media
}

func parseMediaStatus(status C.int, mediaType int) string {
	const (
		IFM_AVALID = 0x00000001
		IFM_ACTIVE = 0x00000002
	)

	type statusDesc struct {
		validMask int
		activeBit int
		inactive  string
		active    string
	}

	descs := map[int]statusDesc{
		0x00000020: {IFM_AVALID, IFM_ACTIVE, "no carrier", "active"},
		0x00000040: {IFM_AVALID, IFM_ACTIVE, "no ring", "inserted"},
		0x00000060: {IFM_AVALID, IFM_ACTIVE, "no ring", "inserted"},
		0x00000080: {IFM_AVALID, IFM_ACTIVE, "no network", "active"},
		0x000000a0: {IFM_AVALID, IFM_ACTIVE, "no network", "active"},
	}

	st := int(status)

	if desc, ok := descs[mediaType]; ok {
		if st&desc.validMask != 0 {
			if st&desc.activeBit != 0 {
				return desc.active
			}
			return desc.inactive
		}
	}

	return "unknown"
}

func parseND6Options(flags uint32) []string {
	const (
		ND6_IFF_PERFORMNUD       = 0x01
		ND6_IFF_ACCEPT_RTADV     = 0x02
		ND6_IFF_PREFER_SOURCE    = 0x04
		ND6_IFF_IFDISABLED       = 0x08
		ND6_IFF_DONT_SET_IFROUTE = 0x10
		ND6_IFF_AUTO_LINKLOCAL   = 0x20
		ND6_IFF_NO_RADR          = 0x40
		ND6_IFF_NO_PREFER_IFACE  = 0x80
	)

	opts := []struct {
		mask uint32
		name string
	}{
		{ND6_IFF_PERFORMNUD, "PERFORMNUD"},
		{ND6_IFF_ACCEPT_RTADV, "ACCEPT_RTADV"},
		{ND6_IFF_PREFER_SOURCE, "PREFER_SOURCE"},
		{ND6_IFF_IFDISABLED, "IFDISABLED"},
		{ND6_IFF_DONT_SET_IFROUTE, "DONT_SET_IFROUTE"},
		{ND6_IFF_AUTO_LINKLOCAL, "AUTO_LINKLOCAL"},
		{ND6_IFF_NO_RADR, "NO_RADR"},
		{ND6_IFF_NO_PREFER_IFACE, "NO_PREFER_IFACE"},
	}

	var out []string
	for _, opt := range opts {
		if flags&opt.mask != 0 {
			out = append(out, opt.name)
		}
	}
	return out
}

func List() ([]*Interface, error) {
	var addrs *C.struct_ifaddrs
	if C.getifaddrs(&addrs) != 0 {
		return nil, fmt.Errorf("getifaddrs failed")
	}
	defer C.freeifaddrs(addrs)

	seen := make(map[string]bool)
	var result []*Interface

	for a := addrs; a != nil; a = a.ifa_next {
		if a.ifa_addr == nil {
			continue
		}

		name := C.GoString(a.ifa_name)
		if seen[name] {
			continue
		}
		seen[name] = true

		iface, err := Get(name)
		if err != nil {
			continue
		}
		result = append(result, iface)
	}

	return result, nil
}

func Get(name string) (*Interface, error) {
	fd6 := C.socket(C.AF_INET6, C.SOCK_DGRAM, 0)
	if fd6 < 0 {
		return nil, fmt.Errorf("socket(AF_INET6) failed")
	}
	defer C.close(fd6)

	var addrs *C.struct_ifaddrs
	if C.getifaddrs(&addrs) != 0 {
		return nil, fmt.Errorf("getifaddrs failed")
	}
	defer C.freeifaddrs(addrs)

	var iface *Interface
	for a := addrs; a != nil; a = a.ifa_next {
		if a.ifa_addr == nil {
			continue
		}

		ifaceName := C.GoString(a.ifa_name)
		if ifaceName != name {
			continue
		}

		if iface == nil {
			var err error
			iface, err = getInterfaceInfo(name)
			if err != nil {
				return nil, err
			}
			iface.Driver = getSysctlProperty(name, "driver")
			iface.Description = getSysctlProperty(name, "desc")
		}

		cname := C.CString(name)
		defer C.free(unsafe.Pointer(cname))

		switch a.ifa_addr.sa_family {
		case C.AF_INET:
			sa := (*C.struct_sockaddr_in)(unsafe.Pointer(a.ifa_addr))
			ip := net.IP(C.GoBytes(unsafe.Pointer(&sa.sin_addr), 4))

			var netmask net.IPMask
			if a.ifa_netmask != nil {
				nm := (*C.struct_sockaddr_in)(unsafe.Pointer(a.ifa_netmask))
				netmask = net.IPMask(C.GoBytes(unsafe.Pointer(&nm.sin_addr), 4))
			}

			var broadcast net.IP
			baPtr := C.get_broadaddr(a)
			if iface.Flags.Raw&C.IFF_BROADCAST != 0 && baPtr != nil {
				ba := (*C.struct_sockaddr_in)(unsafe.Pointer(baPtr))
				broadcast = make(net.IP, 4)
				C.memcpy(unsafe.Pointer(&broadcast[0]), unsafe.Pointer(&ba.sin_addr), 4)
			}

			iface.IPv4 = append(iface.IPv4, IPv4{
				IP:        ip,
				Netmask:   net.IP(netmask).String(),
				Broadcast: broadcast,
			})

		case C.AF_INET6:
			if iface.ND6.Raw == 0 {
				var nd6req C.struct_in6_ndireq
				C.set_ifname(&nd6req.ifname[0], cname)
				if C.ioctl_wrap(fd6, C.SIOCGIFINFO_IN6, unsafe.Pointer(&nd6req)) >= 0 {
					iface.ND6.Raw = uint32(nd6req.ndi.flags)
					iface.ND6.Desc = parseND6Options(iface.ND6.Raw)
				}
			}

			sa := (*C.struct_sockaddr_in6)(unsafe.Pointer(a.ifa_addr))
			ip := net.IP(C.GoBytes(unsafe.Pointer(&sa.sin6_addr), 16))

			prefixLen := 0
			if a.ifa_netmask != nil {
				nm := (*C.struct_sockaddr_in6)(unsafe.Pointer(a.ifa_netmask))
				mask := net.IPMask(C.GoBytes(unsafe.Pointer(&nm.sin6_addr), 16))
				prefixLen, _ = mask.Size()
			}

			scopeID := uint32(sa.sin6_scope_id)
			addr := *(*[16]byte)(unsafe.Pointer(&sa.sin6_addr))
			var caddr C.struct_in6_addr
			C.memcpy(unsafe.Pointer(&caddr), unsafe.Pointer(&addr), C.sizeof_struct_in6_addr)

			raw := C.get_in6_flags(fd6, cname, caddr, C.uint32_t(scopeID))

			plTime := C.uint32_t(0)
			vlTime := C.uint32_t(0)
			C.get_lifetimes(fd6, cname, caddr, C.uint32_t(scopeID), &plTime, &vlTime)

			ipv6 := IPv6{
				IP:           ip,
				PrefixLength: prefixLen,
				ScopeID:      scopeID,
			}

			if raw >= 0 {
				flags := uint32(raw)
				ipv6.AutoConf = flags&C.IN6_IFF_AUTOCONF != 0
				ipv6.Detached = flags&C.IN6_IFF_DETACHED != 0
				ipv6.Deprecated = flags&C.IN6_IFF_DEPRECATED != 0
			}

			ipv6.LifeTimes.Preferred = uint32(plTime)
			ipv6.LifeTimes.Valid = uint32(vlTime)

			iface.IPv6 = append(iface.IPv6, ipv6)

		case C.AF_LINK:
			sdl := (*C.struct_sockaddr_dl)(unsafe.Pointer(a.ifa_addr))
			if sdl.sdl_type == C.IFT_ETHER && sdl.sdl_alen == 6 {
				mac := C.GoBytes(unsafe.Pointer(&sdl.sdl_data[sdl.sdl_nlen]), 6)
				iface.Ether = net.HardwareAddr(mac).String()
			}
		}
	}

	if iface == nil {
		return nil, fmt.Errorf("interface %s not found", name)
	}

	return iface, nil
}
