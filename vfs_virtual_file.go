package emulator

import (
	"os"
)

type VirtualFile struct {
	Name         string
	NameInSystem string
	Description  uintptr
	fo           *os.File
}

func NewVirtualFile(name, nameInSystem string, fo *os.File) *VirtualFile {
	return &VirtualFile{
		Name: name, 
		NameInSystem: nameInSystem,
		Description: fo.Fd(),
		fo: fo,
	}
}
func (vf *VirtualFile) CloseResource() {
	vf.fo.Close()
}