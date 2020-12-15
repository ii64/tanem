package emulator

var (
	JNI_FALSE uint64 = 0
	JNI_TRUE  uint64 = 0

	JNI_VERSION_1_1 uint64 = 0x00010001
	JNI_VERSION_1_2 uint64 = 0x00010002
	JNI_VERSION_1_4 uint64 = 0x00010004
	JNI_VERSION_1_6 uint64 = 0x00010006

	JNI_OK         uint64 = 0  // no error
	JNI_ERR        uint64 = __negVal(-1)  // generic error
	JNI_EDETACHED  uint64 = __negVal(-2)  // thread detached from the VM
	JNI_EVERSION   uint64 = __negVal(-3)  // JNI version error
	JNI_ENOMEM     uint64 = __negVal(-4)  // Out of memory
	JNI_EEXIST     uint64 = __negVal(-5)  // VM already created
	JNI_EINVAL     uint64 = __negVal(-6)  // Invalid argument

	JNI_COMMIT     uint64 = 1  // copy content, do not free buffer
	JNI_ABORT      uint64 = 2  // free buffer w/o copying back

)
func __negVal(x int64) uint64 {
	return uint64(x)
}