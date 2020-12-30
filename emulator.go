package emulator

import (
	"io"
	"os"
	"fmt"
	"math/rand"
	"github.com/mattn/go-colorable"
	zl  "github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)
const maxInt64 uint64 = 1 << 63 - 1
func randomHelper(n uint64) uint64 {
	if n < maxInt64 {
		return uint64(rand.Int63n(int64(n+1)))
	}
	x := rand.Uint64()
	for x > n {
		x = rand.Uint64()
	}
	return x
}
func randUint64(min, max uint64) uint64 {
	return randomHelper(max - min) + min
}

const (
	JsonLog int = iota
	ConsoleLog int = iota
)

type Options struct {
	VfsRoot     string
	ConfigPath  string
	VfpInstSet  bool
	LogColor    bool
	LogAs     	int
	Config      *Config
}
func NewDefaultOptions() *Options {
	return &Options{
		VfsRoot: "vfs",
		ConfigPath: "default.json",
		VfpInstSet: true,
		LogColor: false,
		LogAs: JsonLog,
	}
}

func init() {
	zl.TimeFieldFormat = zl.TimeFormatUnix
	zl.TimestampFieldName = "t"
	zl.LevelFieldName = "l"
	zl.MessageFieldName = "m"
}

type Emulator struct {
	config           *Config
	vfsRoot          string
	configPath       string
	vfpInstSet       bool

	system_prop      map[string]string

	Memory           *MemoryMap
	interruptHandler *InterruptHandler
	syscallHandlers  *SyscallHandlers
	syscallHooks     *SyscallHooks
	Vfs              *VirtualFileSystem
	Hooker           *Hooker
	JavaClassLoader  *JavaClassLoader
	JavaVM           *JavaVM
	Modules          *Modules
	NativeMemory     *NativeMemory
	NativeHooks      *NativeHooks

	logger           zl.Logger
	Pcb              *Pcb
	Mu               uc.Unicorn
}
func NewEmulator(opt *Options) (*Emulator, error) {
	if opt == nil {
		opt = NewDefaultOptions()
	}
	mu, err := uc.NewUnicorn(uc.ARCH_ARM, uc.MODE_ARM)
	if err != nil {
		return nil, err
	}
	emu := &Emulator{
		config: NewDefaultConfig(),
		Mu: mu,
	}
	if opt.Config != nil {
		err = LoadOrCreateConfig(opt.ConfigPath, emu.config)
		if err != nil {
			return nil, err		
		}
	}
	emu.configPath   = opt.ConfigPath
	emu.vfsRoot      = opt.VfsRoot
	emu.vfpInstSet   = opt.VfpInstSet
	var (
		writer io.Writer
		usingColorLog bool
	)
	if opt.LogColor {
		writer = colorable.NewColorableStdout()
		usingColorLog = true
	}else{
		writer = os.Stdout
		usingColorLog = false
	}
	if opt.LogAs == ConsoleLog {
		emu.logger = log.Output(zl.ConsoleWriter{
			Out: writer,
			NoColor: !usingColorLog,
		})
	}else{
		emu.logger = zl.New(os.Stdout).With().
			Timestamp().
			Str("cf", emu.configPath).
			Logger()
	}
	log.Logger = emu.logger
	//
	emu.logger.Info().
		Str("cf", emu.configPath).
		Int("pid", emu.config.Pid).
		Int("uid", emu.config.Uid).
		Str("ip", emu.config.Ip).
		Str("pkgname", emu.config.PkgName).
		Str("androidid", emu.config.AndroidID).
		Str("vfs", emu.vfsRoot).
		Bool("vfp_inst_set", emu.vfpInstSet).
		Msgf("init emu, mac:%X", emu.config.Mac)
	//
	if emu.vfpInstSet {
		emu.logger.Debug().Msg("vfp is enabled")
		emu.enableVfp()
		emu.logger.Debug().Msg("vfp finish")
	}

	emu.Pcb = GetPcb()
	emu.logger.Info().
		//program的pid
		Int("pid", emu.Pcb.GetPid()).
		Msg("pcb")

	//注意，原有缺陷，libc_preinit init array中访问R1参数是从内核传过来的
	//而这里直接将0映射空间，,强行运行过去，因为R1刚好为0,否则会报memory unmap异常
	//FIXME:MRC指令总是返回0,TLS模擬
	//TODO 初始化libc时候R1参数模拟内核传过去的KernelArgumentBlock	
	emu.logger.Debug().
		Err(emu.Mu.MemMapProt(0x0, 0x00001000, uc.PROT_READ | uc.PROT_WRITE)).
		Msg("map memory libc_preinit")

	//Android
	emu.system_prop = map[string]string{
		"libc.debug.malloc.options": "", "ro.build.version.sdk":"19", "ro.build.version.release":"4.4.4","persist.sys.dalvik.vm.lib":"libdvm.so", "ro.product.cpu.abi":"armeabi-v7a", "ro.product.cpu.abi2":"armeabi", 
		"ro.product.manufacturer":"LGE", "ro.debuggable":"0", "ro.product.model":"AOSP on HammerHead","ro.hardware":"hammerhead", "ro.product.board":"hammerhead", "ro.product.device":"hammerhead", 
		"ro.build.host":"833d1eed3ea3", "ro.build.type":"user", 
		"ro.secure":"1", "wifi.interface":"wlan0", "ro.product.brand":"Android",
	}
	emu.Memory = NewMemoryMap(emu.Mu,
		MAP_ALLOC_BASE,
		MAP_ALLOC_BASE+MAP_ALLOC_SIZE,
	)
	//Stack
	emu.logger.Debug().
		Str("begin", ConvHex("%08X",STACK_ADDR)).
		Str("size", ConvHex("%08X",STACK_SIZE)).
		Msg("mapping stack memory")
	addr, err := emu.Memory.Map(
		STACK_ADDR,
		STACK_SIZE,
		uc.PROT_READ | uc.PROT_WRITE,
		nil, 0,
	)
	_ = addr
	if err != nil {
		emu.logger.Debug().Err(err).Msg("failed to map stack memory")
		return nil, err
	}
	emu.Mu.RegWrite(uc.ARM_REG_SP, STACK_ADDR + STACK_SIZE)
	sp, err := emu.Mu.RegRead(uc.ARM_REG_SP)
	if err != nil {
		return nil, err
	}
	emu.logger.Info().Str("stack", ConvHex("%08X", sp)).Msg("stack address")

	//CPU
	emu.logger.Debug().Msg("init syscall handler")
	emu.interruptHandler = NewInterruptHandler(emu.Mu)
	emu.interruptHandler.SetLogger(emu.logger)
	emu.syscallHandlers = NewSyscallHandlers(emu.interruptHandler)
	emu.syscallHandlers.SetLogger(emu.logger)
	emu.syscallHooks = NewSyscallHooks(emu.Mu, emu.syscallHandlers)
	emu.syscallHooks.SetLogger(emu.logger)

	// File System
	emu.logger.Debug().Msg("init vfs")
	emu.Vfs = NewVirtualFileSystem(
		emu.vfsRoot,
		emu.syscallHandlers,
		emu.Memory,
		emu.logger,
		emu.config,
	)

	// Hooker
	emu.logger.Debug().Msg("init hooker")
	emu.Memory.Map(
		HOOK_MEMORY_BASE, HOOK_MEMORY_SIZE,
		uc.PROT_READ | uc.PROT_WRITE | uc.PROT_EXEC,
		nil, 0,
	)
	emu.Hooker = NewHooker(
		emu,
		emu.logger,
		HOOK_MEMORY_BASE, HOOK_MEMORY_SIZE,
	)

	// JavaVM
	emu.logger.Debug().Msg("init jclassloader, jvm")
	emu.JavaClassLoader = NewJavaClassLoader()
	emu.JavaVM = NewJavaVM(
		emu,
		emu.JavaClassLoader,
		emu.Hooker,
		emu.logger)

	// Executable data.
	emu.logger.Debug().Msg("init modules")
	emu.Modules = NewModules(emu, emu.vfsRoot, emu.logger)
	// Native
	emu.logger.Debug().Msg("init native memory")
	emu.NativeMemory = NewNativeMemory(
		emu.Mu, emu.Memory,
		emu.syscallHandlers,
		emu.Vfs,
		emu.logger,
	)
	emu.logger.Debug().Msg("init native hooks")
	emu.NativeHooks  = NewNativeHooks(
		emu, emu.NativeMemory, 
		emu.Modules, emu.Hooker,
		emu.Vfs,
		emu.logger,
	)

	emu.logger.Debug().Msg("register java class")
	emu.addClasses()

	//映射常用的文件，cpu一些原子操作的函数实现地方	
	path := fmt.Sprintf("%s/system/lib/vectors", emu.vfsRoot)
	emu.logger.Debug().Msgf("loading to memory: %s", path)
	vfo, err := MyOpen(path, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	vf := NewVirtualFile("[vectors]", path, vfo)
	_, err = emu.Memory.Map(0xffff0000, 0x1000, uc.PROT_EXEC | uc.PROT_READ, vf, 0)
	if err != nil {
		return nil, err
	}
	
	//映射app_process，android系统基本特征
	path = fmt.Sprintf("%s/system/bin/app_process32", emu.vfsRoot)
	emu.logger.Debug().Msgf("loading to memory: %s", path)
	inf, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	sz := uint64(inf.Size())
	vfo, err = MyOpen(path, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	vf = NewVirtualFile("/system/bin/app_process32", path, vfo)
	_, err = emu.Memory.Map(0xab006000, sz, uc.PROT_WRITE | uc.PROT_READ, vf, 0)
	if err != nil {
		return nil, err
	}

	return emu, nil
}
//
func (emu *Emulator) LoadLibrary(filename string, doInit bool) (*Module, error) {
	return emu.Modules.LoadModule(filename, doInit)
}
func (emu *Emulator) CallSymbol(module *Module, symbolName string, args ...interface{}) (uint64, error) {
	symbolAddress, exist := module.FindSymbol(symbolName)
	if !exist {
		return 0, ErrELFSymbolNotFound
	}
	return emu.CallNative(symbolAddress, args...)
}
func (emu *Emulator) CallNative(address uint32, args ...interface{}) (uint64, error) {
	// Detect JNI call
	isJNI := false
	if len(args) >= 1 {
		if v, ok := args[0].(uint64); ok {
			if v == emu.JavaVM.addressPtr || v == emu.JavaVM.JniEnv.addressPtr {
				isJNI = true
			}
		}
	}
	err := NativeWriteArgs(emu, args...)
	if err != nil {
		return 0, err
	}
	stopPos := randUint64(
		HOOK_MEMORY_BASE,
		HOOK_MEMORY_BASE+HOOK_MEMORY_SIZE,
	) | 1
	err = emu.Mu.RegWrite(uc.ARM_REG_LR, stopPos)
	if err != nil {
		return 0, err
	}
	err = emu.Mu.Start(uint64(address), stopPos - 1)
	if err != nil {
		return 0, err
	}
	// Read result from locals if jni
	defer func(){
		if isJNI {
			emu.JavaVM.JniEnv.ClearLocals()
		}
	}()
	ret, err := emu.Mu.RegRead(uc.ARM_REG_R0)
	if err != nil {
		return 0, err
	}
	if isJNI {
		res := emu.JavaVM.JniEnv.GetLocalReference(ret)
		return res, nil //JAVA_NULL=0
	}
	return ret, nil
}


//
func (emu *Emulator) addClasses() {
	// load base java class
}
//
func (emu *Emulator) enableVfp() {
	// https://github.com/unicorn-engine/unicorn/blob/8c6cbe3f3cabed57b23b721c29f937dd5baafc90/tests/regress/arm_fp_vfp_disabled.py#L15
	// MRC p15, #0, r1, c1, c0, #2
	// ORR r1, r1, #(0xf << 20)
	// MCR p15, #0, r1, c1, c0, #2
	// MOV r1, #0
	// MCR p15, #0, r1, c7, c5, #4
	// MOV r0,#0x40000000
	// FMXR FPEXC, r0
	code := []byte{
		0x11, 0xEE, 0x50, 0x1F,
		0x41, 0xF4, 0x70, 0x01,
		0x01, 0xEE, 0x50, 0x1F,
		0x4F, 0xF0, 0x00, 0x01,
		0x07, 0xEE, 0x95, 0x1F,
		0x4F, 0xF0, 0x80, 0x40,
		0xE8, 0xEE, 0x10, 0x0A,
		// vpush {d8}
		0x2d, 0xed, 0x02, 0x8b,
	}
	var (
		address  uint64 = 0x1000
		mem_size uint64 = 0x1000
	)
	emu.logger.Debug().Err(
		emu.Mu.MemMap(address, mem_size),
	).Uint64("addr", address).Uint64("sz", mem_size).
	Msg("vfp mem_map")
	emu.logger.Debug().Err(
		emu.Mu.MemWrite(address, code),
	).Uint64("addr", address).Uint64("sz", mem_size).
	Msg("vfp mem_write")
	emu.logger.Debug().Err(
		emu.Mu.RegWrite(uc.ARM_REG_SP, address + mem_size),
	).Uint64("addr", address).Uint64("sz", mem_size).
	Msg("vfp reg_write")
	emu.logger.Debug().Err(
		emu.Mu.Start(address | 1, address + uint64(len(code))),
	).Uint64("addr", address).Uint64("sz", mem_size).
	Msg("vfp emu_start")
	emu.logger.Debug().Err(
		emu.Mu.MemUnmap(address, mem_size),
	).Uint64("addr", address).Uint64("sz", mem_size).
	Msg("vfp mem_unmap")
}







