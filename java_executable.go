package emulator

func Executable() *javaClass {
	jo := JavaClassDef()
	jo.SetJvmName("java/lang/reflect/Executable")
	jo.AddField(
		JavaFieldDef("accessFlags").
		Sig("I"))
	return jo
}