// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

//go:build freebsd

package sysctl

// #include <sys/types.h>
// #include <sys/sysctl.h>
// #include <stdlib.h>
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

func GetString(name string) (string, error) {
	bytes, err := GetBytes(name)
	if err != nil {
		return "", err
	}

	if len(bytes) > 0 && bytes[len(bytes)-1] == 0 {
		bytes = bytes[:len(bytes)-1]
	}

	return string(bytes), nil
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

func Set(name string, value []byte) error {
	newlen := C.size_t(len(value))
	nameC := C.CString(name)
	defer C.free(unsafe.Pointer(nameC))

	var newp unsafe.Pointer
	if len(value) > 0 {
		newp = unsafe.Pointer(&value[0])
	} else {
		newp = unsafe.Pointer(uintptr(0))
	}

	_, err := C.sysctlbyname(nameC, nil, nil, newp, newlen)
	return err
}

func SetInt32(name string, value int32) error {
	nameC := C.CString(name)
	defer C.free(unsafe.Pointer(nameC))
	newlen := C.size_t(unsafe.Sizeof(value))
	newp := unsafe.Pointer(&value)

	_, err := C.sysctlbyname(nameC, nil, nil, newp, newlen)
	return err
}
