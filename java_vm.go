package emulator

import (
	"github.com/pkg/errors"
	bin  "encoding/binary"
	zl   "github.com/rs/zerolog"
//	uc   "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

type JavaVM struct {
	emu     *Emulator
	jcl     *JavaClassLoader
	hooker  *Hooker

	addressPtr uint64
	address    uint64

	logger zl.Logger

	JniEnv  *JniEnv
}

func NewJavaVM(emu *Emulator, jcl *JavaClassLoader, hooker *Hooker, logger zl.Logger) *JavaVM {
	jvm := &JavaVM{
		emu: emu,
		jcl: jcl,
		hooker: hooker,
	}
	addrPtr, addr := hooker.WriteFunctionTable(map[uint64]HookerCallback{
		3: jvm.destroyJavaVm,
		4: jvm.attachCurrentThread,
		5: jvm.detachCurrentThread,
		6: jvm.getEnv,
		7: jvm.attachCurrentThread,
	})
	jvm.addressPtr = addrPtr
	jvm.address = addr
	jvm.JniEnv = NewJniEnv(emu, jcl, hooker)
	return jvm
}
func (jvm *JavaVM) AddrPtr() uint64 {
	return jvm.addressPtr
}
func (jvm *JavaVM) Addr() uint64 {
	return jvm.address
}
func (jvm *JavaVM) destroyJavaVm(ctx NativeMethodContext) error {
	return ErrNotImplemented
}
func (jvm *JavaVM) attachCurrentThread(ctx NativeMethodContext) error {
	return ErrNotImplemented
}
func (jvm *JavaVM) detachCurrentThread(ctx NativeMethodContext) error {
	// TODO: NooOO idea.	
	return nil
}
func (jvm *JavaVM) getEnv(ctx NativeMethodContext) error {
	args := ctx.GetArgs(3)
	java_vm, env, version := args[0], args[1], args[2]
	jvm.logger.Debug().Msgf("java_vm: 0x%08x", java_vm)
	jvm.logger.Debug().Msgf("env: 0x%08x", env)
	jvm.logger.Debug().Msgf("version: 0x%08x", version)	
	by := make([]byte, 4)
	bin.LittleEndian.PutUint32(by, uint32(jvm.JniEnv.addressPtr))
	err := ctx.Mu().MemWrite(env, by)
	if err != nil {
		jvm.logger.Debug().Uint64("envAddr", env).Err(err).Msg("cannot write JniEnv ptr")
		return errors.Wrap(ctx.Return(JNI_ERR), "failed to write return GetEnv()")
	}
	return errors.Wrap(ctx.Return(JNI_OK), "failed to write return GetEnv()")
}














