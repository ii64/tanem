package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	tn "github.com/ii64/tanem"
	"github.com/mattn/go-colorable"
	cli "github.com/urfave/cli/v2"
)

const (
	// BANNER for startup
	BANNER = "\033[1;37m   __                           \n  / /_____ _____  ___  ____ ___ \n / __/ __ `/ __ \\/ _ \\/ __ `__ \\\n/ /_/ /_/ / / / /  __/ / / / / /\n\\__/\\__,_/_/ /_/\\___/_/ /_/ /_/\033[0m \033[1;36mv{version}\033[0m\n"
)

var (
	// GIT_COMMIT git rev
	GIT_COMMIT = "<none>"
	// CLI_VERSION cli version
	CLI_VERSION = "0.1.1"
)
var (
	// OUT windows ansi color compatible
	OUT = colorable.NewColorableStdout()
)

func newTanemCmd(args []string) error {
	app := &cli.App{
		Name:  "tanem",
		Usage: "The Android Native Emulator",
		Flags: []cli.Flag{
			//config
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c", "conf"},
				Usage:   "Load configuration file",
				Value:   "default.json",
			},
			&cli.BoolFlag{
				Name:    "environ",
				Aliases: []string{"env"},
				Usage:   "Show environment",
			},
			//vfs
			&cli.StringFlag{
				Name:  "vfs",
				Usage: "Virtual file system root directory",
				Value: "vfs",
			},
			//vfsinstSet
			&cli.BoolFlag{
				Name:  "no-vfp-inst-set",
				Usage: "Disable fp vfp",
				Value: false,
			},
			//log as json
			&cli.BoolFlag{
				Name:  "jsonlog",
				Usage: "Log as json format",
			},
			//
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show version",
			},
		},
		Action: func(ctx *cli.Context) error {
			startTime := time.Now()
			defer func() {
				fmt.Printf("\nExec time: %s\n", time.Now().Sub(startTime))
			}()
			if ctx.Bool("version") {
				cmdVersion()
				return nil
			}
			opt := &tn.Options{
				VfsRoot:    ctx.String("vfs"),
				ConfigPath: ctx.String("config"),
				VfpInstSet: !ctx.Bool("no-vfp-inst-set"),
				LogAs:      tn.ConsoleLog,
				LogColor:   true,
			}
			if ctx.Bool("jsonlog") {
				opt.LogAs = tn.JsonLog
				opt.LogColor = false
			}
			emu, err := tn.NewEmulator(opt)
			if err != nil {
				return err
			}

			// TODO
			_ = emu

			return nil
		},
	}
	showBanner()
	return app.Run(args)
}

func showBanner() {
	fmt.Fprintf(OUT, "%s\n", strings.Replace(BANNER, "{version}", CLI_VERSION, -1))
}

func cmdVersion() {
	var formatted = [][]string{
		{"Version:", "v" + CLI_VERSION},
		{"Git commit:", GIT_COMMIT},
		{"Go version:", runtime.Version()},
		{"OS/Arch:", runtime.GOOS + "/" + runtime.GOARCH},
	}
	var maxLen = 0
	for _, col := range formatted {
		if len(col[0]) > maxLen {
			maxLen = len(col[0])
		}
	}
	for _, col := range formatted {
		fmt.Printf("%s     %s\n", col[0]+strings.Repeat(" ", maxLen-len(col[0])), col[1])
	}
}

func main() {
	if err := newTanemCmd(os.Args); err != nil {
		panic(err)
	}
}
