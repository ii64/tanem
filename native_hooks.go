package emulator

import (
	zl  "github.com/rs/zerolog"
)

type NativeHooks struct {
	emu *Emulator
	nm  *NativeMemory
	ms  *Modules
	hk  *Hooker
	vfs *VirtualFileSystem
	logger zl.Logger
}
func NewNativeHooks(emu *Emulator,nm *NativeMemory,ms *Modules,hk *Hooker,vfs *VirtualFileSystem, logger zl.Logger) *NativeHooks {
	nh := &NativeHooks{
		emu: emu,nm: nm,ms: ms,hk: hk,vfs: vfs,
		logger: logger,
	}
	return nh
}