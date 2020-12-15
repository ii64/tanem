package emulator

import (
//	log "github.com/rs/zerolog/log"
	zl  "github.com/rs/zerolog"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)


type NativeMemory struct {
	mu      uc.Unicorn
	mem     *MemoryMap
	sh      *SyscallHandlers
	vfs     *VirtualFileSystem
	logger  zl.Logger
}
func NewNativeMemory(mu uc.Unicorn, mem *MemoryMap, sh *SyscallHandlers, vfs *VirtualFileSystem, logger zl.Logger) *NativeMemory {
	nm := &NativeMemory{
		mu: mu,
		mem:mem,sh: sh,
		vfs: vfs,
		logger: logger,
	}
	nm.sh.SetHandler(0x2D, "brk", 1, nm.handeBrk)
	nm.sh.SetHandler(0x5B, "munmap", 2, nm.handleMunmap)
	nm.sh.SetHandler(0x7D, "mprotect", 3, nm.handleMprotect)
	nm.sh.SetHandler(0xC0, "mmap2", 6, nm.handleMmap2)
	nm.sh.SetHandler(0xDC, "madvise", 3, nm.handleMadvise)
	return nm
}
func (nm *NativeMemory) handeBrk(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	//TODO: set errno
	//TODO: implement
	jx := int64(-1)
	return uint64(jx), true
}
func (nm *NativeMemory) handleMunmap(mu uc.Unicorn, args ...uint64) (uint64, bool){
	addr, len_in := args[0], args[1]
	err := nm.mem.Unmap(addr, len_in)
	if err != nil {
		ret := int64(-1)
		return uint64(ret), true
	}
	return 0, true
}
func (nm *NativeMemory) handleMprotect(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	addr, len_in, prot := args[0], args[1], args[2]
	err := nm.mem.Protect(addr, len_in, int(prot))
	if err != nil {
		ret := int64(-1)
		return uint64(ret), true
	}
	return 0, true
}
func (nm *NativeMemory) handleMmap2(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	addr, length, prot, flags, fd, offset := args[0], args[1], args[2], args[3], args[4], args[5]
	_ = flags
	var (
		res uint64
		err error
	)
	if fd != 0xffffffff {
		if fd <= 2 {
			panic("not implemented")
		}
		if !nm.vfs.pcb.HasFd(uintptr(fd)) {
			panic(ErrNotImplemented)
		}
		vf := nm.vfs.pcb.fds[uintptr(fd)]
		res, err = nm.mem.Map(addr, length, int(prot), vf, offset)
	}else{
		res, err = nm.mem.Map(addr, length, int(prot), nil, 0)
	}
	if err != nil {
		nm.logger.Debug().Err(err).Msg("mmap got error!")
		return 0, true
	}
	nm.logger.Debug().Msgf("mmap return 0x%08X", res)
	return res, true
}
func (nm *NativeMemory) handleMadvise(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	//We don't need your advise.
	return 0, true
}

















