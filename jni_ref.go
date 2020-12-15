package emulator

type jobject struct {
	val interface{}
}

func NewJObject(val interface{}) *jobject {
	return &jobject{
		val:val,
	}
}
func (jx *jobject) setVal(val interface{}) {
	jx.val = val
}
func (jx *jobject) Value() interface{} {
	return jx.val
}

type jclass struct {
	jobject
}
func NewJClass(val interface{}) *jclass {
	jx := &jclass{}
	jx.setVal(val)
	return jx
}