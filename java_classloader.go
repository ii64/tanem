package emulator

type JavaClassLoader struct {
	ClassById   map[uint64]*javaClass
	ClassByName map[string]*javaClass
}
func NewJavaClassLoader() *JavaClassLoader {
	jcl := &JavaClassLoader{
		ClassById: map[uint64]*javaClass{},
		ClassByName: map[string]*javaClass{},
	}
	return jcl
}
func (jc *JavaClassLoader) AddClass(cls *javaClass, force bool) error {
	if _, exist := jc.ClassById[cls.JvmId]; exist && !force {
		return ErrJavaClassLoaded
	}
	if _, exist := jc.ClassByName[cls.JvmName]; exist && !force {
		return ErrJavaClassLoaded
	}
	cls.Class = NewClass(cls)
	jc.ClassById[cls.JvmId] = cls
	jc.ClassByName[cls.JvmName] = cls
	return nil
}