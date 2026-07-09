// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package node

import "testing"

func BenchmarkGetNode(b *testing.B) {
	f := &fakeFSNode{n: Node{ID: 1, Name: "bench"}}
	var fs FSNode = f

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.GetNode()
	}
}
