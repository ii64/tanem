package cmd

import (
	"fmt"
	"time"
	"strings"
	"runtime"
	cli "github.com/urfave/cli/v2"
	tn  "github.com/ii64/tanem"
)

func NewTanemCmd(args []string) error {
	app := &cli.App{
		Name: "tanem",
		Usage: "The Android Native Emulator",
		Flags: []cli.Flag{
			//config
			&cli.StringFlag{
				Name: "config",
				Aliases: []string{"c", "conf"},
				Usage: "Load configuration file",
				Value: "default.json",
			},
			&cli.BoolFlag{
				Name: "environ",
				Aliases: []string{"env"},
				Usage: "Show environment",
			},
			//vfs
			&cli.StringFlag{
				Name: "vfs",
				Usage: "Virtual file system root directory",
				Value: "vfs",
			},
			//vfsinstSet
			&cli.BoolFlag{
				Name: "no-vfp-inst-set",
				Usage: "Disable fp vfp",
				Value: false,
			},
			//log as json
			&cli.BoolFlag{
				Name: "jsonlog",
				Usage: "Log as json format",
			},
			//
			&cli.BoolFlag{
				Name: "version",
				Aliases: []string{"v"},
				Usage: "Show version",
			},
		},
		Action: func(ctx *cli.Context) error {
			startTime := time.Now()
			defer func() {
				fmt.Printf("\nExec time: %s\n", time.Now().Sub(startTime))
			}()
			if ctx.Bool("version") {
				CmdVersion()
				return nil
			}
			opt := &tn.Options{
				VfsRoot: ctx.String("vfs"),
				ConfigPath: ctx.String("config"),
				VfpInstSet: !ctx.Bool("no-vfp-inst-set"),
				LogAs: tn.ConsoleLog,
				LogColor: true,
			}
			if ctx.Bool("jsonlog") {
				opt.LogAs = tn.JsonLog
				opt.LogColor = false
			}
			emu, err := tn.NewEmulator(opt)
			if err != nil {
				return err
			}
			

			_ = emu


			return nil
		},
	}
	return app.Run(args)
}

var (
	GIT_COMMIT = ""
)

func CmdVersion() {
	var formatted = [][]string{
		{"Git commit:", GIT_COMMIT},
		{"Go version:", runtime.Version()},
		{"OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH},
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
