// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validate

import (
	"io/ioutil"
	"log"
	"os"
)

var PrintDebugInfo = false

var logger *log.Logger

func init() {
	w := ioutil.Discard

	if PrintDebugInfo {
		w = os.Stderr
	}

	logger = log.New(w, "", log.Lshortfile)
	log.SetFlags(log.Lshortfile)
}
