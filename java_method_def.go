package emulator

type MethodContext struct {
	emu *Emulator
	retval interface{}
}
func (ctx *MethodContext) GetEmu() *Emulator {
	return ctx.emu
}
func (ctx *MethodContext) GetArg(i int) interface{} {
	return nil
}
func (ctx *MethodContext) GetArgMethodId(i int) uint64 {
	return ctx.GetArg(i).(uint64)
}
func (ctx *MethodContext) GetArgString(i int) string {
	return ctx.GetArg(i).(string)
}
func (ctx *MethodContext) GetArgObject(i int) *javaClass {
	return ctx.GetArg(i).(*javaClass)
}
func (ctx *MethodContext) GetArgArrayObject(i int) []*javaClass {
	return ctx.GetArg(i).([]*javaClass)
}
func (ctx *MethodContext) Call() {
	// wip
}
func (ctx *MethodContext) GetReturn() interface{} {
	return ctx.retval
}
func (ctx *MethodContext) ReturnString(s string) {
	ctx.Return(s)
}
func (ctx *MethodContext) Return(v interface{}) {
	ctx.retval = v
}

type MethodFunction func(*MethodContext)

type javaMethod struct {
	JvmId       uint64
	Name        string
	Signature   string
	cb          MethodFunction

	native      bool
	nativeAddr  uint64

	argList     []string
	ignore      bool
	modifier    uint64
}
/*
Usage:
JavaMethodDef("getString", false).
	Sig("(Landroid/content/ContentResolver;Ljava/lang/String;)Ljava/lang/String;").
	Args([]string{"jobject", "jstring"}).
	Callback()

JavaMethodDef("beHealthy", true).
	Sig("(II[B)[B").

*/
func JavaMethodDef(name string, native bool) *javaMethod {
	return &javaMethod{
		JvmId: NextMethodId(),
		Name: name,
		native: native,
	}
}
//WIP:will set callback,args that calls native function
func (jm *javaMethod) Native() *javaMethod {
	jm.native = true
	return jm
}
func (jm *javaMethod) Modifier(m uint64) *javaMethod {
	jm.modifier = m
	return jm
}
func (jm *javaMethod) Sig(sig string) *javaMethod {
	jm.Signature = sig
	return jm
}
func (jm *javaMethod) Args(args ...string) *javaMethod {
	jm.argList = args
	return jm
}
func (jm *javaMethod) Ignore() *javaMethod {
	jm.ignore = true
	return jm
}
func (jm *javaMethod) Callback(f MethodFunction) *javaMethod {
	if jm.native {
		return nil
	}
	jm.cb = f
	return jm
}