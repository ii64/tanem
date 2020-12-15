package emulator

import (
	log "github.com/rs/zerolog/log"
)

func Method(declaringClass *javaClass, met *javaMethod) *javaClass {
	jo := JavaClassDef()
	jo.SetJvmName("java/lang/reflect/Method")
	jo.SetJvmSuper(Executable())
	slot := met.JvmId
	jo.AddField(
		JavaFieldDef("slot").Sig("I").
		Value(slot))
	jo.AddField(
		JavaFieldDef("declaringClass").
		Sig("Ljava/lang/Class;").
		Value(declaringClass))
	jo.FindFieldByNameOnly("accessFlags").Value(met.modifier)
	jo.AddMethod(
		JavaMethodDef("getMethodModifiers", false).
		Sig("(Ljava/lang/Class;I)I").
		Args("jobject", "jint").
		Callback(func(ctx *MethodContext){
			clazz := ctx.GetArg(0).(*javaClass)
			id := ctx.GetArgMethodId(1)
			method := clazz.FindMethodById(id)
			log.Debug().Msgf("Method.getMethodModifiers(%s, %s)",clazz.GetJvmName(), method.Name)
			ctx.Return(method.modifier)
		}))
	jo.AddMethod(
		JavaMethodDef("invoke", false).
		Sig("(Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;").
		Args("jobject", "jobject").
		Callback(func(ctx *MethodContext){
			obj := ctx.GetArg(0)
			args := ctx.GetArg(1)
			log.Debug().Msgf("Method.invoke(%T, %T)", obj, args)
			if obj == nil {
				// static method
				met.cb(ctx)
			}else{
				met.cb(ctx)
			}
		}))
	return jo
}












