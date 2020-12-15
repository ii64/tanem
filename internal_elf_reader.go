package emulator

import (
	"os"
	"bytes"
	"github.com/pkg/errors"
	bin "encoding/binary"
	log "github.com/rs/zerolog/log"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

var (
	// PT
	PT_NULL      uint32 = 0
	PT_LOAD      uint32 = 1
	PT_DYNAMIC   uint32 = 2
	PT_INTERP    uint32 = 3
	PT_NOTE      uint32 = 4
	PT_SHLIB     uint32 = 5
	PT_PHDR      uint32 = 6
	// DT
	DT_NULL	     uint32 = 0
	DT_NEEDED	 uint32 = 1
	DT_PLTRELSZ	 uint32 = 2
	DT_PLTGOT	 uint32 = 3
	DT_HASH		 uint32 = 4
	DT_STRTAB	 uint32 = 5
	DT_SYMTAB	 uint32 = 6
	DT_RELA		 uint32 = 7
	DT_RELASZ	 uint32 = 8
	DT_RELAENT	 uint32 = 9
	DT_STRSZ	 uint32 = 10
	DT_SYMENT	 uint32 = 11
	DT_INIT      uint32 = 0x0c
	DT_INIT_ARRAY    uint32 = 0x19
	DT_FINI_ARRAY    uint32 = 0x1a
	DT_INIT_ARRAYSZ  uint32 = 0x1b
	DT_FINI_ARRAYSZ  uint32 = 0x1c
	DT_SONAME	  uint32 = 14
	DT_RPATH 	  uint32 = 15
	DT_SYMBOLIC	  uint32 = 16
	DT_REL	      uint32 = 17
	DT_RELSZ	  uint32 = 18
	DT_RELENT	  uint32 = 19
	DT_PLTREL	  uint32 = 20
	DT_DEBUG      uint32 = 21
	DT_TEXTREL    uint32 = 22
	DT_JMPREL	  uint32 = 23
	DT_LOPROC	  uint32 = 0x70000000
	DT_HIPROC	  uint32 = 0x7fffffff
	// SHN
	SHN_UNDEF         uint16 = 0
	SHN_LORESERVE     uint16 = 0xff00
	SHN_LOPROC	      uint16 = 0xff00
	SHN_HIPROC	      uint16 = 0xff1f
	SHN_ABS	          uint16 = 0xfff1
	SHN_COMMON	      uint16 = 0xfff2
	SHN_HIRESERVE	  uint16 = 0xffff
	SHN_MIPS_ACCOMON  uint16 = 0xff00
	// STB
	STB_LOCAL     uint16 = 0
	STB_GLOBAL    uint16 = 1
	STB_WEAK      uint16 = 2
	STT_NOTYPE    uint16 = 0
	STT_OBJECT    uint16 = 1
	STT_FUNC      uint16 = 2
	STT_SECTION   uint16 = 3
	STT_FILE      uint16 = 4
)

type ELFReader struct {
	file      *os.File
	filename  string
	phoff     uint64
	phdrNum   uint64

	initArrayOff  uint32
	initArraySize uint32
	initOff       uint32

	nbucket       uint32
	nchain        uint32
	bucket        uint32
	chain         uint32
	pltGot        uint32
	pltRel        uint32
	pltRelCount   uint32
	rel           uint32
	relCount      uint32

	dynOff        uint32
	dynStrBuf     []byte
	dynStrOff     uint32
	dynStrSz      uint32
	dynSymOff     uint32

	soNeeded  []string
	phdrs     []phdr
	loads     []phdr
	dynSym    []dyn
	rels      rels
	sz        uint32
}
func NewELFReader(filename string) (*ELFReader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open ELF file")
	}
	elfr := &ELFReader{
		initArrayOff: 0,
		initArraySize: 0,
		initOff: 0,
		phdrs: []phdr{},
		loads: []phdr{},
		dynSym: []dyn{},
		soNeeded: []string{},
		rels: rels{},
		filename: filename,
		file: f,
		sz: 0,
	}
	ehdr32_sz    := 52
	phdr32_sz    := 32
	elf32_dyn_sz := 8
	elf32_sym_sz := 16
	elf32_rel_sz := 8
	ehdr32 := make([]byte, ehdr32_sz)
	cnt, err := f.Read(ehdr32)
	ehdr32 = ehdr32[:cnt]
	if err != nil || cnt != ehdr32_sz {
		return nil, errors.Wrap(err, "ehdr32 count mismatch")
	}
	// LE
	//_, _ , _, _, _, phoff, _, _, _, _, phdr_num, _, _, _ = struct.unpack("<16sHHIIIIIHHHHHH", ehdr_bytes)
	// <16s H H I I [I] I I H H [H] H H H

	// uint32
	phoff    := bin.LittleEndian.Uint32(ehdr32[28:28+4])
	// unsigned short, uint16
	phdrNum := bin.LittleEndian.Uint16(ehdr32[44:44+2])
	log.Debug().Msgf("ELF reader: phoff:%08X phdrNum:%08X",phoff, phdrNum)
	elfr.phoff   = uint64(phoff)
	elfr.phdrNum = uint64(phdrNum)



	_, err = f.Seek(int64(phoff), 0)
	if err != nil {
		return nil, errors.Wrap(err, "seek to phoff failed")
	}

	elfr.sz = 0
	var (
		dynOff uint32 = 0
		i uint16 // sz phdr
	)
	for i = 0; i < phdrNum; i++ {
		tmp := make([]byte, phdr32_sz)
		cnt, err := f.Read(tmp)
		tmp = tmp[:cnt]
		if err != nil {
			return nil, ErrELFReadFail
		}
		//"<IIIIIIII"
		//p_type, p_offset, p_vaddr, p_paddr, p_filesz, p_memsz, p_flags, p_align
		phdrx := phdr{
			pType:   bin.LittleEndian.Uint32(tmp[0:4]),
			pOffset: bin.LittleEndian.Uint32(tmp[4:8]),
			pVaddr:  bin.LittleEndian.Uint32(tmp[8:12]),
			pPaddr:  bin.LittleEndian.Uint32(tmp[12:16]),
			pFilesz: bin.LittleEndian.Uint32(tmp[16:20]),
			pMemsz:  bin.LittleEndian.Uint32(tmp[20:24]),
			pFlags:  bin.LittleEndian.Uint32(tmp[24:28]),
			pAlign:  bin.LittleEndian.Uint32(tmp[28:32]),
		}
		elfr.phdrs = append(elfr.phdrs, phdrx)
		if phdrx.pType == PT_DYNAMIC {
			dynOff = phdrx.pOffset
		}else if phdrx.pType == PT_LOAD {
			elfr.loads = append(elfr.loads, phdrx)
		}
		elfr.sz = elfr.sz + phdrx.pMemsz
	}
	if !(dynOff > 0) {
		return nil, ErrELFReadNoDynamic
	}
	elfr.dynOff = dynOff
	_, err = f.Seek(int64(dynOff), 0)
	if err != nil {
		return nil, errors.Wrap(err, "seek failed to dyn offset")
	}
	var (
		readPos = dynOff

		dynStrOff uint32 = 0
		dynStrSz  uint32 = 0
		dynStrBuf []byte = []byte{}

		dynSymOff   uint32 = 0
		nsymbol     uint32 = 0 //
		foundNsymbol  bool = false
		relOff      uint32 = 0
		relCount    uint32 = 0
		relpltOff   uint32 = 0
		relpltCount uint32 = 0
		dtNeeded           = []uint32{}
	)
	_ = dynStrBuf
	for {
		tmp := make([]byte, elf32_dyn_sz)
		cnt, err := f.Read(tmp)
		tmp = tmp[:cnt]
		if err != nil {
			return nil, errors.Wrap(err, "cannot read elf32 dyn bytes")
		}
		tmp = tmp[:cnt]
		dTag     := bin.LittleEndian.Uint32(tmp[0:4])
		dValPtr  := bin.LittleEndian.Uint32(tmp[4:8])
		if dTag == DT_NULL {
			break
		}else if dTag == DT_RELA {
			return nil, ErrELF64NotSupported
		}else if dTag == DT_REL {
			relOff = dValPtr
		}else if dTag == DT_RELSZ {
			relCount = uint32(dValPtr) / uint32(elf32_rel_sz)
		}else if dTag == DT_JMPREL {
			relpltOff = dValPtr
		}else if dTag == DT_PLTRELSZ {
			relpltCount = uint32(dValPtr) / uint32(elf32_rel_sz)
		}else if dTag == DT_SYMTAB {
			dynSymOff = dValPtr
		}else if dTag == DT_STRTAB {
			dynStrOff = dValPtr
		}else if dTag == DT_STRSZ {
			dynStrSz = dValPtr
		}else if dTag == DT_HASH {
			curPos, err := f.Seek(0, 1)
			if err != nil {
				return nil, errors.Wrap(err, "cannot get current read position")
			}
			_, err = f.Seek(int64(dValPtr), 0)
			hashHeads := make([]byte,8)
			cnt, err = f.Read(hashHeads)
			hashHeads = hashHeads[:cnt]
			if err != nil {
				return nil, errors.Wrap(err, "cannot read DT_HASH")
			}
			_, err = f.Seek(curPos, 0)
			if err != nil {
				return nil, errors.Wrap(err, "cannot seek to readPos")
			}
			elfr.nbucket = bin.LittleEndian.Uint32(hashHeads[0:4])
			elfr.nchain  = bin.LittleEndian.Uint32(hashHeads[4:8])
			elfr.bucket  = dValPtr + 8
			elfr.chain   = dValPtr + 8 + (elfr.nbucket * 4)
			nsymbol = elfr.nchain
			foundNsymbol = true
		}else if dTag == DT_INIT {
			elfr.initOff = dValPtr
		}else if dTag == DT_INIT_ARRAY {
			elfr.initArrayOff = dValPtr
		}else if dTag == DT_INIT_ARRAYSZ {
			elfr.initArraySize = dValPtr
		}else if dTag == DT_NEEDED {
			dtNeeded = append(dtNeeded, dValPtr)
		}else if dTag == DT_PLTGOT {
			elfr.pltGot = dValPtr
		}
		readPos = readPos + uint32(len(tmp))
	}
	if !foundNsymbol {
		return nil, ErrELFNHash
	}
	elfr.dynStrOff = dynStrOff
	elfr.dynSymOff = dynSymOff
	
	elfr.dynStrSz  = dynStrSz

	elfr.pltRel      = relpltOff
	elfr.pltRelCount = relpltCount

	elfr.rel      = relOff
	elfr.relCount = relCount

	_, err = f.Seek(int64(dynStrOff), 0)
	if err != nil {
		return nil, errors.Wrap(err, "cannot seek to dynStrOff")
	}
	tmp := make([]byte, dynStrSz)
	cnt, err = f.Read(tmp)
	tmp = tmp[:cnt]
	elfr.dynStrBuf = tmp

	_, err = f.Seek(int64(dynSymOff), 0)
	if err != nil {
		return nil, errors.Wrap(err, "cannot seek to dynSymOff")
	}
	var jj uint32
	for jj = 0; jj < nsymbol; jj++ {
		symBytes := make([]byte, elf32_sym_sz)
		cnt, err := f.Read(symBytes)
		symBytes = symBytes[:cnt]
		if err != nil {
			return nil, errors.Wrap(err, "cannot read elf32 sym bytes")
		}
		// "<IIIccH"
		// 12 + 4 = 16
		stName  := bin.LittleEndian.Uint32(symBytes[0:4])
		stVal   := bin.LittleEndian.Uint32(symBytes[4:8])
		stSize  := bin.LittleEndian.Uint32(symBytes[8:12])
		// 255
		stInfo  := uint32(symBytes[12:13][0])
		// 255
		stOther := uint32(symBytes[13:14][0])
		stShndx := bin.LittleEndian.Uint16(symBytes[14:16])
		var intStInfo uint64 = uint64(stInfo)
		//if stInfo == 2 {
		//	intStInfo = bin.LittleEndian.Uint16()
		//}
		stInfoBind := elf_st_bind(intStInfo)
		stInfoType := elf_st_type(intStInfo)
		name := elfr.StNameToName(stName)
		dy := dyn{
			Name: name,
			StName: stName,
			StValue: stVal,
			StSize: stSize,
			StInfo: stInfo,
			StOther: stOther,
			StShndx: stShndx,

			StInfoBind: stInfoBind,
			StInfoType: stInfoType,
		}
		elfr.dynSym = append(elfr.dynSym, dy)
	}

	_, err = f.Seek(int64(relOff), 0)
	if err != nil {
		return nil, errors.Wrap(err, "cannot seek to relOff")
	}

//	log.Debug().Msgf(">>> relOffset %08X count %08X", relOff, relCount)
	relTable := []relx{}
	for jj = 0; jj < relCount; jj++ {
		tmp := make([]byte, elf32_rel_sz)
		cnt, err := f.Read(tmp)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read elf32 rel bytes")
		}
		tmp = tmp[:cnt]
		rOffset := bin.LittleEndian.Uint32(tmp[0:4])
		rInfo := bin.LittleEndian.Uint32(tmp[4:8])
		rInfoSym  := elf32_r_sym(rInfo)
		rInfoType := elf32_r_type(rInfo)
		relTable = append(relTable, relx{
			ROffset: rOffset,
			RInfo: rInfo,
			RInfoType: rInfoType,
			RInfoSym: rInfoSym,
		})
	}
	elfr.rels.dynrel = relTable

	_, err = f.Seek(int64(relpltOff), 0)
	if err != nil {
		return nil, errors.Wrap(err, "cannot seek to relPltOff")
	}
	relpltTable := []relx{}
	for jj = 0; jj < relpltCount; jj++ {
		tmp := make([]byte, elf32_rel_sz)
		cnt, err := f.Read(tmp)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read elf32 rel bytes")
		}
		tmp = tmp[:cnt]
		rOffset := bin.LittleEndian.Uint32(tmp[0:4])
		rInfo := bin.LittleEndian.Uint32(tmp[4:8])		
		rInfoSym  := elf32_r_sym(rInfo)
		rInfoType := elf32_r_type(rInfo)
		relpltTable = append(relpltTable, relx{
			ROffset: rOffset,
			RInfo: rInfo,
			RInfoType: rInfoType,
			RInfoSym: rInfoSym,
		})
	}
	elfr.rels.relplt = relpltTable
//	log.Debug().Str("filename", filename).Msg("load ok")
	for _, needed := range dtNeeded {
		endId := bytes.Index(elfr.dynStrBuf[int(needed):], []byte{0x0})
		if endId < 0 {
			continue
		}
		elfr.soNeeded = append(elfr.soNeeded, string( elfr.dynStrBuf[int(needed):int(needed)+endId]) )
	}
	return elfr, nil
}
func (elfr *ELFReader) StNameToName(stname uint32) string {
	endId := bytes.Index(elfr.dynStrBuf[int(stname):], []byte{0x0})
//	log.Debug().Int("endId", endId).Uint32("stName", stname).Msg("STNAMETONAME")
	if endId < 0 {
		return  ""
	}
	return string(elfr.dynStrBuf[int(stname):int(stname)+endId])
}

func (elfr *ELFReader) GetLoad() []phdr {
	return elfr.loads
}
func (elfr *ELFReader) GetSymbols() []dyn {
	return elfr.dynSym
}
func (elfr *ELFReader) GetRels() rels {
	return elfr.rels
}
func (elfr *ELFReader) GetDynStringByRelSym(relSym int) string {
	nsym := len(elfr.dynSym)
	if relSym > nsym {
		return ""
	}
	sym := elfr.dynSym[relSym]
	return elfr.StNameToName(sym.StName)
}
func (elfr *ELFReader) GetInitArray() (uint32	, uint32) {
	return elfr.initArrayOff, elfr.initArraySize
}
func (elfr *ELFReader) GetInit() uint32 {
	return elfr.initOff
}
func (elfr *ELFReader) GetSoNeeded() []string {
	return elfr.soNeeded
}

func (elfr *ELFReader) WriteSoInfo(mu uc.Unicorn, loadBase, infoBase uint64) (uint32, error) {
	//在虚拟机中构造一个soinfo结构
	if !(len(elfr.filename)<128) {
		return 0, ErrELFSOFileTooLong
	}
	var errorsx = []error{
		//name
		WriteUtf8(mu, infoBase+0, []byte(elfr.filename)),
		//phdr
		mu.MemWrite(infoBase+128, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase+elfr.phoff))
			return newVal
		}()),
		//phnum
		mu.MemWrite(infoBase+132, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(elfr.phdrNum))
			return newVal
		}()),
		//entry
		mu.MemWrite(infoBase+136, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//base
		mu.MemWrite(infoBase+140, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase))
			return newVal
		}()),
		//size
		mu.MemWrite(infoBase+144, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, elfr.sz)
			return newVal
		}()),
		//unused1
		mu.MemWrite(infoBase+148, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//dynamic
		mu.MemWrite(infoBase+152, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.dynOff)
			return newVal
		}()),
		//unused2
		mu.MemWrite(infoBase+156, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//unused3
		mu.MemWrite(infoBase+160, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//next
		mu.MemWrite(infoBase+164, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//flags
		mu.MemWrite(infoBase+168, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//strtab
		mu.MemWrite(infoBase+172, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.dynStrOff)
			return newVal
		}()),
		//symtab
		mu.MemWrite(infoBase+176, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.dynSymOff)
			return newVal
		}()),
		//nbucket
		mu.MemWrite(infoBase+180, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, elfr.nbucket)
			return newVal
		}()),
		//nchain
		mu.MemWrite(infoBase+184, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, elfr.nchain)
			return newVal
		}()),

		//bucket
		mu.MemWrite(infoBase+188, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.bucket)
			return newVal
		}()),
		//nchain
		mu.MemWrite(infoBase+192, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.nbucket)
			return newVal
		}()),

		//plt_got
		mu.MemWrite(infoBase+196, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.pltGot)
			return newVal
		}()),

		//plt_rel
		mu.MemWrite(infoBase+200, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.pltRel)
			return newVal
		}()),
		//plt_rel_count
		mu.MemWrite(infoBase+204, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(elfr.pltRelCount))
			return newVal
		}()),

		//rel
		mu.MemWrite(infoBase+208, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.rel)
			return newVal
		}()),
		//rel_count
		mu.MemWrite(infoBase+212, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, elfr.relCount)
			return newVal
		}()),

		//preinit_array
		mu.MemWrite(infoBase+216, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//preinit_array_count
		mu.MemWrite(infoBase+220, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),

		//init_array
		mu.MemWrite(infoBase+224, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.initArrayOff)
			return newVal
		}()),
		//init_array_count
		mu.MemWrite(infoBase+228, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, elfr.initArraySize/4)
			return newVal
		}()),

		//finit_array
		mu.MemWrite(infoBase+232, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//finit_array_count
		mu.MemWrite(infoBase+236, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),

		//init_func
		mu.MemWrite(infoBase+240, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, uint32(loadBase)+elfr.initOff)
			return newVal
		}()),
		//fini_func
		mu.MemWrite(infoBase+244, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),

		//ARM_exidx
		mu.MemWrite(infoBase+248, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),
		//ARM_exidx_count
		mu.MemWrite(infoBase+252, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 0)
			return newVal
		}()),

		//ref_count
		mu.MemWrite(infoBase+256, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 1)
			return newVal
		}()),

		//link_map
		mu.MemWrite(infoBase+260, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}),

		//constructors_called
		mu.MemWrite(infoBase+280, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 1)
			return newVal
		}()),

		//Elf32_Addr load_bias
		mu.MemWrite(infoBase+284, func()[]byte{
			newVal := make([]byte, 4)
			bin.LittleEndian.PutUint32(newVal, 1)
			return newVal
		}()),
	}
	//explicit
	for _, err := range errorsx {
		if err != nil {
			return 0, errors.Wrap(err, "write so info failed")
		}
	}
	var soinfo_sz uint32 = 288
	return soinfo_sz, nil
}


func elf32_r_sym(r uint32) uint32 {
	return r >> 8
}
func elf32_r_type(r uint32) uint32 {
	return r & 0xff
}
func elf_st_bind(st uint64) uint64 {
	return st >> 4
}
func elf_st_type(st uint64) uint64 {
	return st & 0xf
}

// rels name
type rels struct {
	dynrel []relx
	relplt []relx
}
type relx struct {
	ROffset    uint32
	RInfo      uint32
	RInfoType  uint32
	RInfoSym   uint32
}

type dyn struct {
	Name     string
	StName   uint32
	StValue  uint32
	StSize   uint32
	StInfo   uint32
	StOther  uint32
	StShndx  uint16

	StInfoBind uint64
	StInfoType uint64
}

type phdr struct {
	pType    uint32
	pOffset  uint32
	pVaddr   uint32
	pPaddr   uint32
	pFilesz  uint32
	pMemsz   uint32
	pFlags   uint32
	pAlign   uint32
}











