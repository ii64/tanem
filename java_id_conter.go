package emulator

var (
	curJvmId        uint64 = 1
	curJvmMethodId  uint64 = 1
	curJvmFieldId   uint64 = 1
)

func NextClassId() uint64 {
	id := curJvmId
	curJvmId = curJvmId + 1
	return id
}
func NextMethodId() uint64 {
	id := curJvmMethodId
	curJvmMethodId = curJvmMethodId + 1
	return id
}
func NextFieldId() uint64 {
	id := curJvmFieldId
	curJvmFieldId = curJvmFieldId + 1
	return id
}