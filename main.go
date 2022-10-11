package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/remeh/diago/pprof"
	"github.com/remeh/diago/proto"
)

func main() {
	runtime.LockOSThread()
	if config.File == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	var err error

	// read the pprof file
	// ----------------------

	var pprofProfile *pprof.Profile

	if pprofProfile, err = proto.ReadProtoFile(config.File); err != nil {
		fmt.Println("err:", err)
		os.Exit(-1)
	}

	// start the gui
	// ----------------------

	gui := NewGUI(pprofProfile)
	gui.OpenWindow()
}

func ReadProtoFile(s string) {
	panic("unimplemented")
}
