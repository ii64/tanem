package emulator

import (
	zl  "github.com/rs/zerolog"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)
type InterruptHandler struct {
	mu       uc.Unicorn
	handler  map[uint32]func(uc.Unicorn)

	logger   zl.Logger
}
func NewInterruptHandler(mu uc.Unicorn) *InterruptHandler {
	ih := &InterruptHandler{
		mu: mu,
		handler: map[uint32]func(uc.Unicorn){},
	}
	ih.mu.HookAdd(uc.HOOK_INTR, ih.hookInterrupt, 1, 0)
	return ih
}
func (ih *InterruptHandler) SetLogger(logger zl.Logger) {
	ih.logger = logger
}
func (ih *InterruptHandler) SetHandler(intno uint32, handler func(uc.Unicorn)) {
	ih.handler[intno] = handler
}
func (ih *InterruptHandler) hookInterrupt(mu uc.Unicorn, intno uint32) {
	cb, exist := ih.handler[intno]
	if !exist {
		regx, err := mu.RegRead(uc.ARM_REG_PC)
		ih.logger.Debug().
			Err(err).
			Msg("reading reg PC")
		ih.logger.Debug().
			Uint32("intno", intno).
			Uint64("pc", regx).
			Msg("unhandled interrupt")
		ih.logger.Debug().Err(mu.Stop()).Msg("stopping emulation")
		//panic(fmt.Sprintf("Unhandled interrupt %d at %x, stopping emulation", intno, regx))
		return
	}
	cb(mu)
}