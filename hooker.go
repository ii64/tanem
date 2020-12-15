package emulator

import (
	"fmt"
	"bytes"
	bin "encoding/binary"
	zl  "github.com/rs/zerolog"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
	ks  "github.com/keystone-engine/keystone/bindings/go/keystone"
)

type HookerCallback func(NativeMethodContext)error
type Hooker struct {
	emu    *Emulator
	ks     *ks.Keystone
	logger zl.Logger
	size   uint64

	currId      uint64
	hookMagic   uint64
	hookStart   uint64
	hookCurrent uint64
	hooks map[uint64]HookerCallback
}
func NewHooker(emu *Emulator, logger zl.Logger, base, size uint64) *Hooker {
	hk := &Hooker{
		emu: emu,
		logger: logger,
		size: size,
		currId: 0xFF00,
		hooks: map[uint64]HookerCallback{},
	}
	ks, err := ks.New(ks.ARCH_ARM, ks.MODE_THUMB)
	if err != nil {
		hk.logger.Debug().Err(err).Msg("keystone init failed")
		return nil
	}
	hk.logger.Debug().Msg("keystone ok")
	hk.ks = ks
	hk.hookMagic   = base
	hk.hookStart   = base + 4
	hk.hookCurrent = hk.hookStart
	hk.emu.Mu.HookAdd(uc.HOOK_CODE, hk.hook, hk.hookStart, hk.hookStart + size)
	return hk
}
func (hk *Hooker) getNextId() uint64 {
	curId := hk.currId
	hk.currId = hk.currId + 1
	return curId
}
func (hk *Hooker) writeFunction(f HookerCallback) (uint64, error) {
	hookId := hk.getNextId()
	hookAddr := hk.hookCurrent
	addrHex := fmt.Sprintf("0x%x", hookId)
	asm := "PUSH {R4,LR}\n" +
	        "MOV R4, #" + addrHex + "\n" +
	        "IT AL\n" +
	        "POP {R4,PC}"
//	fmt.Println(asm)
	code, v, ok := hk.ks.Assemble(asm, 0)
	if !ok {
		hk.logger.Debug().
			Str("hAddr", addrHex).
			Msgf("keystone cannot write hook function %+v", v)
		return 0, ErrAsmFailed
	}
	if v != 4 {
		hk.logger.Debug().
			Str("hAddr", addrHex).
			Int("len", len(code)).
			Bytes("code", code).
			Err(ErrUnexpectedAsmLength).
			Msgf("invalid asm length %+#v", v)
		return 0, ErrUnexpectedAsmLength
	}
	hk.logger.Debug().
		Str("hAddr", addrHex).
		Msg("write_function")
	hk.hookCurrent   = hk.hookCurrent + uint64(len(code))
	hk.hooks[hookId] = f
	return hookAddr, nil
}
func maxFFunc(m map[uint64]HookerCallback) (r uint64) {
	for k, _ := range m {
		if k > r {
			r = k
		}
	}
	return r
}
func (hk *Hooker) WriteFunctionTable(table map[uint64]HookerCallback) (uint64, uint64) {
	var (
		index   uint64
		address uint64
		err error
	)
	// First, we write every function and store its result address.
	indexMax := maxFFunc(table)
	hookMap := map[uint64]uint64{}
	for k, vf := range table {
		address, err = hk.writeFunction(vf)
		if err != nil {
			continue
		}
		hookMap[k] = address
	}
	// Then we write the function table.
	tableBytes := []byte{}
	tableAddr := hk.hookCurrent
	for index = 0; index < indexMax; index++ {
		if o, exist := hookMap[index]; exist {
			address = o
		}else{
			address = 0
		}
		by := make([]byte, 4)
		bin.LittleEndian.PutUint32(by,uint32(address + 1)) // explicit
		tableBytes = append(tableBytes, by...)
	}
	hk.emu.Mu.MemWrite(tableAddr, tableBytes)
	hk.hookCurrent = hk.hookCurrent + uint64(len(tableBytes))
	// Then we write the a pointer to the table.
	ptrAddr := hk.hookCurrent
	by := make([]byte, 4)
	bin.LittleEndian.PutUint32(by, uint32(tableAddr))
	hk.emu.Mu.MemWrite(ptrAddr, by)
	hk.hookCurrent = hk.hookCurrent + 4
	return ptrAddr, tableAddr
}
func (hk *Hooker) hook(mu uc.Unicorn, addr uint64, size uint32) {
	code, err := hk.emu.Mu.MemRead(addr, uint64(size))
	if err != nil {
		//hk.logger.Debug().Err(err).Msg("hook memread failed")
		return
	}
	// "IT AL"
	cmp := []byte{0xE8, 0xBF}
	if size != 2 || bytes.Compare(code, cmp) != 0 {
		//hk.logger.Debug().Err(ErrUnexpectedAsmLength).Msg("not 'IT AL' instruction")
		return
	}
	// Find hook.
	hookId, err := hk.emu.Mu.RegRead(uc.ARM_REG_R4)
	if err != nil {
		return
	}
	hookFunc := hk.hooks[hookId]
	err = hookFunc(NativeMethodContext{hk.emu, mu})
	if err != nil {
		hk.logger.Info().Err(err).Msg("hook function callback error, stopping emulation")
		mu.Stop()
		return
	}
}










	










