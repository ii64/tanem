package main

import (
	"fmt"
	"time"
	zl "github.com/rs/zerolog"
	emulator "github.com/ii64/tanem"
)

var (
	start = time.Now()
)

func main() {
	defer func(){
		fmt.Printf("Exec time: %s\n", time.Now().Sub(start))
	}()
	//
	fmt.Println("Emulate example v0")	
	zl.SetGlobalLevel(zl.DebugLevel)
	//
	cnf := emulator.NewDefaultOptions()
	cnf.LogAs = emulator.ConsoleLog
	cnf.LogColor = true
	//
	emu, err := emulator.NewEmulator(cnf)
	if err != nil {
		panic(err)
	}
	libx, err := emu.LoadLibrary("./bin/libcms_new.so", true)
	if err != nil {
		panic(err)
	}
	for _, md := range emu.Modules.GetModules() {
		fmt.Println(md.Name())
	}
	addr, exist := libx.FindSymbol("JNI_OnLoad")
	fmt.Printf("JNI_OnLoad addr:0x%08X exist:%v\n", addr, exist)
}