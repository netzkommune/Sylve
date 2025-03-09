// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"strings"
	"sylve/internal/utils/sysctl"
)

func GetGeomXML() []byte {
	output, err := sysctl.GetBytes("kern.geom.confxml")
	if err != nil {
		return []byte{}
	}

	return output
}

func GetDiskTypeFromUUID(uuid string, defaultValue string) string {
	var diskTypes = map[string]string{
		/* This list is not exhaustive, but it covers the most common types, please feel free to add more */
		"00000000-0000-0000-0000-000000000000": "Unused Entry",
		"024DEE41-33E7-11D3-9D69-0008C781F39F": "MBR",
		"C12A7328-F81F-11D2-BA4B-00A0C93EC93B": "EFI",
		"21686148-6449-6E6F-744E-656564454649": "BIOS Boot",
		"D3BFE2DE-3DAF-11DF-BA40-E3A556D89593": "Intel Rapid Start (iFFS)",
		"F4019732-066E-4E12-8273-346C5641494F": "Sony Boot",
		"BFBFAFE7-A34F-448A-9A5B-6213EB736C22": "Lenovo Boot",
		"E3C9E316-0B5C-4DB8-817D-F92DF00215AE": "Windows MSR",
		"EBD0A0A2-B9E5-4433-87C0-68B6B72699C7": "Windows Basic Data",
		"5808C8AA-7E8F-42E0-85D2-E1E90434CFB3": "Windows LDM Metadata",
		"AF9B60A0-1431-4F62-BC68-3311714A69AD": "Windows LDM Data",
		"DE94BBA4-06D1-4D40-A16A-BFD50179D6AC": "Windows Recovery",
		"37AFFC90-EF7D-4E96-91C3-2D7AE055B174": "IBM GPFS",
		"E75CAF8F-F680-4CEE-AFA3-B001E56EFC2D": "Windows Storage Spaces",
		"558D43C5-A1AC-43C0-AAC8-D1472B2923D1": "Windows Storage Replica",
		"75894C1E-3AEB-11D3-B7C1-7B03A0000000": "HP-UX Data",
		"E2A1E728-32E3-11D6-A682-7B03A0000000": "HP-UX Service",
		"0FC63DAF-8483-4772-8E79-3D69D8477DE4": "Linux Filesystem",
		"A19D880F-05FC-4D3B-A006-743F0F84911E": "Linux RAID",
		"4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709": "Linux Root (x86-64)",
		"44479540-F297-41B2-9AF7-D131D5F0458A": "Linux Root (x86)",
		"0657FD6D-A4AB-43C4-84E5-0933C84B4F4F": "Linux Swap",
		"E6D6D379-F507-44C2-A23C-238F2A3DF928": "Linux LVM",
		"933AC7E1-2EB4-4F13-B844-0E14E2AEF915": "Linux /home",
		"BC13C2FF-59E6-4262-A352-B275FD6F7172": "Linux /boot",
		"CA7D7CCB-63ED-4C53-861C-1742536059CC": "Linux LUKS Encryption",
		"516E7CBA-6ECF-11D6-8FF8-00022D09712B": "ZFS",
		"516E7CB6-6ECF-11D6-8FF8-00022D09712B": "UFS",
		"83BD6B9D-7F41-11DC-BE0B-001560B84F0F": "Boot",
		"516E7CB5-6ECF-11D6-8FF8-00022D09712B": "Swap",
		"6A898CC3-1DD2-11B2-99A6-080020736631": "macOS APFS/ZFS",
		"7C3457EF-0000-11AA-AA11-00306543ECAC": "macOS APFS Container",
		"6A85CF4D-1DD2-11B2-99A6-080020736631": "Solaris Root",
		"6A87C46F-1DD2-11B2-99A6-080020736631": "Solaris Swap",
		"6A8B642B-1DD2-11B2-99A6-080020736631": "Solaris Backup",
		"49F48D32-B10E-11DC-B99B-0019D1879648": "NetBSD Swap",
		"49F48D5A-B10E-11DC-B99B-0019D1879648": "NetBSD FFS",
		"FE3A2A5D-4F32-41A7-B725-ACCC3285A309": "Chrome OS Kernel",
		"3CB8E202-3B7E-47DD-8A3C-7FF2A13CFCEC": "Chrome OS Root",
		"45B0969E-9B03-4F30-B4C6-B4B80CEFF106": "Ceph Journal",
		"4FBD7E29-9D25-41B8-AFD0-062C0CEFF05D": "Ceph OSD",
		"824CC7A0-36A8-11E3-890A-952519AD3F61": "OpenBSD Data",
		"C91818F9-8025-47AF-89D2-F030D7000C2C": "Plan 9",
		"AA31E02A-400F-11DB-9590-000C2911D1B8": "VMware VMFS",
		"9E1A2D38-C612-4316-AA26-8B49521E5A8B": "PowerPC PReP Boot",
		"2568845D-2332-4675-BC39-8FA5A4748D15": "Android Bootloader",
		"49A4D17F-93A3-45C1-A0DE-F50B2EBE2599": "Android Boot",
		"38F428E6-D326-425D-9140-6E0EA133647C": "Android System",
		"A893EF21-E428-470A-9E55-0668FD91A2D9": "Android Cache",
		"DC76DDA9-5AC1-491C-AF42-A82591580C0D": "Android Data",
		"8C8F8EFF-AC95-4770-814A-21994F2DBC8F": "VeraCrypt Encrypted",
	}

	uuid = strings.ToUpper(uuid)

	if diskType, ok := diskTypes[uuid]; ok {
		return diskType
	}

	return defaultValue
}
