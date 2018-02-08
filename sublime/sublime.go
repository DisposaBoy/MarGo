package sublime

import (
	"disposa.blue/margo/mgcli"
	"fmt"
	"github.com/urfave/cli"
	"go/build"
	"os"
	"os/exec"
	"strings"
)

var (
	Command = cli.Command{
		Name:            "sublime",
		Aliases:         []string{"subl"},
		Usage:           "",
		Description:     "",
		SkipFlagParsing: true,
		SkipArgReorder:  true,
		Action:          mgcli.Action(mainAction),
	}
)

type cmdHelper struct {
	name     string
	args     []string
	outToErr bool
	env      []string
}

func (c cmdHelper) run() error {
	cmd := exec.Command(c.name, c.args...)
	cmd.Stdin = os.Stdin
	if c.outToErr {
		cmd.Stdout = os.Stderr
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	cmd.Env = c.env

	fmt.Fprintf(os.Stderr, "run%q\n", append([]string{c.name}, c.args...))
	return cmd.Run()
}

func mainAction(c *cli.Context) error {
	args := c.Args()
	tags := []string{"margo"}
	if extensionPkgExists() {
		tags = []string{"margo margo_extension", "margo"}
	}
	if err := goInstallAgent(os.Getenv("MARGO_SUBLIME_GOPATH"), tags); err != nil {
		return fmt.Errorf("cannot install margo.sublime: %s", err)
	}
	name := "margo.sublime"
	if exe, err := exec.LookPath(name); err == nil {
		name = exe
	}
	return cmdHelper{name: name, args: args}.run()
}

func goInstallAgent(gp string, tags []string) error {
	var env []string
	if gp != "" {
		env = make([]string, 0, len(os.Environ())+1)
		// I don't remember the rules about duplicate env vars...
		for _, s := range os.Environ() {
			if !strings.HasPrefix(s, "GOPATH=") {
				env = append(env, s)
			}
		}
		env = append(env, "GOPATH="+gp)
	}

	cmdpath := "disposa.blue/margo/cmd/margo.sublime"
	if s := os.Getenv("MARGO_SUBLIME_CMDPATH"); s != "" {
		cmdpath = s
	}

	race := os.Getenv("MARGO_INSTALL_FLAGS_RACE") == "1"
	var err error
	for _, tag := range tags {
		args := []string{"install", "-v", "-tags", tag, cmdpath}
		if race {
			args = append([]string{"install", "-race"}, args[1:]...)
		}
		err = cmdHelper{
			name:     "go",
			args:     args,
			outToErr: true,
			env:      env,
		}.run()
		if err == nil {
			return nil
		}
	}
	return err
}

func extensionPkgExists() bool {
	_, err := build.Import("margo", "", 0)
	return err == nil
}