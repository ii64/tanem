package emulator

import (
	"os"
	"syscall"
)

var (
	GPcb *Pcb
)

func GetPcb() *Pcb {
	if GPcb == nil {
		GPcb = NewPcb()
	}
	return GPcb
}

type Pcb struct {
	fds map[uintptr]*VirtualFile
	pid int
}
func NewPcb() *Pcb {
	return &Pcb{
		fds: map[uintptr]*VirtualFile{
			uintptr(syscall.Stdin):   NewVirtualFile("stdin",  "", os.Stdin),
			uintptr(syscall.Stdout):  NewVirtualFile("stdout", "", os.Stdout),
			uintptr(syscall.Stderr):  NewVirtualFile("stderr", "", os.Stderr),
		},
		pid: os.Getpid(),
	}
}
func (p *Pcb) GetPid() int {
	return p.pid
}
func (p *Pcb) AddFd(name, nameInSystem string, fo *os.File) uintptr {
	var x = NewVirtualFile(name, nameInSystem, fo)
	p.fds[x.Description] = x
	return x.Description
}
func (p *Pcb) GetFdDetail(fd uintptr) *VirtualFile {
	ob, exist := p.fds[fd]
	if !exist {
		return nil
	}
	return ob
}
func (p *Pcb) HasFd(fd uintptr) bool {
	_, exist := p.fds[fd]
	return exist
}
func (p *Pcb) Remove(fd uintptr) {
	if p.HasFd(fd) {
		p.fds[fd].CloseResource()
		delete(p.fds, fd)
	}
}








