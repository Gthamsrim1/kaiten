package runtime

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Runtime struct{}

const (
	childMode = "__kaiten_child__"
	hostname  = "kaiten"
)

func New() *Runtime {
	return &Runtime{}
}

func (r *Runtime) Run(args []string) error {
	if len(args) == 0 {
		return errors.New("missing command")
	}

	if args[0] == childMode {
		return r.child(args[1:])
	}

	if args[0] != "run" {
		return fmt.Errorf("unknown command %q", args[0])
	}

	return r.parent(args[1:])
}

func (r *Runtime) parent(args []string) error {
	cmd := exec.Command("/proc/self/exe", append([]string{childMode}, args...)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS,
	}

	return cmd.Run()
}

func (r *Runtime) child(args []string) error {
	if len(args) == 0 {
		return errors.New("missing command")
	}

	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return err
	}

	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
