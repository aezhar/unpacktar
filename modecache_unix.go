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
	"io/fs"
	"time"

	"golang.org/x/sys/unix"
)

func hasPAXHeader(n *modeCacheNode, name string) (has bool) {
	_, has = n.header.PAXRecords[name]
	return
}

func applyModesNode(n *modeCacheNode) error {
	if n.fullPath == "" {
		return nil
	}

	atime := time.Now()
	if hasPAXHeader(n, "atime") {
		atime = n.header.AccessTime
	}
	mtime := n.header.ModTime

	err := unix.Lutimes(n.fullPath, []unix.Timeval{
		{atime.Unix(), atime.UnixNano() / 1000 % 1000},
		{mtime.Unix(), mtime.UnixNano() / 1000 % 1000},
	})
	if err != nil {
		return err
	}

	if mode := n.header.FileInfo().Mode(); mode.Type() != fs.ModeSymlink {
		err = unix.Chmod(n.fullPath, uint32(mode.Perm()))
		if err != nil {
			return err
		}
	}

	return nil
}
