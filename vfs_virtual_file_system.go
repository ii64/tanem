package emulator

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"math/rand"
	fp  "path/filepath"
	zl  "github.com/rs/zerolog"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

var (
	OVERRIDE_URANDOM = false
	OVERRIDE_URANDOM_INT = 1
	S_STATUS = `
Name:   {pkg_name}
State:  R (running)
Tgid:   1434
Pid:    1434
PPid:   197
TracerPid:      0
Uid:    10054   10054   10054   10054
Gid:    10054   10054   10054   10054
FDSize: 512
Groups: 1015 1028 3003 50054 
VmPeak:  1229168 kB
VmSize:  1115232 kB
VmLck:         0 kB
VmPin:         0 kB
VmHWM:    179992 kB
VmRSS:    179836 kB
VmData:   191904 kB
VmStk:       136 kB
VmExe:         8 kB
VmLib:     48448 kB
VmPTE:       536 kB
VmSwap:        0 kB
Threads:        105
SigQ:   0/12272
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000001204
SigIgn: 0000000000000000
SigCgt: 00000002000094f8
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: fffffff000000000
Cpus_allowed:   f
Cpus_allowed_list:      0-3
voluntary_ctxt_switches:        5225
nonvoluntary_ctxt_switches:     11520
`
)

type VirtualFileSystem struct {
	vfsRoot  string
	sh       *SyscallHandlers
	pcb      *Pcb
	mem      *MemoryMap
	logger   zl.Logger
	config   *Config
}
func NewVirtualFileSystem(root string,sh *SyscallHandlers,mem *MemoryMap,logger zl.Logger, config *Config) *VirtualFileSystem {
	vfs := &VirtualFileSystem{
		vfsRoot: root,
		sh: sh,
		pcb: GetPcb(),
		mem: mem,
		logger: logger,
		config: config,
	}
	vfs.ClearProcDir()
	vfs.sh.SetHandler(0x3, "read", 3, vfs.readHandle)
	vfs.sh.SetHandler(0x4, "write", 3, vfs.writeHandle)
	vfs.sh.SetHandler(0x5, "open", 3, vfs.openHandle)
	vfs.sh.SetHandler(0x6, "close", 1, vfs.closeHandle)
	vfs.sh.SetHandler(0x0A, "unlink", 1, vfs.unlinkHandle)
	vfs.sh.SetHandler(0x13, "lseek", 3, vfs.lseekHandle)
	vfs.sh.SetHandler(0x21, "access", 2, vfs.accessHandle)
	vfs.sh.SetHandler(0x27, "mkdir", 2, vfs.mkdirHandle)
	vfs.sh.SetHandler(0x36, "ioctl", 6, vfs.ioctlHandle)
	vfs.sh.SetHandler(0x37, "fcntl", 6, vfs.fcntl64Handle)
	vfs.sh.SetHandler(0x92, "writev", 3, vfs.writevHandle)
	vfs.sh.SetHandler(0xC3, "stat64", 2, vfs.stat64Handle)
	vfs.sh.SetHandler(0xC4, "lstat64", 2, vfs.lstat64Handle)
	vfs.sh.SetHandler(0xC5, "fstat64", 2, vfs.fstat64Handle)
	vfs.sh.SetHandler(0xD9, "getdents64", 3, vfs.getdents64Handle)
	vfs.sh.SetHandler(0xDD, "fcntl64", 6, vfs.fcntl64Handle)
	vfs.sh.SetHandler(0x10A, "statfs64", 3, vfs.statfs64Handle)
	vfs.sh.SetHandler(0x142, "openat", 4, vfs.openatHandle)
	vfs.sh.SetHandler(0x147, "fstatat64", 4, vfs.fstatat64Handle)
	vfs.sh.SetHandler(0x14c, "readlinkat", 4, vfs.readlinkatHandle)
	vfs.sh.SetHandler(0x14e, "faccessat", 4, vfs.faccessatHandle)
	return vfs
}
//with caution of escape-access
func (vfs *VirtualFileSystem) TranslatePath(filename string) string {
	return VfsPathToSystemPath(vfs.vfsRoot, filename)
}
//WIP, add failback
func (vfs *VirtualFileSystem) ClearProcDir() error {
	proc := "/proc"
	proc = vfs.TranslatePath(proc)
	items, err := ioutil.ReadDir(proc)
	if err != nil {
		return err
	}
	numericPrefix := []string{"0","1","2","3","4","5","6","7","8","9"}
	for _, item := range items {
		if item.IsDir() {
			hasNumericPrefix := false
			for _,np := range numericPrefix {
				if strings.HasPrefix(item.Name(), np) {
					hasNumericPrefix = true
					break
				}
			}
			if hasNumericPrefix {
				os.RemoveAll(proc+"/"+item.Name())
			}
		}
	}
	return nil
}
// syscall read
func (vfs *VirtualFileSystem) readHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	vfs.logger.Debug().Msg("read called")
	return 0, true
}
// syscall write
func (vfs *VirtualFileSystem) writeHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall open
func (vfs *VirtualFileSystem) openHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall close
func (vfs *VirtualFileSystem) closeHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall unlink
func (vfs *VirtualFileSystem) unlinkHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall lseek
func (vfs *VirtualFileSystem) lseekHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall access
func (vfs *VirtualFileSystem) accessHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall mkdir
func (vfs *VirtualFileSystem) mkdirHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall ioctl
func (vfs *VirtualFileSystem) ioctlHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall fcntl
func (vfs *VirtualFileSystem) fcntlHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall writev
func (vfs *VirtualFileSystem) writevHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall stat64
func (vfs *VirtualFileSystem) stat64Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall lstat64
func (vfs *VirtualFileSystem) lstat64Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall fstat64
func (vfs *VirtualFileSystem) fstat64Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall getdents64
func (vfs *VirtualFileSystem) getdents64Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall fcntl64
func (vfs *VirtualFileSystem) fcntl64Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall statfs64
func (vfs *VirtualFileSystem) statfs64Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall openat
func (vfs *VirtualFileSystem) openatHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall fstatat64
func (vfs *VirtualFileSystem) fstatat64Handle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall readlinkat
func (vfs *VirtualFileSystem) readlinkatHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}
// syscall faccessaat
func (vfs *VirtualFileSystem) faccessatHandle(mu uc.Unicorn, args ...uint64) (uint64, bool) {
	return 0, true
}

func (vfs *VirtualFileSystem) createFdLink(fd uintptr, target string) {
	if isWin {
		return
	}
	if fd >= 0 {
		pid := vfs.pcb.GetPid()
		fdbase := fmt.Sprintf("/proc/%d/fd/", pid)
		fdbase = vfs.TranslatePath(fdbase)
		_, err := os.Stat(fdbase)
		if err == nil {
			err = os.Remove(fdbase)
			if err != nil {
				vfs.logger.Debug().Err(err).Msg("failed to delete fd symlink")
			}
		}
		p := fmt.Sprintf("%s/%d", fdbase, fd)
		err = os.Symlink(target, p)
		if err != nil {
			vfs.logger.Debug().Str("from",target).Str("to", p).Err(err).Msg("failed to create symlink fd")
		}
	}
	return
}
func (vfs *VirtualFileSystem) delFdLink(fd uintptr) {
	if isWin {
		return
	}
	if fd >= 0 {
		pid := vfs.pcb.GetPid()
		fdbase := fmt.Sprintf("/proc/%d/fd/", pid)
		fdbase = vfs.TranslatePath(fdbase)
		p := fmt.Sprintf("%s/%d", fdbase, fd)
		if _, err := os.Stat(fdbase); err == nil {
			err = os.Remove(p)
			if err != nil {
				vfs.logger.Debug().Err(err).Msg("failed to delete fd")
			}
		}
	}
}
func (vfs *VirtualFileSystem) openFile(filename string, mode int) *os.File {
	var (
		f *os.File
		err error
	)
	filepath := vfs.TranslatePath(filename)
	vfs.logger.Debug().Str("path", filepath).Msg("open file")
	if filename == "/dev/urandom" {
		parent := fp.Dir(filepath)
		os.MkdirAll(parent, os.ModePerm)
		f, err = os.Create(filepath)
		if err != nil {
			vfs.logger.Debug().Err(err).Msg("failed to open file")
			return nil
		}
		ran := make([]byte, 128)
		rand.Read(ran)
		f.Write(ran)
	}else if strings.HasPrefix(filename, "/proc") {
		parent := fp.Dir(filepath)
		os.MkdirAll(parent, os.ModePerm)
		pobj := GetPcb()
		pid := pobj.GetPid()
		filename2 := strings.Replace(filename, fmt.Sprintf("%s", pid), "self", -1)
		mapPath := "/proc/self/maps"
		if filename2 == mapPath {
			f, err = os.Create(filepath)
			if err != nil {
				vfs.logger.Debug().Err(err).Msg("failed to open file")
				return nil
			}
			err = vfs.mem.DumpMaps(f)
			if err != nil {
				vfs.logger.Debug().Err(err).Msg("failed to dump maps memory")
				return nil
			}
		}
		cmdlinePath := "/proc/self/cmdline"
		if filename2 == cmdlinePath {
			f, err = os.Create(filepath)
			if err != nil {
				vfs.logger.Debug().Err(err).Msg("failed to open file")
				return nil
			}
			f.Write([]byte(vfs.config.PkgName))
		}
		cgroupPath := "/proc/self/cgroup"
		if filename2 == cgroupPath {
			f, err = os.Create(filepath)
			if err != nil {
				vfs.logger.Debug().Err(err).Msg("failed to open file")
				return nil
			}
			content := fmt.Sprintf("2:cpu:/apps\n1:cpuacct:/uid/%d\n",vfs.config.Uid)
			f.Write([]byte(content))
		}
		statusPath := "/proc/self/status"
		if filename2 == statusPath {
			f, err = os.Create(filepath)
			if err != nil {
				vfs.logger.Debug().Err(err).Msg("failed to open file")
				return nil
			}
			statsx := []byte(strings.Replace(S_STATUS, "{pkg_name}", vfs.config.PkgName, -1))
			f.Write(statsx)
		}

		return f
	}
	//
	virtualFile := []string{
		"/dev/log/main",
		"/dev/log/events",
		"/dev/log/radio",
		"/dev/log/system",
		"/dev/input/event0",
	}
	for _, _ = range virtualFile {
		parent := fp.Dir(filepath)
		os.MkdirAll(parent, os.ModePerm)
		f, err = os.Create(filepath)
		if err != nil {
			continue
		}
		f.Close()
	}
	wi, err := os.Stat(filepath)
	if err != nil {
		vfs.logger.Debug().Err(err).Msg("failed to see stat, it may does not exist")
		return nil
	}
	if !wi.IsDir() {
		flags := os.O_RDWR
		if (mode & 100) != 0 {
			flags |= os.O_CREATE
		}
		if (mode & 2000) != 0 {
			flags |= os.O_APPEND
		}
		fo, err := MyOpen(filepath, flags)
		if err != nil {
			vfs.logger.Debug().Err(err).Msgf("failed to open file, flags %d", flags)
			return nil
		}
		fdx := vfs.pcb.AddFd(filename, filepath, fo)
		vfs.createFdLink(fdx, filepath)
		return fo
	}
	return nil
}


















