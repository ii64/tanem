package emulator

type JniEnv struct {
	emu *Emulator
	jcl *JavaClassLoader
	hk  *Hooker

	addressPtr uint64
}

func NewJniEnv(emu *Emulator, jcl *JavaClassLoader, hk *Hooker) *JniEnv {
	je := &JniEnv{
		emu:emu, jcl:jcl, hk:hk,
	}
	return je
}
func (je *JniEnv) AddLocalReference(obj *jobject) uint64 {
	return 0
}
func (je *JniEnv) GetLocalReference(idx uint64) uint64 {
	return 0
}
func (je *JniEnv) ClearLocals() {
	return
}