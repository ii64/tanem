package emulator

import (
	"strings"
	log "github.com/rs/zerolog/log"
)

var (
	basicType = []string{"Z", "B", "C", "D", "F", "I", "J", "S"}
)

func isInBasicType(s string) bool {
	for _, v := range basicType {
		if v == s {
			return true
		}
	}
	return false
}

type jClass struct {
	*javaClass
	refJvmName string
	clazz *javaClass
}
func NewClass(clazz *javaClass) *jClass {
	jo := &jClass{
		clazz: clazz,
		refJvmName: clazz.JvmName,
	}
	jo.javaClass = JavaClassDef()
	jo.SetJvmName("java/lang/Class")
	jo.AddMethod(
		JavaMethodDef("getClassLoader", false).
		Sig("()Ljava/lang/ClassLoader;").
		Callback(jo.getClassLoader))
	jo.AddMethod(JavaMethodDef("getName", false).
		Sig("()Ljava/lang/String;").
		Callback(jo.getName))
	jo.AddMethod(JavaMethodDef("getCanonicalName", false).
		Sig("()Ljava/lang/String;").
		Callback(jo.getCanonicalName))
	jo.AddMethod(JavaMethodDef("getDeclaredField", false).
		Sig("(Ljava/lang/String;)Ljava/lang/reflect/Field;").
		Args("jstring").
		Callback(jo.getDeclaredField))
	jo.AddMethod(JavaMethodDef("getDeclaredMethod", false).
		Sig("(Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;").
		Args("jstring", "jobject").
		Callback(jo.getDeclaredMethod))
	return jo
}
//
func (jo *jClass) GetJniDescription() string {
	return jo.refJvmName
}
func (jo *jClass) GetClazz() *javaClass {
	return jo.clazz
}
//
func (jo *jClass) getClassLoader(ctx *MethodContext) {
	ctx.Return(jo.clazz)
}
func (jo *jClass) getName(ctx *MethodContext) {
	ctx.ReturnString(strings.Replace(jo.refJvmName, "/", ".", -1))
}
func (jo *jClass) getCanonicalName(ctx *MethodContext) {
	jo.getName(ctx)
	name := ctx.GetReturn().(string)
	if name[0] == '[' {
		dims := 0
		for _, ch := range name {
			if ch == '[' {
				dims++
			}else{
				break
			}
		}
		name = name[dims:]
		if name[0] == 'L' {
			name = name[1:]
		}
		for i := 0; i < dims; i++ {
			name = name + "[]"
		}
	}
	name = strings.Replace(name, "$", ".", -1)
	ctx.ReturnString(name)
}
func (jo *jClass) getDeclaredField(ctx *MethodContext) {
	name := ctx.GetArgString(0)
	log.Debug().
		Str("name", name).
		Msg("getDeclaredField")
	ctx.Return(Field(jo.clazz, name))
}
func (jo *jClass) getDeclaredMethod(ctx *MethodContext) {
	name := ctx.GetArgString(0)
	arrjobj := ctx.GetArgArrayObject(1)
	log.Debug().Msgf("getDeclaredMethod name:[%T]", name)
	sbuf := "("
	for _, item := range arrjobj {
		desc := item.GetJniDescription()
		if desc[0] == '[' || isInBasicType(desc) {
			sbuf = sbuf + desc
		}else{
			sbuf = sbuf + "L"
			sbuf = sbuf + desc
			sbuf = sbuf + ";"
		}
	}
	sbuf = sbuf + ")"
	met := jo.clazz.FindMethodSigWithNoRet(name, sbuf)
	reflected_method := Method(jo.clazz, met)
	log.Debug().Msgf("getDeclaredMethod return:[%T]", reflected_method)
	ctx.Return(reflected_method)
}






















