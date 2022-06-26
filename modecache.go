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
	"path"
	"path/filepath"
)

type modeCacheNode struct {
	header   *tar.Header
	fullPath string
	children []*modeCacheNode
}

type modeCache struct {
	rootPath string
	nodes    map[string]*modeCacheNode
}

func (c *modeCache) get(fpath string) *modeCacheNode {
	out, ok := c.nodes[fpath]
	if !ok {
		out = &modeCacheNode{}
		c.nodes[fpath] = out
	}
	return out
}

func (c *modeCache) add(fullPath string, h *tar.Header) error {
	relPath, err := filepath.Rel(c.rootPath, fullPath)
	if err != nil {
		return err
	}

	n := &modeCacheNode{header: h, fullPath: fullPath}

	parent := c.get(path.Dir(relPath))
	parent.children = append(parent.children, n)

	c.nodes[relPath] = n
	return nil
}

func newModeCache(rootPath string) *modeCache {
	return &modeCache{
		rootPath: rootPath,
		nodes:    make(map[string]*modeCacheNode),
	}
}

func applyModesNodes(mc *modeCache, n *modeCacheNode) error {
	for i := range n.children {
		if err := applyModesNodes(mc, n.children[i]); err != nil {
			return err
		}
	}

	if err := applyModesNode(n); err != nil {
		return fmt.Errorf("unpack/applyMode: %q: %w", n.fullPath, err)
	}

	return nil
}

func applyModes(mc *modeCache) error {
	root, ok := mc.nodes["."]
	if !ok {
		return errors.New("root not found")
	}

	return applyModesNodes(mc, root)
}
