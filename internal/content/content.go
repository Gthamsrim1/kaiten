// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package content

type Content interface {
	Read(offset int64, p []byte) (int, error)
	Write(offset int64, p []byte) (int, error)
	Size() uint64
	Resize(size uint64) error

	Bytes() ([]byte, error)
	Backing() *Backing
}
