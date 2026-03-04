// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: pr-reviewer <pr-url>\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "not yet implemented\n")
	os.Exit(1)
}
