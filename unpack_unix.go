// Copyright 2022 individual contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// <https://www.apache.org/licenses/LICENSE-2.0>
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

//go:build freebsd || linux

package unpacktar

import (
	"archive/tar"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func unpackBlockCharFifo(h *tar.Header, path string) error {
	mode := uint32(h.Mode & 07777)
	switch h.Typeflag {
	case tar.TypeBlock:
		mode |= unix.S_IFBLK
	case tar.TypeChar:
		mode |= unix.S_IFCHR
	case tar.TypeFifo:
		mode |= unix.S_IFIFO
	}

	return unix.Mknod(path, mode, int(unix.Mkdev(uint32(h.Devmajor), uint32(h.Devminor))))
}

func unpackPlatformSpecific(target string, h *tar.Header, fullPath string) error {
	switch h.Typeflag {
	case tar.TypeLink:
		linkName := h.Linkname
		if !filepath.IsAbs(linkName) {
			linkName = filepath.Join(target, linkName)
		}
		return os.Link(linkName, fullPath)
	case tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
		return unpackBlockCharFifo(h, fullPath)
	default:
		return fmt.Errorf("%x: %w", h.Typeflag, ErrUnsupportedType)
	}
}
