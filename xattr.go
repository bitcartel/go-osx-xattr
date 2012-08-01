// Copyright 2012 Bitcartel Software. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

// Package xattr wraps OS X functions to manipulate the extended attributes of a file, directory and symbolic link.
//
// The functions are a wrapper around xattr system calls.
// One difference is that position and size arguments are not implemented, since the caller can create slices from returned data.
package xattr

//#include <stdlib.h>
//#include <sys/xattr.h>
import "C"
import "unsafe"
import "bytes"

// Option flags to Xattr wrapping those in <sys/xattr.h>    
const (
	XATTR_NOFOLLOW          int = C.XATTR_NOFOLLOW        /* Don't follow symbolic links */
	XATTR_CREATE            int = C.XATTR_CREATE          /* set the value, fail if attr already exists */
	XATTR_REPLACE           int = C.XATTR_REPLACE         /* set the value, fail if attr does not exist */
	XATTR_SHOWCOMPRESSION   int = C.XATTR_SHOWCOMPRESSION /* option for f/getxattr() and f/listxattr() to expose the HFS Compression extended attributes */
	XATTR_NOSECURITY        int = C.XATTR_NOSECURITY      /* Set this to bypass authorization checking (eg. if doing auth-related work) */
	XATTR_NODEFAULT         int = C.XATTR_NODEFAULT       /* Set this to bypass the default extended attribute file (dot-underscore file) */
	XATTR_MAXNAMELEN            = 127
	XATTR_FINDERINFO_NAME       = "com.apple.FinderInfo"
	XATTR_RESOURCEFORK_NAME     = "com.apple.ResourceFork"
	XATTR_MAXSIZE               = (64 * 1024 * 1024)
)

// Removexattr will remove the named attribute from file at path
func Removexattr(path string, name string, options int) error {
	p := C.CString(path)
	n := C.CString(name)
	defer C.free(unsafe.Pointer(p))
	defer C.free(unsafe.Pointer(n))
	_, err := C.removexattr(p, n, C.int(options))
	if err != nil {
		return err
	}
	return nil
}

// Setxattr will set the value for a named attribute on file at path
func Setxattr(path string, name string, data []byte, options int) error {
	p := C.CString(path)
	n := C.CString(name)
	defer C.free(unsafe.Pointer(p))
	defer C.free(unsafe.Pointer(n))

	_, err := C.setxattr(p, n, unsafe.Pointer(&data[0]), C.size_t(len(data)), 0, C.int(options))
	if err != nil {
		return err
	}
	return nil
}

// Getxattr will return the value for the named attribute from a file at path
func Getxattr(path string, name string, options int) ([]byte, error) {
	p := C.CString(path)
	n := C.CString(name)
	defer C.free(unsafe.Pointer(p))
	defer C.free(unsafe.Pointer(n))

	// get size of data for attribute, type _Ctype_ssize_t
	attrsize, err := C.getxattr(p, n, nil, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	size := int(attrsize)

	// get data for attribute
	buf := make([]byte, size)
	x, err := C.getxattr(p, n, unsafe.Pointer(&buf[0]), C.size_t(size), 0, C.int(options))
	if err != nil {
		return nil, err
	}

	return buf[:x], nil
}

// Listxattr will return a list of attribute names found for a file at path
func Listxattr(path string, options int) ([]string, error) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))

	// get size of buffer needed for attribute names, type is _Ctype_ssize_t
	listsize, err := C.listxattr(p, nil, 0, 0)
	if err != nil {
		return nil, err
	}

	// get attribute names
	size := int(listsize)
	buf := make([]byte, size)
	_, err = C.listxattr(p, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(size), C.int(options))
	if err != nil {
		return nil, err
	}

	// split attribute names into string array
	var result []string
	b := bytes.Split(buf, []byte{0})
	for _, v := range b {
		// Split returns an empty slice after last separator, so check length.
		if len(v) > 0 {
			result = append(result, string(v))
		}
	}
	return result, nil
}
