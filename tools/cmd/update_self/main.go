// License: GPLv3 Copyright: 2022, Kovid Goyal, <kovid at kovidgoyal.net>

package update_self

import (
	"fmt"
	"kitty"
	"os"
	"path/filepath"
	"runtime"

	"kitty/tools/cli"
	"kitty/tools/tty"
	"kitty/tools/tui"
	"kitty/tools/utils"

	"golang.org/x/sys/unix"
)

var _ = fmt.Print

type Options struct {
	FetchVersion string
}

func fetch_latest_version() (string, error) {
	b, err := utils.DownloadAsSlice("https://sw.kovidgoyal.net/kitty/current-version.txt", nil)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch the latest available kitty version: %w", err)
	}
	return string(b), nil
}

func update_self(version string) (err error) {
	exe := ""
	exe, err = os.Executable()
	if err != nil {
		return fmt.Errorf("Failed to determine path to kitty-tool: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}
	if !kitty.IsStandaloneBuild {
		return fmt.Errorf("This is not a standalone kitty-tool executable. You must update all of kitty instead.")
	}
	rv := "v" + version
	if version == "nightly" {
		rv = version
	}
	if version == "latest" {
		rv, err = fetch_latest_version()
		if err != nil {
			return
		}
	}
	dest, err := os.CreateTemp(filepath.Dir(exe), "kitty-tool.")
	if err != nil {
		return err
	}
	defer func() { os.Remove(dest.Name()) }()

	url := fmt.Sprintf("https://github.com/kovidgoyal/kitty/releases/download/%s/kitty-tool-%s-%s", rv, runtime.GOOS, runtime.GOARCH)
	if !tty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println("Downloading:", url)
		err = utils.DownloadToFile(exe, url, nil, nil)
		if err != nil {
			return err
		}
		fmt.Println("Downloaded to:", exe)
	} else {
		err = tui.DownloadFileWithProgress(exe, url, true)
		if err != nil {
			return err
		}
	}
	fmt.Print("Updated to: ")
	return unix.Exec(exe, []string{"kitty-tool", "--version"}, os.Environ())
}

func EntryPoint(root *cli.Command) {
	sc := root.AddSubCommand(&cli.Command{
		Name:             "update-self",
		Usage:            "update-self [options ...]",
		ShortDescription: "Update this kitty-tool binary",
		HelpText:         "Update this kitty-tool binary in place to the latest available version.",
		Run: func(cmd *cli.Command, args []string) (ret int, err error) {
			if len(args) != 0 {
				return 1, fmt.Errorf("No command line arguments are allowed")
			}
			opts := &Options{}
			err = cmd.GetOptionValues(opts)
			if err != nil {
				return 1, err
			}
			return 0, update_self(opts.FetchVersion)
		},
	})
	sc.Add(cli.OptionSpec{
		Name:    "--fetch-version",
		Default: "latest",
		Help:    "The version to fetch. The special words :code:`latest` and :code:`nightly` fetch the latest stable and nightly release respectively. Other values can be, for example: 0.27.1.",
	})
}