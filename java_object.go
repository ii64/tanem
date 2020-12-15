package emulator

func Object() *javaClass {
	jo := JavaClassDef()
	jo.SetJvmName("java/lang/Object")
	return jo
}