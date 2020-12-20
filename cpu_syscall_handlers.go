package emulator

import (
	zl  "github.com/rs/zerolog"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

type SyscallCallback func(uc.Unicorn, ...uint64)(uint64, bool)

type SyscallHandler struct {
	Idx uint64
	Name string
	ArgCount int
	Callback SyscallCallback
}
func NewSyscallHandler() *SyscallHandler {
	return &SyscallHandler{}
}

type SyscallHandlers struct {
	ih        *InterruptHandler
	handler   map[uint64]*SyscallHandler

	logger   zl.Logger
}
func NewSyscallHandlers(ih *InterruptHandler) *SyscallHandlers {
	if ih == nil {
		return nil
	}
	sh := &SyscallHandlers{
		ih: ih,
		handler: map[uint64]*SyscallHandler{},
	}
	sh.ih.SetHandler(2, sh.handleSyscall)
	return sh
}
func (sh *SyscallHandlers) SetLogger(logger zl.Logger) {
	sh.logger = logger
}
func (sh *SyscallHandlers) SetHandler(idx uint64, name string, argCount int, callback SyscallCallback) {
	s := NewSyscallHandler()
	s.Idx = idx
	s.Name = name
	s.ArgCount = argCount
	s.Callback = callback	
	sh.handler[idx] = s
}
func (sh *SyscallHandlers) handleSyscall(mu uc.Unicorn) {
	idx, err := mu.RegRead(uc.ARM_REG_R7)
	if err != nil {
		sh.logger.Debug().Err(err).Msg("read reg R7 failed")
	}
	lr, err := mu.RegRead(uc.ARM_REG_LR)
	if err != nil {
		sh.logger.Debug().Err(err).Msg("read reg LR failed")
	}
	sh.logger.Info().Msgf("syscall %s lr=0x%08X", ConvHex("0x%X",idx), lr)
	pc, err := mu.RegRead(uc.ARM_REG_PC)
	if err != nil {
		sh.logger.Debug().Err(err).Msg("read reg PC failed")
	}
	var args []uint64
	for reg_idx := uc.ARM_REG_R0; reg_idx < (uc.ARM_REG_R6+1); reg_idx++ {
		arg, err := mu.RegRead(reg_idx)
		if err != nil {
			sh.logger.Debug().Int("reg", reg_idx).Err(err).Msg("read reg failed")
			continue
		}
		args = append(args, arg)
	}
	if h, exist := sh.handler[idx]; exist {
		sh.logger.Debug().
			Str("id", ConvHex("0x%X",idx)).
			Str("name", h.Name).
			Str("args", ConvHex("0x%X",args)).
			Str("pc", ConvHex("0x%08X",pc)).
			Msg("executing syscall")
		ret, hasRet := h.Callback(mu, args[:h.ArgCount]...)
		if hasRet {
			err = mu.RegWrite(uc.ARM_REG_R0, ret)
		}
		sh.logger.Debug().
			Str("id", ConvHex("0x%X",idx)).
			Str("name", h.Name).
			Bool("hasRet", hasRet).
			Str("pc", ConvHex("0x%08X",pc)).
			Str("ret", ConvHex("0x%X",ret)).
			Err(err).
			Msg("write return")
	}else{
		sh.logger.Debug().
			Str("idx", ConvHex("0x%X",idx)).
			Str("args", ConvHex("0x%X",args)).
			Str("pc", ConvHex("0x%08X",pc)).
			Msg("unhandled syscall")
		sh.logger.Debug().Err(mu.Stop()).Msg("stopping emulation")
	}
}
