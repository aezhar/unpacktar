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

package unpacktar

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/multierr"
)

var ErrUnsupportedType = errors.New("unsupported type")

func unpackFile(r io.Reader, h *tar.Header, path string) (err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Close(f))

	if _, err := io.CopyN(f, r, h.Size); err != nil {
		return err
	}

	return f.Sync()
}

func To(r io.Reader, target string) error {
	mc := newModeCache(target)
	tr := tar.NewReader(r)
	for h, err := tr.Next(); !errors.Is(err, io.EOF); h, err = tr.Next() {
		switch {
		case err != nil:
			return err
		case h == nil:
			continue
		case h.Name == "pax_global_header":
			continue
		}

		fullPath := filepath.Join(target, h.Name)
		switch h.Typeflag {
		case tar.TypeCont, tar.TypeXGlobalHeader, tar.TypeXHeader:
			continue
		case tar.TypeReg, tar.TypeRegA:
			err = unpackFile(tr, h, fullPath)
		case tar.TypeDir:
			err = os.MkdirAll(fullPath, 0700)
		case tar.TypeSymlink:
			err = os.Symlink(h.Linkname, fullPath)
		default:
			err = unpackPlatformSpecific(h)
		}
		if err != nil {
			return fmt.Errorf("unpack: %q: %w", fullPath, err)
		}

		if err := mc.add(fullPath, h); err != nil {
			return err
		}
	}

	return applyModes(mc)
}
