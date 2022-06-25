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

package unpacktar_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"text/tabwriter"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	"github.com/aezhar/unpacktar"
)

func listDir(rootPath string) (string, error) {
	var buf bytes.Buffer

	tw := tabwriter.NewWriter(&buf, 8, 8, 1, ' ', tabwriter.AlignRight)
	err := filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == rootPath {
			return nil
		}

		sysStat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("%q: missing stat", path)
		}

		nameCol, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		if info.Mode().Type() == fs.ModeSymlink {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}

			nameCol = fmt.Sprintf("%s -> %s", nameCol, linkTarget)
		}

		mtime := time.Unix(sysStat.Mtim.Sec, sysStat.Mtim.Nsec)

		_, err = fmt.Fprintf(
			tw,
			"%s\t%d/%d\t%d\t%s\t %s\n",
			info.Mode().String(),
			sysStat.Uid,
			sysStat.Gid,
			sysStat.Size,
			mtime.Format(time.RFC3339),
			nameCol,
		)
		return err
	})
	if err != nil {
		return "", err
	}

	if err := tw.Flush(); err != nil {
		return "", nil
	}

	return buf.String(), nil
}

func TestUnpackTo(t *testing.T) {
	tempDir := t.TempDir()

	f := golden.Open(t, "testfile.tar.input")
	defer f.Close()

	err := unpacktar.To(f, tempDir)
	assert.NilError(t, err)

	listing, err := listDir(tempDir)
	assert.NilError(t, err)

	golden.Assert(t, listing, "testfile.tar.golden")

	filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return os.Chmod(path, 0700)
		}
		return os.Chmod(path, 0600)
	})
}
