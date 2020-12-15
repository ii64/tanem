package emulator

import (
	"strings"
)

type javaClass struct {
	JvmId        uint64
	JvmName      string
	JvmFields    map[uint64]*javaField
	JvmMethods   map[uint64]*javaMethod
	JvmIgnore    bool
	JvmSuper    *javaClass

	Class       *jClass
}
func JavaClassDef() *javaClass {
	return &javaClass{
		JvmId: NextClassId(),
		JvmFields:  map[uint64]*javaField{},
		JvmMethods: map[uint64]*javaMethod{},
	}
}
func (jcd *javaClass) GetJniDescription() string {
	return jcd.JvmName
}
func (jcd *javaClass) GetJvmId() uint64 {
	return jcd.JvmId
}
func (jcd *javaClass) GetJvmName() string {
	return jcd.JvmName
}
func (jcd *javaClass) SetJvmSuper(cls *javaClass) {
	jcd.JvmSuper = cls
}
func (jcd *javaClass) SetJvmName(s string) {
	jcd.JvmName = s
}
func (jcd *javaClass) GetMethodByName() {

}
func (jcd *javaClass) GetJvmIgnore() bool {
	return jcd.JvmIgnore
}
func (jcd *javaClass) GetJvmSuper() *javaClass {
	return jcd.JvmSuper
}
//
func (jcd *javaClass) AddField(jf *javaField) {
	jcd.JvmFields[jf.JvmId]  = jf
}
func (jcd *javaClass) AddMethod(jm *javaMethod) {
	jcd.JvmMethods[jm.JvmId] = jm
}

//
func (jcd *javaClass) FindMethod(name, signature string) *javaMethod {
	for _, met := range jcd.JvmMethods {
		if met.Name == name && met.Signature == signature {
			return met
		}
	}
	if jcd.JvmSuper != nil {
		return jcd.JvmSuper.FindMethod(name, signature)
	}
	return nil
}
//用于支持java反射，java反射签名都没有返回值
//@param signature_no_ret something like (ILjava/lang/String;) 注意，没有返回值
func (jcd *javaClass) FindMethodSigWithNoRet(name, signature string) *javaMethod {
	for _, met := range jcd.JvmMethods {
		if met.Name == name && strings.HasPrefix(met.Signature, signature) {
			return met
		}
	}
	if jcd.JvmSuper != nil {
		return jcd.JvmSuper.FindMethodSigWithNoRet(name, signature)
	}
	return nil
}
//
func (jcd *javaClass) FindMethodById(jvmId uint64) *javaMethod {
	if met, exist := jcd.JvmMethods[jvmId]; exist {
		return met
	}
	if jcd.JvmSuper != nil {
		return jcd.JvmSuper.FindMethodById(jvmId)
	}
	return nil
}
//
func (jcd *javaClass) FindField(name, signature string, isStatic bool) *javaField {
	for _, field := range jcd.JvmFields {
		if field.Name == name && field.signature == signature && field.isStatic == isStatic {
			return field
		}
	}
	if jcd.JvmSuper != nil {
		return jcd.JvmSuper.FindField(name, signature, isStatic)
	}
	return nil
}
func (jcd *javaClass) FindFieldById(jvmId uint64) *javaField {
	if field, exist := jcd.JvmFields[jvmId]; exist {
		return field
	}
	if jcd.JvmSuper != nil {
		return jcd.JvmSuper.FindFieldById(jvmId)
	}
	return nil
}
func (jcd *javaClass) FindFieldByNameOnly(name string) *javaField {
	for _, field := range jcd.JvmFields {
		if field.Name == name {
			return field
		}
	}
	if jcd.JvmSuper != nil {
		return jcd.JvmSuper.FindFieldByNameOnly(name)
	}
	return nil
}




