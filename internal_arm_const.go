package emulator

// From http://infocenter.arm.com/help/topic/com.arm.doc.ihi0044f/IHI0044F_aaelf.pdf
var (
	R_ARM_ABS32 uint32     = 2
	R_ARM_GLOB_DAT uint32  = 21
	R_ARM_JUMP_SLOT uint32 = 22
	R_ARM_RELATIVE uint32  = 23
	//64
	R_AARCH64_GLOB_DAT uint32  = 1025
	R_AARCH64_JUMP_SLOT uint32 = 1026
	R_AARCH64_RELATIVE uint32  = 1027
)
