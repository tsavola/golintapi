// Copyright (c) 2019 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const golint = "golint"

var lineRegexp = regexp.MustCompile(`^(.*):[0-9]+:[0-9]+:\s+(.*)`)

var ignoredMessages = []*regexp.Regexp{
	regexp.MustCompile(`^comment on `),
	regexp.MustCompile(`^const [a-z_]`),
	regexp.MustCompile(`^exported [a-z]+ [A-Za-z0-9_.]+ should have comment `),
	regexp.MustCompile(`^if block ends with a return statement, so drop this else and outdent its block`),
	regexp.MustCompile(`^should replace [A-Za-z0-9_.]+ [+-]= `),
	regexp.MustCompile(`^package comment should be of the form `),
	regexp.MustCompile(`^receiver name [A-Za-z0-9_]+ should be consistent with previous receiver name `),
}

func main() {
	cmd := exec.Command(golint, os.Args[1:]...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	r := bufio.NewReader(stdout)

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
			os.Exit(1)
		}

		fields := lineRegexp.FindSubmatch(bytes.TrimSpace(line))
		filename := string(fields[1])
		message := fields[2]

		if strings.HasPrefix(filename, "internal/") || strings.Contains(filename, "/internal/") {
			continue
		}

		var ignored bool

		for _, re := range ignoredMessages {
			if re.Match(message) {
				ignored = true
				break
			}
		}

		if ignored {
			continue
		}

		os.Stdout.Write(line)
	}

	if err := cmd.Wait(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.Exited() {
			os.Exit(exit.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "%s: %v\n", golint, err)
			os.Exit(1)
		}
	}
}
