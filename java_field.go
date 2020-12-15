package emulator

import (
	log "github.com/rs/zerolog/log"
)

func AccessibleObject() *javaClass {
	jo := JavaClassDef()
	jo.SetJvmName("java/lang/reflect/AccessibleObject")
	jo.AddMethod(
		JavaMethodDef("setAccessible", false).
		Sig("(Z)V").
		Callback(func(ctx *MethodContext){
			log.Debug().Msg("AccessibleObject setAccessible call skip")
		}))
	return jo
}

func Field(parent *javaClass, fieldName string) *javaClass {
	_=parent
	jo := JavaClassDef()
	jo.SetJvmSuper(AccessibleObject())

	jo.SetJvmName("java/lang/reflect/Field")
	jo.AddMethod(
		JavaMethodDef("get", false).
		Args("jobject").
		Sig("(Ljava/lang/Object;)Ljava/lang/Object;").
		Callback(func(ctx *MethodContext){
			self := ctx.GetArg(0).(*javaClass)
			log.Debug().Msgf("Field.get(%T)", self)
			ctx.Return(self.FindFieldByNameOnly(fieldName))
		}))
	return jo
}