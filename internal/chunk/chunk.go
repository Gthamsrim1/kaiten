// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chunk

type Chunk struct {
	Offset int64
	Length uint32
	Hash   [32]byte
	Data   []byte
}
