package command

import (
    "bytes"
    "fmt"
    "os/exec"
    "syscall"
    "time"
    "os"
    "gopkg.in/bufio.v1"
)

// 默认超时时间/秒
const DEFAULT_TIMEOUT = time.Second * 3600

type ErrExecTimeout struct {
    Duration time.Duration
}

func IsErrExecTimeout(err error) bool {
    _, ok := err.(ErrExecTimeout)
    return ok
}

func (err ErrExecTimeout) Error() string {
    return fmt.Sprintf("execution is timeout [duration: %v]", err.Duration)
}

type Command struct {
    name         string
    args         []string
    stdout       *bytes.Buffer
    stderr       *bytes.Buffer
    Pid          int
    ProcessState *os.ProcessState
}

func NewCommand(cmd string) *Command {
    args := []string{"-c", cmd}
    return &Command{
        name:   "/bin/sh",
        args:   args,
        stdout: new(bytes.Buffer),
        stderr: new(bytes.Buffer),
    }
}

func (c *Command) Run() error {
    return c.RunInDirTimeout("", 0)
}

func (c *Command) RunTimeout(timeout time.Duration) error {
    return c.RunInDirTimeout("", timeout)
}

func (c *Command) RunInDir(dir string) error {
    return c.RunInDirTimeout(dir, 0)
}

func (c *Command) Stdout() []byte {
    return c.stdout.Bytes()
}

func (c *Command) Stderr() []byte {
    return c.stderr.Bytes()
}

func (c *Command) RunInDirTimeout(dir string, timeout time.Duration) error {
    var err error
    cmd := exec.Command(c.name, c.args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // 设置进程组
    cmd.Dir = dir
    cmd.Stdout = c.stdout
    cmd.Stderr = c.stderr

    if (timeout == -1) {
        timeout = DEFAULT_TIMEOUT
    }
    if err = cmd.Start(); err != nil {
        return err
    }
    defer func() {
        c.Pid = cmd.Process.Pid
        c.ProcessState = cmd.ProcessState
    }()
    if timeout == 0 {
        err = cmd.Wait()
        return c.concatenateError(err)
    } else {
        done := make(chan error)
        go func() {
            done <- cmd.Wait()
        }()
        var err error
        select {
        case <-time.After(timeout):
            if cmd.Process != nil && cmd.ProcessState == nil {
                // 使用cmd.Process.Kill()无法杀掉子进程，改用 syscall.Kill() 杀掉整个进程组
                if err = syscall.Kill(0 - cmd.Process.Pid, syscall.SIGKILL); err != nil {
                    return fmt.Errorf("fail to kill process: %v", err)
                }
            }
            <-done
            return ErrExecTimeout{timeout}
        case err = <-done:
            return c.concatenateError(err)
        }
    }
}

func (c *Command) Dump() string {
    var buf bufio.Buffer
    buf.WriteString(fmt.Sprintf("cmd: %s %s '%s'\n", c.name, c.args[0], c.args[1]))
    buf.WriteString(fmt.Sprintf("pid: %d\n", c.Pid))
    buf.WriteString(fmt.Sprintf("exit: %s\n", c.ProcessState.String()))
    buf.WriteString(fmt.Sprintf("stdout: %s\n", c.stdout.String()))
    buf.WriteString(fmt.Sprintf("stderr: %s\n", c.stderr.String()))
    return buf.String()
}

func (c *Command) concatenateError(err error) error {
    if err == nil || c.stderr.Len() == 0 {
        return err
    }
    return fmt.Errorf("%v - %s", err, c.stderr.String())
}