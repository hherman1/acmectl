package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"9fans.net/go/acme"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "acmectl",
	Short: "acmectl allows you to manipulate the acme editor easily from the shell",
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "new creates a new acme window and returns its ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		win, err := acme.New()
		if err != nil {
			return fmt.Errorf("create window: %w", err)
		}
		fmt.Println(win.ID())
		return nil
	},
	Args: cobra.NoArgs,
}

var ctlCmd = &cobra.Command{
	Use:   "ctl <winid> <ctl>",
	Short: "ctl sends an acme control message to the given window",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("parse winid: %w", err)
		}
		win, err := acme.Open(id, nil)
		if err != nil {
			return fmt.Errorf("open window: %w", err)
		}
		return win.Ctl(strings.Join(args[1:], " "))
	},
}

// Writes to the given file. Implements io.Writer
type writer struct {
	win *acme.Win
	f   string
}

func (w writer) Write(p []byte) (n int, err error) {
	return w.win.Write(w.f, p)
}

var writeCmd = &cobra.Command{
	Use:   "write <winid> <winfile>",
	Short: "write copies stdin to the window's winfile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("parse winid: %w", err)
		}
		win, err := acme.Open(id, nil)
		if err != nil {
			return fmt.Errorf("open window: %w", err)
		}
		_, err = win.Seek(args[1], 0, 2)
		if err != nil {
			return fmt.Errorf("seek to end: %w", err)
		}
		dst := writer{
			win: win,
			f:   args[1],
		}
		_, err = io.Copy(dst, os.Stdin)
		if err != nil {
			return fmt.Errorf("copy data: %w", err)
		}

		return nil
	},
}

var readCmd = &cobra.Command{
	Use:   "read <winid> <winfile>",
	Short: "read cats the entire contents of the given winfile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("parse winid: %w", err)
		}
		win, err := acme.Open(id, nil)
		if err != nil {
			return fmt.Errorf("open window: %w", err)
		}
		bs, err := win.ReadAll(args[1])
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		_, err = os.Stdout.Write(bs)
		if err != nil {
			return fmt.Errorf("write to stdout: %w", err)
		}
		return nil
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls <winid>",
	Short: "ls lists the available window files",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(`addr
body
ctl
data
event
tag
xdata`)
		return nil
	},
}

var onCmd = &cobra.Command{
	Use:   "onexec <winid> <event> <command...>",
	Short: "Intercepts matching execute events from  the given windows event channel. When found, runs the command. Exits when the window is closed.",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("parse winid: %w", err)
		}
		ev := args[1]
		cmdName := args[2]
		cargs := args[3:]

		win, err := acme.Open(id, nil)
		if err != nil {
			return fmt.Errorf("open window: %w", err)
		}

		for e := range win.EventChan() {
			switch e.C2 {
			case 'x', 'X': // execute
				if string(e.Text) == ev {
					exe := exec.Command(cmdName, cargs...)
					exe.Stdout = os.Stdout
					exe.Stderr = os.Stderr
					err := exe.Run()
					if err != nil {
						fmt.Println(err)
						continue
					}
					continue
				}
			}
			win.WriteEvent(e)
		}
		return nil
	},
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	rootCmd.AddCommand(newCmd, ctlCmd, writeCmd, readCmd, lsCmd, onCmd)
	return rootCmd.Execute()
}
