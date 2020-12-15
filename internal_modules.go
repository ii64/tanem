package emulator

import (
	"os"
	"github.com/pkg/errors"
	bin "encoding/binary"
	zl  "github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

type Module struct {
	address      uint64
	size         uint64
	filename     string
	symbols      map[string]uint32
	initArray    []uint32
	symbolLookup map[uint32]string
	soinfoPtr    uint32
}
func NewModule(
	filename string,
	base,
	size uint64,
	symRes map[string]uint32,
	initArray []uint32,
	soinfoPtr uint32) *Module {
	m := &Module{
		filename: filename,
		address: base,
		size: size,
		symbols: symRes,
		symbolLookup: map[uint32]string{},
		initArray: initArray,
		soinfoPtr: soinfoPtr,
	}
	for symName,symAddr := range m.symbols {
		m.symbolLookup[symAddr] = symName
	}
	return m
}
func (m *Module) Name() string {
	return m.filename
}
func (m *Module) FindSymbol(symbolStr string) (uint32, bool) {
	addr, exist := m.symbols[symbolStr]
	return addr, exist
}
func (m *Module) IsSymbolAddr(addr uint32) (string, bool) {
	if name, exist := m.symbolLookup[addr]; exist {
		return name, true
	}else if name, exist = m.symbolLookup[addr+1]; exist {
		return name, true
	}
	return "", false
}
func (m *Module) CallInit(emu *Emulator) {
	for _, funPtr := range m.initArray {
		funAddr := funPtr
		ret, err := emu.CallNative(funAddr)
		log.Debug().Uint64("ret", ret).Err(err).Msgf("Calling Init_array %s function: 0x%08X!!", m.filename, funAddr)
	}
}



///
type Modules struct {
	emu            *Emulator
	vfsRoot        string
	modules        []*Module
	counterMemory  uint64
	symbolHooks    map[string]uint64
	soinfoAreaBase uint64
	logger         zl.Logger
}

func NewModules(emu *Emulator, vfsRoot string, logger zl.Logger) *Modules {
	ms := &Modules{
		emu: emu,
		vfsRoot: vfsRoot,
		modules: []*Module{},
		counterMemory: BASE_ADDR,
		symbolHooks: map[string]uint64{},
		logger: logger,
	}
	var soinfoAreaSz uint64 = 0x40000
	addr, err := ms.emu.Memory.Map(0, soinfoAreaSz, uc.PROT_WRITE | uc.PROT_READ,
		nil, 0)
	if err != nil {
		return nil
	}
	ms.soinfoAreaBase = addr
	return ms
}
func (ms *Modules) GetModules() []*Module {
	return ms.modules
}
func (ms *Modules) AddSymbolHook(symName string, addr uint64) {
	ms.symbolHooks[symName] = addr
}
func (ms *Modules) FindSymbol(addr uint32) (string, bool) {
	for _,module := range ms.modules {
		if sl, exist := module.symbolLookup[addr]; exist {
			return sl, true
		}
	}
	return "", false
}
func (ms *Modules) FindSymbolStr(symbolStr string) (uint64, bool) {
	for _,module := range ms.modules {
		sym, found := module.FindSymbol(symbolStr)
		if found {
			return uint64(sym), true
		}
	}
	return 0, false
}
//
func (ms *Modules) FindModule(addr uint64) *Module {
	for _, module := range ms.modules {
		if module.address == addr {
			return module
		}
	}
	return nil
}
func (ms *Modules) FindModuleByName(filename string) *Module {
	abs := filename
	for _, module := range ms.modules {
		absm := module.filename
		if abs == absm {
			return module
		}
	}
	return nil
}
//
func (ms *Modules) MemReserve(start, end uint64) uint64{
//	ms.logger.Debug().Msgf("memreserve: start:%08X end:%08X", start, end)
	sz_aligned := PageEnd(end) - PageStart(start)
	ret := ms.counterMemory
	ms.counterMemory = ms.counterMemory + sz_aligned
	return ret
}
//
func (ms *Modules) LoadModule(filename string, doInit bool) (*Module, error) {
	m := ms.FindModuleByName(filename)
	if m != nil {
		return m, nil
	}
	ms.logger.Debug().Str("path", filename).Msg("loadmodule")
	reader, err := NewELFReader(filename)
	if err != nil {
		return nil, err
	}
	// Parse program header (Execution view).

	// - LOAD (determinate what parts of the ELF file get mapped into memory)
	loadSegments := reader.GetLoad()
//	ms.logger.Debug().Msgf("load Segments %+#v", loadSegments)
	// Find bounds of the load segments
	var (
		boundLow  uint32 = 0
		boundHigh uint32 = 0
	)
	for _, seg := range loadSegments {
		pMemsz := seg.pMemsz
		if pMemsz == 0 {
			continue
		}
		pVaddr := seg.pVaddr
		if boundLow > pVaddr {
			boundLow = pVaddr
		}
		high := pVaddr + pMemsz
		if boundHigh < high {
			boundHigh = high
		}
	}
//	ms.logger.Debug().Msgf("boundLow: %08X boundHigh: %08X", boundLow, boundHigh)
	relpt, err := SystemPathToVfsPath(ms.vfsRoot, filename)
	if err != nil {
		ms.logger.Debug().Err(err).Str("filename", filename).Str("rootfs", ms.vfsRoot).Msg("convert system path to vfs path failed")
		return nil, err
	}
	fo, err := MyOpen(filename, os.O_RDONLY)
	if err != nil {
		ms.logger.Debug().Err(err).Str("filename", filename).Str("rootfs", ms.vfsRoot).Msg("failed to open file")
		return nil, err
	}
	loadBase := ms.MemReserve(uint64(boundLow), uint64(boundHigh))
	vf := NewVirtualFile(
		relpt, filename, fo)
//	var lastSegSz uint64
	for _, seg := range loadSegments {
		pFlags := seg.pFlags
		prot := GetSegmentProtection(pFlags)
		if prot == 0 {
			prot = uc.PROT_ALL
		}
		pVaddr := uint64(seg.pVaddr)
		segStart := loadBase + pVaddr
		segPageStart := PageStart(segStart)
		fileStart := uint64(seg.pOffset)
		fileEnd := fileStart + uint64(seg.pFilesz)
		filePageStart := PageStart(fileStart)
		fileLength := fileEnd - filePageStart
		if !(fileLength>0) {
			ms.logger.Debug().Msg("loaded file is empty")
			//return nil
		}
		if fileLength > 0 {
			_, err := ms.emu.Memory.Map(segPageStart, fileLength, prot, vf, filePageStart)
			if err != nil {
				return nil, errors.Wrap(err, "cannot map memory seg page virtual file")
			}
		}
		segEnd := segStart + uint64(seg.pMemsz)
		segPageEnd := PageEnd(segEnd)
		segFileEnd := PageEnd(segStart+uint64(seg.pFilesz))
//		lastSegSz = segPageEnd
		_, err = ms.emu.Memory.Map(segFileEnd, segPageEnd-segFileEnd, prot, nil, 0)
		if err != nil {
			return nil, errors.Wrap(err, "cannot map memory seg page")
		}	
	}

	initArrayOffset, initArraySize := reader.GetInitArray()
	initArray := []uint32{}
	initOffset := reader.GetInit()

//	log.Debug().Msgf(">>> init arrOffs:%08X arrOffsz:%08X offst:%08X", initArrayOffset, initArraySize, initOffset)

	soNeeded := reader.GetSoNeeded()
	for _, soName := range soNeeded {
		path := VfsPathToSystemPath(ms.vfsRoot, soName)
		_, err := os.Stat(path)
		if err != nil {
			ms.logger.Debug().Err(err).Str("soPath", path).Msgf("%s is required by %s but not exist", soName, filename)
			continue
		}
		_, err = ms.LoadModule(path, true)
		if err != nil {
			ms.logger.Debug().Str("path", path).Err(err).Msg("required module failed to load")
			continue
		}
	}

	rels := reader.GetRels()
	symbols := reader.GetSymbols()
	symbolsResolved := map[string]uint32{}

	for _, sym := range symbols {
		symAddr, resolved := ms.elfGetSymVal(uint32(loadBase), sym)
		// filter
		if resolved {
			symbolsResolved[sym.Name] = symAddr
		}
	}

	// Relocate.
	relocx := func(relname string, rel relx) (bool, error) { // continue or return
		rInfoSym := rel.RInfoSym
		if int(rInfoSym) >= len(symbols) {
//			ms.logger.Debug().Int("rInfoSym", int(rInfoSym)).Msg("rInfoSym more than symbols size, continue")
			return true, nil
		}
		sym := symbols[int(rInfoSym)]
		symVal := sym.StValue

		relAddr := loadBase + uint64(rel.ROffset)
		relInfoType := rel.RInfoType
// Still debugging		
//		ms.logger.Debug().Msgf("rel base:%08X rOffset:%08X type:%d", loadBase, rel.ROffset, relInfoType)
		symName := reader.GetDynStringByRelSym(int(rInfoSym))
//		if relAddr > lastSegSz {
//			ms.logger.Debug().Msgf("symbol %q address 0x%08X lookup over page end", symName, relAddr)
//			return true, nil
//		}
//		if reader.filename == "vfs/system/lib/libc.so" {
//			log.Debug().Msgf(">>>> %s relInfoType %08X | relAddr %08X == loadBase %08X + roffset %08X", relname, relInfoType, relAddr, loadBase, rel.ROffset)
//		}
		if relInfoType == R_ARM_ABS32 {
			if symAddr, exist := symbolsResolved[symName]; exist {
				valOrigBytes, err := ms.emu.Mu.MemRead(uint64(relAddr), 4)
				if err != nil {
					return false, errors.Wrap(err, "failed to read from rel address R_ARM_ABS32")
				}
//				log.Debug().Msgf(">>>> Rel ABS32 symaddr:%08X symname:%s", symAddr, symName)
				valOrig := bin.LittleEndian.Uint32(valOrigBytes)
				//R_ARM_ABS32 how to relocate see android linker source code
				//*reinterpret_cast<Elf32_Addr*>(reloc) += sym_addr;
				val := symAddr + valOrig
				newVal := make([]byte, 4)
				bin.LittleEndian.PutUint32(newVal, val)
				err = ms.emu.Mu.MemWrite(uint64(relAddr), newVal)
				if err != nil {
					return false, errors.Wrap(err, "unable to rewrite rel address")
				}
			}
		}else if (relInfoType == R_ARM_GLOB_DAT || 
			relInfoType == R_ARM_JUMP_SLOT ||
			relInfoType == R_AARCH64_GLOB_DAT ||
			relInfoType == R_AARCH64_JUMP_SLOT) {
			// Resolve the symbol.
			//R_ARM_GLOB_DAT，R_ARM_JUMP_SLOT how to relocate see android linker source code
			//*reinterpret_cast<Elf32_Addr*>(reloc) = sym_addr;
			if valx, exist := symbolsResolved[symName]; exist {
				valB := make([]byte, 4)
				bin.LittleEndian.PutUint32(valB, valx)
//				log.Debug().Msgf(">>> Rel GLOB/JUMPSLOT symname:%s val:%08X", symName, valx)
				err = ms.emu.Mu.MemWrite(relAddr, valB)
				if err != nil {
					return false, errors.Wrap(err, "unable to write on rel address GLOB_DATA, JUMP_SLOT")
				}
			}else{
				log.Debug().Msgf(">>> not resolved %s",symName)
			}
		}else if (relInfoType == R_ARM_RELATIVE ||
				relInfoType == R_AARCH64_RELATIVE) {
			if symVal == 0 {
				// Load address at which it was linked originally.
				valOrigBytes, err := ms.emu.Mu.MemRead(relAddr, 4)
				if err != nil {
					return false, errors.Wrap(err, "failed to read from rel address RELATIVE")
				}
				valOrig := bin.LittleEndian.Uint32(valOrigBytes)
				// Create the new value
				val := uint32(loadBase) + valOrig
//				log.Debug().Msgf(">>> Rel RELATIVE relAddr:%08X newVal:%08X orig:%08X", relAddr, val, valOrig)
				// Write the new value
				newVal := make([]byte, 4)
				bin.LittleEndian.PutUint32(newVal, val)
				err = ms.emu.Mu.MemWrite(uint64(relAddr), newVal)
				if err != nil {
					return false, errors.Wrap(err, "unable to rewrite rel address")
				}
			}else{
					return false, ErrNotImplemented
			}
		}else{
			ms.logger.Debug().Msgf("unhandled relocation type %d", relInfoType)
		}
		return true, nil
	}
	//**rels
//	if reader.filename == "./bin/libcms_new.so" {
//		ms.logger.Debug().Msgf("librels %+v", rels)
//	}
	//dynrel
	for _, rel := range rels.dynrel {
		if isContinue, err := relocx("dynrel", rel); isContinue {
			continue
		}else{
			return nil, err
		}
	}
	//relplt
	for _, rel := range rels.relplt {
		if isContinue, err := relocx("relplt", rel); isContinue {
			continue
		}else{
			return nil, err
		}
	}
	//
//	log.Debug().Msgf(">>> Add base:%08X Init_offset %08X Init_array %08X", loadBase, initOffset, initArrayOffset)
	if initOffset != 0 {
		initArray = append(initArray, uint32(loadBase)+initOffset)
	}
	for i := 0; i < int(initArraySize)/4; i++ {
		b, err := ms.emu.Mu.MemRead(loadBase+uint64(initArrayOffset), 4)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read init array offset")
		}
		funPtr := bin.LittleEndian.Uint32(b)
		if funPtr != 0 {
			initArray = append(initArray, uint32(funPtr))
		}
//		log.Debug().Msgf(">> A INIT ptr %08X", funPtr)
		initArrayOffset = initArrayOffset + 4
	}
//	log.Debug().Msgf(">> INITARRAY:%v",initArray)

	write_sz, err := reader.WriteSoInfo(ms.emu.Mu, loadBase, ms.soinfoAreaBase)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write so info")
	}
	module := NewModule(filename,
		loadBase,
		uint64(boundHigh-boundLow),
		symbolsResolved,
		initArray,
		uint32(ms.soinfoAreaBase),
	)
	ms.modules = append(ms.modules, module)
	ms.soinfoAreaBase = ms.soinfoAreaBase + uint64(write_sz)
	if doInit {
		module.CallInit(ms.emu)
	}
	ms.logger.Debug().Msgf("finish load lib %s base 0x%08X", filename, loadBase)
	return module, nil
}

func (ms *Modules) elfGetSymVal(elfbase uint32, sym dyn) (uint32, bool) {
	if symAddr, exist := ms.symbolHooks[sym.Name]; exist {
		return uint32(symAddr), true
	}
	if sym.StShndx == SHN_UNDEF {
		target, resolved := ms.elfLookupSymbol(sym.Name)
		// target is none
		if !resolved {
			// Extern symbol not found
			if uint16(sym.StInfoBind) == STB_WEAK {
				// Weak symbol initialized
				return 0, true
			}else{
				ms.logger.Debug().Str("symName", sym.Name).Msg("undefined external symbol")
				return 0, false
			}
		}
		return target, true
	}else if sym.StShndx == SHN_ABS {
		// Absolute symbol.
		return elfbase + sym.StValue, true
	}
	// Internally defined symbol.
	return elfbase + sym.StValue, true
}
func (ms *Modules) elfLookupSymbol(name string) (uint32, bool) {
	for _, md := range ms.modules {
		if addr, exist := md.symbols[name]; exist {
			if addr != 0 {
				return addr, true
			}
		}
	}
	return 0, false
}












