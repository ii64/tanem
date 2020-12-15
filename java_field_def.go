package emulator

type FieldValue interface{}
type javaField struct {
	JvmId       uint64
	Name        string
	signature   string
	isStatic    bool
	staticValue FieldValue
	value       FieldValue
	ignore      bool
}
/*
Usage:
JavaFieldDef("st").
   Sig("Ljava/lang/String;").
   Static("hello").
   Ignore()
*/
func JavaFieldDef(name string) *javaField {
	return &javaField{
		JvmId: NextFieldId(),
		Name: name,
	}
}
func (jf *javaField) Sig(sig string) *javaField {
	jf.signature = sig
	return jf
}
func (jf *javaField) Ignore() *javaField {
	jf.ignore = true
	return jf
}
func (jf *javaField) Value(value FieldValue) *javaField {
	jf.value = value
	return jf
}
func (jf *javaField) Static(staticValue FieldValue) *javaField {
	if staticValue == nil {
		return nil
	}
	jf.isStatic = true
	jf.staticValue = staticValue
	return jf
}