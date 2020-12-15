package emulator

import (
	"os"
	"io"
	"fmt"
	"sort"
	log "github.com/rs/zerolog/log"
	uc  "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

type filemap struct {
	offset uint64
	sz     uint64
	fs     *os.File
}
type MemoryMap struct {
	mu    uc.Unicorn

	allocMinAddr uint64
	allocMaxAddr uint64
	fileMapAddr  map[uint64]filemap
}
func NewMemoryMap(mu uc.Unicorn, allocMinAddr, allocMaxAddr uint64) *MemoryMap {
	return &MemoryMap{
		mu: mu,
		allocMinAddr: allocMinAddr,
		allocMaxAddr: allocMaxAddr,
		fileMapAddr: map[uint64]filemap{},
	}
}
func (m *MemoryMap) CheckAddr(addr uint64, prot int) (bool, error) {
	memregs, err := m.mu.MemRegions()
	if err != nil {
		return false, err
	}	
	for _, memreg := range memregs {
		if addr >= memreg.Begin && addr < memreg.End {
			return true, nil
		}
	}
	return false, nil
}
func (m *MemoryMap) IsMultiple(addr uint64) bool {
	return addr % PAGE_SIZE == 0
}
func (m *MemoryMap) IsOverlap(addr1,end1, addr2,end2 uint64) bool {
	return (addr1 <= addr2 && end1 >= end2) || (addr2 <= addr1 && end2 >= end1) || (end1 > addr2 && addr1 < end2) || (end2 > addr1 && addr2 < end1)
}
func (m *MemoryMap) mapInternal(address, size uint64, prot int) (uint64, error) {
	var (
		memregs []*uc.MemRegion
		err error
	)
	if prot == 0 {
		prot = uc.PROT_READ | uc.PROT_WRITE
	}
	if size <= 0 {
		return 0, ErrHeapLessEqualZero
	}
	if address == 0 {
		memregs, err = m.mu.MemRegions()
		if err != nil {
			return 0, err
		}
		sort.Slice(memregs, func(i, j int) bool {
			// previous using <
			return memregs[i].Begin > memregs[j].Begin
		})
		var map_base uint64

		lregs := len(memregs)
		if lregs < 1 {
			map_base = m.allocMinAddr
		}else{
			prefer_start := m.allocMinAddr
			next_loop := true
			for {
				if !next_loop {
					break
				}
				next_loop = false
				for _, r := range memregs {
					if m.IsOverlap(prefer_start, prefer_start+size, r.Begin, r.End+1) {
						prefer_start = r.End+1
						next_loop = true
						break
					}
				}
			}
			map_base = prefer_start
		}
		if map_base > m.allocMaxAddr || map_base < m.allocMinAddr {
			return 0, fmt.Errorf("%w: map_base:0x%08X out of range (0x%08X-0x%08X)!!!", ErrMmapError, map_base, m.allocMinAddr, m.allocMaxAddr)
		}
		//print("before mem_map addr:0x%08X, sz:0x%08X"%(map_base, size))
		log.Debug().Msgf("before mem_map addr:0x%08X, sz:0x%08X", map_base, size)
		err = m.mu.MemMapProt(map_base, size, prot)
		if err != nil {
			return 0, err
		}
		return map_base, nil
	}else{
		//MAP_FIXED
		err = m.mu.MemMapProt(address, size, prot)
		if errno, able := err.(uc.UcError); err != nil && able && errno == uc.ERR_MAP {
			blocks := map[uint64]bool{}
			extra_protect := map[uint64]bool{}
			for b := address; b < (address+size); b=(b+0x1000) {
				blocks[b] = true
			}
			memregs, err = m.mu.MemRegions()
			if err != nil {
				return 0, err
			}
			for _, memreg := range memregs {
				raddr := memreg.Begin
				rend  := memreg.End+1
				for b := raddr; b < rend; b=(b+0x1000) {
					if _, inx := blocks[b]; inx {
						delete(blocks, b)
						extra_protect[b] = true
					}
				}
			}
			for b_map, _ := range blocks {
				err = m.mu.MemMapProt(b_map, 0x1000, prot)
				if err != nil {
					return 0, err
								}
			}
			for b_protect, _ := range extra_protect {
				err = m.mu.MemProtect(b_protect, 0x1000, prot)
				if err != nil {
					return 0, err
				}
			}
			return address, nil
		}
		return address, nil
	}
	return 0, ErrUnknownBehavior
}
func (m *MemoryMap) Map(address, size uint64, prot int, vf interface{}, offset uint64) (uint64, error) {
	if !m.IsMultiple(address) {
		return 0, fmt.Errorf("%w: (%d mod %d = %d)", ErrMapAddrNotMultiple, address, PAGE_SIZE, address % PAGE_SIZE)
	}
	log.Debug().Msgf("mapped addr:0x%08X end:0x%08X sz:0x%08X off:0x%08X", address, address+size, size, offset)
	//print("map addr:0x%08X, end:0x%08X, sz:0x%08X off=0x%08X"%(address, address+size, size, offset))
	al_address := address
	al_size := PageEnd(al_address+size) - al_address
	res_addr, err := m.mapInternal(al_address, al_size, prot)
	if err != nil {
		return 0, fmt.Errorf("map error: %w", err)
	}
	if o, ok := vf.(*VirtualFile); ok {
		oriOff, err := o.fo.Seek(0, 1)
		if err != nil {
			return 0, fmt.Errorf("map error: %w", err)
		}
		//change
		_, err = o.fo.Seek(int64(offset), 0)
		if err != nil {
			return 0, fmt.Errorf("map error: %w", err)
		}
		data, err := m.readFully(o.fo, size)
		if err != nil {
			log.Debug().Err(err).Msg("read fully error")
		}
		log.Debug().Msgf("read for offset %d sz %d data sz:%d",
			offset, size, len(data))
		err = m.mu.MemWrite(res_addr, data)
		if err != nil {
			return 0, fmt.Errorf("map fs mem error: %w", err)
		}
		m.fileMapAddr[al_address] = filemap{
			sz: al_address+al_size,
			offset: offset,
			fs: o.fo,
		}
		_, err = o.fo.Seek(oriOff, 0)//
		if err != nil {
			return 0, fmt.Errorf("map error: %w", err)
		}
	}
	return res_addr, nil
}
func (m *MemoryMap) readFully(fd *os.File, size uint64) ([]byte, error) {
	resx := make([]byte, int(size))
	cnt, err := fd.Read(resx)
	if err != nil {
		return nil, err
	}
	return resx[:cnt], nil
}
func (m *MemoryMap) Protect(addr, lenx uint64, prot int) error {
	if !m.IsMultiple(addr) {
		return fmt.Errorf("%w: (%d mod %d = %d)", ErrMapAddrNotMultiple, addr, PAGE_SIZE, addr % PAGE_SIZE)
	}
	len_in := PageEnd(addr-lenx) - addr
	err := m.mu.MemProtect(addr, len_in, prot)
	if err != nil {
		//TODO: just for debug
		log.Debug().Msgf("Warnning mprotect with addr:0x%08X len:0x%08X prot:0x%08X failed!!!", addr, lenx, prot)
	}
	return err
}
func (m *MemoryMap) Unmap(addr, size uint64) error {
	if !m.IsMultiple(addr) {
		return fmt.Errorf("%w: (%d mod %d = %d)", ErrMapAddrNotMultiple, addr, PAGE_SIZE, addr % PAGE_SIZE)
	}
	size = PageEnd(addr+size) - addr
	log.Debug().Msgf("unmap 0x%08X sz:0x%08X end=0x%08X", addr,size, addr+size)
	if fx, exist := m.fileMapAddr[addr]; exist {
		if addr+size != fx.sz {
			return fmt.Errorf("unmap error, range 0x%08X-0x%08X does not match file map range 0x%08X-0x%08X from file %s",
				addr, addr+size, addr, fx.sz, fx.fs.Name())
		}
		delete(m.fileMapAddr, addr)
	}
	return m.mu.MemUnmap(addr, size)
}

func (m *MemoryMap) getMapAttr(start, end uint64) (uint64, string) {
	for addr, v := range m.fileMapAddr {
		mstart := addr
		mend := v.sz
		if start >= mstart && end <= mend {
			vf := v.fs
			return v.offset, vf.Name()
		}
	}
	return 0, ""
}
type attrMem struct {
	Begin, End uint64
	Prot string
	Offset uint64
	Name string
}
func (m *MemoryMap) getAttrs(mreg *uc.MemRegion) (ret *attrMem) {
	r := ""
	kx := mreg.Prot
	if kx & 0x1 == 1 {
		r = r + "r"
	}else{
		r = r + "-"
	}
	if kx & 0x2 == 1 {
		r = r + "w"
	}else{
		r = r + "-"
	}
	if kx & 0x3 == 1 {
		r = r + "x"
	}else{
		r = r + "-"
	}
	r = r + "p"
	off, name := m.getMapAttr(mreg.Begin, mreg.End+1)
	return &attrMem{
		Begin: mreg.Begin, End: mreg.End,
		Prot: r,
		Offset: off,
		Name: name,
	}
}
func (m *MemoryMap) DumpMaps(wrt io.Writer) error {
	memregs, err := m.mu.MemRegions()
	if err != nil {
		return err
	}
	sort.Slice(memregs, func(i, j int) bool {
		return memregs[i].Begin > memregs[j].Begin
	})
	//不有memreg
	if len(memregs) < 1 {
		return nil
	}
	s_attr := m.getAttrs(memregs[0])
	output := []*attrMem{}
	start := s_attr.Begin
	for _, mm := range memregs[1:] {
		attr := m.getAttrs(mm)
		if s_attr.End == attr.Begin && (
			s_attr.Prot == attr.Prot &&
			s_attr.Offset == attr.Offset &&
			s_attr.Name == attr.Name) {
		}else{
			output = append(output, &attrMem{
				Begin: start,
				End: s_attr.End,
				Prot: s_attr.Prot,
				Offset: s_attr.Offset,
				Name: s_attr.Name,
			})
			start = attr.Begin
		}
		s_attr = attr
	}
	output = append(output, &attrMem{
		Begin: start,
		End:s_attr.End,
		Prot: s_attr.Prot,
		Offset: s_attr.Offset,
		Name: s_attr.Name,
	})

	for _, item := range output {
		fmrt := fmt.Sprintf(
			"%08x-%08x %s %08x 00:00 0 \t\t %s\n",
			item.Begin, item.End, item.Prot,
			item.Offset, item.Name,
		)
		wrt.Write([]byte(fmrt))
	}
	return nil
}
















