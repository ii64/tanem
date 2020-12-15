package emulator

import (
	"os"
	"strings"
	"runtime"
	"path/filepath"
)

var (
	isWin = false

	PF_X uint32 = 0x1
	PF_W uint32 = 0x2
	PF_R uint32 = 0x4
)

func init() {
	isWin = runtime.GOOS == "windows"
}

func VfsPathToSystemPath(vfs_root, path string) string {
	if isWin {
		path = strings.Replace(path, ":", "_", -1)
	}
	if filepath.IsAbs(path) {
		return vfs_root + path
	}
	return vfs_root + "/system/lib/" + path
}

func SystemPathToVfsPath(vfs_root, path string) (string, error) {
	pat, err := filepath.Rel(vfs_root, path)
	if err != nil {
		return "", err
	}
	return "/" + pat, nil
}


func PageStart(addr uint64) uint64 {
	return addr & -(PAGE_SIZE)
}

func PageEnd(addr uint64) uint64 {
	return PageStart(addr) + PAGE_SIZE
}

func GetSegmentProtection(prot_in uint32) int {
	prot := 0
	if (prot_in & PF_R) != 0 {
		prot |= 1
	}
	if (prot_in & PF_W) != 0 {
		prot |= 2
	}
	if (prot_in & PF_X) != 0 {
		prot |= 4
	}
	return prot
}

/*
def my_open(fd, flag):
    global g_isWin
    if(g_isWin):
        flag = flag | os.O_BINARY
    #
    return os.open(fd, flag)
*/
func MyOpen(fd string, flag int) (*os.File, error) {
	return os.OpenFile(fd, flag, 0755)
}



