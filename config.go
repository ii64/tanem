package emulator

import (
	"os"
	"encoding/json"
)

var (
	STACK_ADDR uint64 = 0x10000000
	STACK_SIZE uint64 = 0x00100000

	HOOK_MEMORY_BASE uint64 = 0x01000000
	HOOK_MEMORY_SIZE uint64 = 0x00200000

	MAP_ALLOC_BASE uint64 = 0x30000000
	MAP_ALLOC_SIZE uint64 = 0xA0000000-MAP_ALLOC_BASE

	BASE_ADDR uint64 = 0xCBBCB000

	PAGE_SIZE uint64 = 0x1000

	STACK_OFFSET uint64 = 8

	WRITE_FSTAT_TIMES = true
)

type Config struct {
	PkgName     string   `json:"pkg_name"`
	Pid         int      `json:"pid"`
	Uid         int      `json:"uid"`
	AndroidID   string   `json:"android_id"`
	Ip          string   `json:"ip"`
	Mac         []byte   `json:"mac"`
}

func NewDefaultConfig() *Config {
	return &Config{
		PkgName: "com.example",
		Pid: 4386,
		Uid: 10023,
		AndroidID: "39cc04a2ae83db0b",
		Ip: "192.168.43.22",
		Mac: []byte{204, 250, 166, 0, 138, 169},
	}
}

func LoadOrCreateConfig(path string, c *Config) error {
	err := LoadConfig(path, c)
	if err != nil && os.IsNotExist(err) {
		return SaveConfig(path, c)
	}
	return err
}

func LoadConfig(path string, c *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return err
	}
	return nil
}
func SaveConfig(path string, c *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return nil
	}
	encx := json.NewEncoder(f)
	return encx.Encode(c)
}