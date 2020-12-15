package emulator

import (
//	zl  "github.com/rs/zerolog"
	"bytes"
	bin "encoding/binary"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

func ReadPtr(mu uc.Unicorn, address uint64) (uint64, error) {
	var sz uint64 = 4
	by, err := mu.MemRead(address, sz)
	if err != nil {
		return 0, err
	}
	res := bin.LittleEndian.Uint32(by)
	return uint64(res), nil
}
func ReadByteArray(mu uc.Unicorn, address, size uint64) ([]byte, error) {
	return mu.MemRead(address, size)
}
func ReadUtf8(mu uc.Unicorn, address uint64) ([]byte, error) {
	var (
		buffAddr     uint64 = address
		buffReadSize uint64 = 32
	)
	bb := []byte{}
	foundNullPos := false
	nullPos := 0
	for {
		if foundNullPos {
			break
		}
		by, err := mu.MemRead(buffAddr, buffReadSize)
		if err != nil {
			return nil, err
		}
		if bytes.Contains(by, []byte{0x0}) {
			nullPos = len(by) + bytes.Index(by, []byte{0x0})
			foundNullPos = true
		}
		bb = append(bb, by...)
	}
	return bb[:nullPos], nil
}
func ReadUints(mu uc.Unicorn, address uint64, num int) ([]uint64, error) {
	var r []uint64
	for i := 0; i < num; i++ {
		by, err := mu.MemRead(address, 4)
		if err != nil {
			return nil, err
		}
		address = address + 4
		r = append(r, uint64(bin.LittleEndian.Uint32(by)))
	}
	return r, nil
}
func WriteUtf8(mu uc.Unicorn, address uint64, val []byte) error {
	val = append(val, 0)
	return mu.MemWrite(address, val)
}
func WriteUints(mu uc.Unicorn, address uint64, nums []uint64) (err error) {
	for _, num := range nums {
		by := make([]byte, 4)
		bin.LittleEndian.PutUint32(by, uint32(num))
		err = mu.MemWrite(address, by)
		if err != nil {
			return err
		}
		address = address + 4
	}
	return nil
}


type RegistryContext struct {
	R0,R1,R2,R3,R4,R5,R6,R7,R8,R9,R10,R11,R12,SP,LR,PC,CPSR uint64
}
func RegContextSave(mu uc.Unicorn) (*RegistryContext, error) {
	var (
		tmp uint64
		err error
	)
	ctx := &RegistryContext{}
	tmp, err = mu.RegRead(uc.ARM_REG_R0)
	if err != nil {
		return nil, err
	}
	ctx.R0 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R1)
	if err != nil {
		return nil, err
	}
	ctx.R1 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R2)
	if err != nil {
		return nil, err
	}
	ctx.R2 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R3)
	if err != nil {
		return nil, err
	}
	ctx.R3 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R4)
	if err != nil {
		return nil, err
	}
	ctx.R4 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R5)
	if err != nil {
		return nil, err
	}
	ctx.R5 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R6)
	if err != nil {
		return nil, err
	}
	ctx.R6 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R7)
	if err != nil {
		return nil, err
	}
	ctx.R7 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R8)
	if err != nil {
		return nil, err
	}	
	ctx.R8 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R9)
	if err != nil {
		return nil, err
	}
	ctx.R9 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R10)
	if err != nil {
		return nil, err
	}
	ctx.R10 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R11)
	if err != nil {
		return nil, err
	}
	ctx.R11 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_R12)
	if err != nil {
		return nil, err
	}	
	ctx.R12 = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_SP)
	if err != nil {
		return nil, err
	}
	ctx.SP = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_LR)
	if err != nil {
		return nil, err
	}
	ctx.LR = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_PC)
	if err != nil {
		return nil, err
	}
	ctx.PC = tmp
	tmp, err = mu.RegRead(uc.ARM_REG_CPSR)
	if err != nil {
		return nil, err
	}
	ctx.CPSR = tmp
	return ctx, nil
}

func RegContextRestore(mu uc.Unicorn, ctx *RegistryContext) error {
	if ctx == nil {
		return nil
	}
	var err error
	err = mu.RegWrite(uc.ARM_REG_R0, ctx.R0)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R1, ctx.R1)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R2, ctx.R2)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R3, ctx.R3)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R4, ctx.R4)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R5, ctx.R5)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R6, ctx.R6)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R7, ctx.R7)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R8, ctx.R8)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R9, ctx.R9)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R10, ctx.R10)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R11, ctx.R11)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_R12, ctx.R12)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_SP, ctx.SP)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_LR, ctx.LR)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_PC, ctx.PC)
	if err != nil {
		return err
	}
	err = mu.RegWrite(uc.ARM_REG_CPSR, ctx.CPSR)
	if err != nil {
		return err
	}
	return nil
}

















