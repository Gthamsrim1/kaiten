// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

type Config struct {
	Repo     string
	Snapshot string

	Command []string

	Hostname string

	WorkDir string

	Env []string

	ReadOnly bool
}

const ChildCommand = "__child__"
