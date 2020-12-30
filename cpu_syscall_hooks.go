package emulator

import (
	zl  "github.com/rs/zerolog"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

var (
	//TIMEOFDAY
	OVERRIDE_TIMEOFDAY      = false
	OVERRIDE_TIMEOFDAY_SEC  = 0
	OVERRIDE_TIMEOFDAY_USEC = 0
	//CLOCK
	OVERRIDE_CLOCK	  = false
	OVERRIDE_CLOCK_TIME     = 0
)

type SyscallHooks struct {
	mu       uc.Unicorn
	sh       *SyscallHandlers
	logger   zl.Logger
}
func NewSyscallHooks(mu uc.Unicorn, sh *SyscallHandlers) *SyscallHooks {
	s := &SyscallHooks{
		mu: mu,
		sh: sh,
	}
	//system call table
	s.sh.SetHandler(0x2, "fork", 0, s.forkHandle)
	s.sh.SetHandler(0x0B, "execve", 3, s.execveHandle)
	s.sh.SetHandler(0x14, "getpid", 0, s.getpidHandle)
	s.sh.SetHandler(0x1A, "ptrace", 4, s.ptraceHandle)
	s.sh.SetHandler(0x25, "kill", 2, s.killHandle)
	s.sh.SetHandler(0x2A, "pipe", 1, s.pipeHandle)
	s.sh.SetHandler(0x43, "sigaction", 3, s.sigactionHandle)
	s.sh.SetHandler(0x4E, "gettimeofday", 2, s.gettimeofdayHandle)
	s.sh.SetHandler(0x72, "wait4", 4, s.wait4Handle)
	s.sh.SetHandler(0x74, "sysinfo", 1, s.sysinfoHandle)
	s.sh.SetHandler(0x78, "clone", 5, s.cloneHandle)
	s.sh.SetHandler(0xAC, "prctl", 5, s.prctlHandle)
	s.sh.SetHandler(0xAF, "sigprocmask", 3, s.sigprocmaskHandle)
	s.sh.SetHandler(0xBA, "sigaltstack", 2, s.sigaltstackHandle)
	s.sh.SetHandler(0xBE, "vfork", 0, s.vforkHandle)
	s.sh.SetHandler(0xC7, "getuid32", 0, s.getuid32Handle)
	s.sh.SetHandler(0xE0, "gettid", 0, s.gettidHandle)
	s.sh.SetHandler(0xF0, "futex", 6, s.futexHandle)
	s.sh.SetHandler(0x10c, "tgkill", 3, s.tgkillHandle)
	s.sh.SetHandler(0x107, "clock_gettime", 2, s.clock_gettimeHandle)
	s.sh.SetHandler(0x119, "socket", 3, s.socketHandle)
	s.sh.SetHandler(0x11a, "bind", 3, s.bindHandle)
	s.sh.SetHandler(0x11b, "connect", 3, s.connectHandle)
	s.sh.SetHandler(0x126, "setsockopt", 5, s.setsockoptHandle)
	s.sh.SetHandler(0x159, "getcpu", 3, s.getcpuHandle)
	s.sh.SetHandler(0x166, "dup3", 3, s.dup3Handle)
	s.sh.SetHandler(0x167, "pipe2", 2, s.pipe2Handle)
	s.sh.SetHandler(0x178, "process_vm_readv", 6, s.process_vm_readvHandle)
	s.sh.SetHandler(0x180, "getrandom", 3, s.getrandomHandle)
	s.sh.SetHandler(0xf0002, "ARM_cacheflush", 0, s.ARM_cacheflushHandle)
	s.sh.SetHandler(0xa2, "nanosleep", 2, s.nanosleepHandle)
	return s
}
func (s *SyscallHooks) SetLogger(logger zl.Logger) {
	s.logger = logger
}
// syscall fork
func (s *SyscallHooks) forkHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	s.logger.Debug().Msg("fork called")
	return 0, true // child process..
}
// syscall execve
func (s *SyscallHooks) execveHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall getpid
func (s *SyscallHooks) getpidHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall ptrace
func (s *SyscallHooks) ptraceHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall kill
func (s *SyscallHooks) killHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall pipe
func (s *SyscallHooks) pipeHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall sigaction
func (s *SyscallHooks) sigactionHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall gettimeofday
func (s *SyscallHooks) gettimeofdayHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall wait4
func (s *SyscallHooks) wait4Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall sysinfo
func (s *SyscallHooks) sysinfoHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall clone
func (s *SyscallHooks) cloneHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall prctl
func (s *SyscallHooks) prctlHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall sigprocmask
func (s *SyscallHooks) sigprocmaskHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall signalstack
func (s *SyscallHooks) sigaltstackHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0,true
}
// syscall vfork
func (s *SyscallHooks) vforkHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall getuid32
func (s *SyscallHooks) getuid32Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall gettid
func (s *SyscallHooks) gettidHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall futex
func (s *SyscallHooks) futexHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	uaddr, op, val, timeout, uaddr2, val3 := args[0], args[1], args[2], args[3], args[4], args[5]
	_,_,_=timeout,uaddr2,val3
	v, err := mu.MemRead(uaddr, 4)
	if err != nil {
		s.logger.Debug().Msg("futex uaddr read failed")
	}
	uaddrVal := LE_BytesToUint(v)
	cmd := op & FUTEX_CMD_MASK
	s.logger.Debug().
		Str("op", ConvHex("%08X", op)).
		Str("cmd", ConvHex("%X", cmd)).
		Str("*uaddr", ConvHex("%08X", uaddrVal)).
		Str("val", ConvHex("%08X", val)).
		Msg("futex call")
	if cmd == FUTEX_WAIT || cmd == FUTEX_WAIT_BITSET {
		if uaddrVal == val {
			//sorry, you can use recoveer anyway..
			panic("ERROR!!! FUTEX_WAIT or FUTEX_WAIT_BITSET dead lock !!! *uaddr == val, impossible for single thread program!!!")
		}
		return 0, true
	}else if cmd == FUTEX_WAKE {
		return 0, true
	}else if cmd == FUTEX_FD {
		panic(ErrNotImplemented)
	}else if cmd == FUTEX_REQUEUE {
		panic(ErrNotImplemented)
	}else if cmd == FUTEX_CMP_REQUEUE {
		panic(ErrNotImplemented)
	}else if cmd == FUTEX_WAKE_BITSET {
		return 0, true
	}else{
		panic(ErrNotImplemented)
	}
	return 0, true
}
// syscall tgkill
func (s *SyscallHooks) tgkillHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall clock_gettime
func (s *SyscallHooks) clock_gettimeHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall socket
func (s *SyscallHooks) socketHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall bind
func (s *SyscallHooks) bindHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall connect
func (s *SyscallHooks) connectHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall setsockopt
func (s *SyscallHooks) setsockoptHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall getcpu
func (s *SyscallHooks) getcpuHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall dup3
func (s *SyscallHooks) dup3Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall pipe2
func (s *SyscallHooks) pipe2Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall process_vm_readv
func (s *SyscallHooks) process_vm_readvHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall getrandom
func (s *SyscallHooks) getrandomHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall ARM_cacheflush
func (s *SyscallHooks) ARM_cacheflushHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall nanosleep
func (s *SyscallHooks) nanosleepHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}










