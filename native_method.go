package emulator

import (
	"fmt"
	"github.com/pkg/errors"
	bin  "encoding/binary"
	log  "github.com/rs/zerolog/log"
	uc   "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

func NativeWriteArgs(emu *Emulator, args ...interface{}) error {
	var err error
	amount := len(args)
	if amount == 0 {
		return nil
	}

	if amount >= 1 {
		err = NativeWriteArgRegister(emu, uc.ARM_REG_R0, args[0])
		if err != nil {
			return errors.Wrap(err, "write reg r0 failed")
		}
	}
	if amount >= 2 {
		err = NativeWriteArgRegister(emu, uc.ARM_REG_R1, args[1])
		if err != nil {
			return errors.Wrap(err, "write reg r1 failed")
		}
	}
	if amount >= 3 {
		err = NativeWriteArgRegister(emu, uc.ARM_REG_R2, args[2])
		if err != nil {
			return errors.Wrap(err, "write reg r2 failed")
		}
	}
	if amount >= 4 {
		err = NativeWriteArgRegister(emu, uc.ARM_REG_R3, args[3])
		if err != nil {
			return errors.Wrap(err, "write reg r3 failed")
		}
	}
	if amount >= 5 {
		spStart, err := emu.Mu.RegRead(uc.ARM_REG_SP)
		if err != nil {
			log.Debug().Err(err).Msg("failed to read registry ARM_REG_SP writeArgs")
			return err
		}
		spCurrent := spStart - STACK_OFFSET
		spCurrent = spCurrent - (4 * (uint64(amount) - 4))
		spEnd := spCurrent
		for _, arg := range args[4:] {
			ptr := NativeTranslateArg(emu, arg)
			byt := make([]byte, 4)
			bin.LittleEndian.PutUint32(byt, uint32(ptr))
			err = emu.Mu.MemWrite(spCurrent, byt)
			if err != nil {
				log.Debug().Err(err).Msg("failed to write sp reg args val")
				return errors.Wrap(err, "failed to write sp reg arg val")
			}
			spCurrent = spCurrent + 4
		}
		err = emu.Mu.RegWrite(uc.ARM_REG_SP, spEnd)
		if err != nil {
			log.Debug().Err(err).Msg("failed to write sp reg end address")
			return errors.Wrap(err, "failed to write sp reg end address")
		}
	}
	return nil
}

func NativeReadArgs(mu uc.Unicorn, argsCount int) []uint64 {
	// init with 0
	nativeArgs := make([]uint64, argsCount)
	if argsCount >= 1 {
		pt, err := mu.RegRead(uc.ARM_REG_R0)
		if err != nil {
			log.Debug().Err(err).Msg("failed to read reg r0")
		}
		nativeArgs[0] = pt
	}
	if argsCount >= 2 {
		pt, err := mu.RegRead(uc.ARM_REG_R1)
		if err != nil {
			log.Debug().Err(err).Msg("failed to read reg r1")
		}
		nativeArgs[1] = pt
	}
	if argsCount >= 3 {
		pt, err := mu.RegRead(uc.ARM_REG_R2)
		if err != nil {
			log.Debug().Err(err).Msg("failed to read reg r2")
		}
		nativeArgs[2] = pt
	}
	if argsCount >= 4 {
		pt, err := mu.RegRead(uc.ARM_REG_R3)
		if err != nil {
			log.Debug().Err(err).Msg("failed to read reg r2")
		}
		nativeArgs[3] = pt
	}
	sp, err := mu.RegRead(uc.ARM_REG_SP)
	if err != nil {
		log.Debug().Err(err).Msg("failed to read reg SP for reading args")
		return nativeArgs
	}
	sp = sp + STACK_OFFSET
	if argsCount >= 5 {
		var i uint64
		for i = 0; i < uint64(argsCount - 4); i++ {
			by, err := mu.MemRead(sp + (i * 4), 4)
			if err != nil {
				log.Debug().Uint64("addrSp", sp+(i*4)).Err(err).Msg("failed to read memory for reading arg")
				continue
			}
			ptr := uint64(bin.LittleEndian.Uint32(by))
			nativeArgs[4+int(i)] = ptr
		}
	}
	return nativeArgs
}

func NativeTranslateArg(emu *Emulator, val interface{}) uint64 {
	switch v := val.(type) {
	case uint64:
		return v
	case int:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case []byte:
		return emu.JavaVM.JniEnv.AddLocalReference(NewJObject(val.([]byte)))
	case *javaClass:
		return emu.JavaVM.JniEnv.AddLocalReference(NewJObject(val.(*javaClass)))
	}
	panic(fmt.Errorf("%w: unable to write response '%T' %#+v", val, val))
	return 0
}


func NativeWriteArgRegister(emu *Emulator, reg int, val interface{}) error {
	ptr := NativeTranslateArg(emu, val)
	return emu.Mu.RegWrite(reg, ptr)
}


type NativeMethodContext struct {
	emu *Emulator
	mu uc.Unicorn
}
func (nmc NativeMethodContext) Emu() *Emulator {
	return nmc.emu
}
func (nmc NativeMethodContext) Mu() uc.Unicorn {
	return nmc.mu
}
func (nmc NativeMethodContext) GetArgs(count int) []uint64 {
	return NativeReadArgs(nmc.mu, count)
}
func (nmc NativeMethodContext) Return(res uint64) error {
	return NativeWriteArgRegister(nmc.emu, uc.ARM_REG_R0, res)
}
func (nmc NativeMethodContext) Return2(low, high uint64) error {
	err := NativeWriteArgRegister(nmc.emu, uc.ARM_REG_R0, low)
	if err != nil {
		return err
	}
	err = NativeWriteArgRegister(nmc.emu, uc.ARM_REG_R1, high)
	if err != nil {
		return err
	}
	return nil
}








