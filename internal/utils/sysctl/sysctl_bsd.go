// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

//go:build darwin || freebsd || openbsd || netbsd

package sysctl

// #include <sys/types.h>
// #include <sys/sysctl.h>
import "C"
import "unsafe"

func GetInt64(name string) (value int64, err error) {
	oldlen := C.size_t(8)
	_, err = C.sysctlbyname(C.CString(name), unsafe.Pointer(&value), &oldlen, nil, 0)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func GetString(name string) (value string, err error) {
	oldlen := C.size_t(0)
	_, err = C.sysctlbyname(C.CString(name), nil, &oldlen, nil, 0)
	if err != nil {
		return "", err
	}

	buf := make([]byte, oldlen)
	_, err = C.sysctlbyname(C.CString(name), unsafe.Pointer(&buf[0]), &oldlen, nil, 0)
	if err != nil {
		return "", err
	}

	return string(buf[:oldlen-1]), nil
}

func GetBytes(name string) (value []byte, err error) {
	oldlen := C.size_t(0)
	_, err = C.sysctlbyname(C.CString(name), nil, &oldlen, nil, 0)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, oldlen)
	_, err = C.sysctlbyname(C.CString(name), unsafe.Pointer(&buf[0]), &oldlen, nil, 0)
	if err != nil {
		return nil, err
	}

	return buf[:oldlen], nil
}
