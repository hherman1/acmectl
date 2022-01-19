package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"9fans.net/go/acme"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "acmectl",
	Short:	"acmectl allows you to manipulate the acme editor easily from the shell",
}

var newCmd = &cobra.Command{
	Use: "new",
	Short: 	"new creates a new acme window and returns its ID",
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
	Use: "ctl <winid> <ctl>",
	Short: 	"ctl sends an acme control message to the given window",
	Args: cobra.MinimumNArgs(2),
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
	f string
}

func (w writer) Write(p []byte) (n int, err error) {
	return w.win.Write(w.f, p)
}

var writeCmd = &cobra.Command{
	Use: "write <winid> <winfile>",
	Short: 	"write copies stdin to the window's winfile",
	Args: cobra.ExactArgs(2),
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
			f: args[1],
		}
		_, err = io.Copy(dst, os.Stdin)
		if err != nil {
			return fmt.Errorf("copy data: %w", err)
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
	rootCmd.AddCommand(newCmd, ctlCmd, writeCmd)
	return rootCmd.Execute()
}
