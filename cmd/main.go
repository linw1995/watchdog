package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
)

func readPrintln(pipe func() (io.ReadCloser, error)) error {
	reader, error := pipe()
	if error != nil {
		return error
	}
	lineReader := bufio.NewReader(reader)
	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			return err
		}
		if len(line) == 0 {
			return nil
		}
		fmt.Printf("%s", line)
	}
}
func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) <= 0 {
		panic("need a command to run")
	}

	name := args[0]
	cmd := exec.Command(name, args[1:]...)

	var g errgroup.Group
	g.Go(func() error {
		g.Go(func() error {
			return readPrintln(cmd.StderrPipe)
		})
		g.Go(func() error {
			return readPrintln(cmd.StdoutPipe)
		})
		err := cmd.Run()
		fmt.Print("\n")
		return err
	})
	if err := g.Wait(); err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
