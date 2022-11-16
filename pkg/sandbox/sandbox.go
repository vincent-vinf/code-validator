package sandbox

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	MaxID = 999

	defaultFileSize = 1024 * 1204 // 1GB
)

type Sandbox interface {
	GetID() int

	Init() error
	Run(cmd string, args []string, opts ...Option) error
	Clean() error

	WriteFile(filepath string, data []byte) error
	ReadFile(filepath string) ([]byte, error)
	RemoveFile(paths ...string) error
}

func New(id int) (Sandbox, error) {
	return NewIsolate(id)
}

func NewIsolate(id int) (*Isolate, error) {
	if id > MaxID {
		return nil, fmt.Errorf("id(%d) out of range (allowed: 0-%d)", id, MaxID)
	}
	i := &Isolate{
		id: id,
	}

	return i, nil
}

type Isolate struct {
	id int

	workdir string
}

func (i *Isolate) Init() error {
	if data, err := exec.Command("isolate", fmt.Sprintf("-b %d", i.id), "--init").Output(); err != nil {
		return fmt.Errorf("init box(%d) err: %w", i.id, err)
	} else {
		i.workdir = strings.TrimSpace(string(data))
	}
	fmt.Println("workdir: ", i.workdir)

	return nil
}
func (i *Isolate) Clean() error {
	if err := exec.Command("isolate", fmt.Sprintf("-b %d", i.id), "--cleanup").Run(); err != nil {
		return fmt.Errorf("clean up box(%d) err: %w", i.id, err)
	}

	return nil
}
func (i *Isolate) Run(cmd string, args []string, opts ...Option) error {
	r := &run{
		// share network
		enableNetwork:  true,
		timeLimit:      time.Minute,
		wallTimeLimit:  time.Minute * 10,
		extraTimeLimit: time.Minute * 2,
		// unlimited processes
		processes: 0,
	}
	for _, opt := range opts {
		opt(r)
	}
	gArgs := r.getArgs(i.workdir)
	gArgs = append(gArgs, fmt.Sprintf("-b %d", i.id), "-s", "--dir=/etc=/etc:noexec", "--run", "--", cmd)
	gArgs = append(gArgs, args...)
	fmt.Println(cmd, " ", strings.Join(args, " "))

	c := exec.Command("isolate", gArgs...)
	c.Stdin = r.stdin
	c.Stdout = r.stdout
	c.Stderr = r.stderr
	err := c.Run()

	if r.meta != nil {
		_ = r.meta.ReadFile(path.Join(i.workdir, "meta"))
	}
	if err != nil {
		return fmt.Errorf("run cmd(%s) args(%s) in box(%d) err: %w", cmd, strings.Join(args, ","), i.id, err)
	}

	return nil
}

func (i *Isolate) GetID() int {
	return i.id
}

func (i *Isolate) WriteFile(filepath string, data []byte) error {
	filepath = path.Clean(filepath)
	stdout, stderr, err := i.runSh(fmt.Sprintf("mkdir -p %s && cat - > %s", path.Dir(filepath), filepath), bytes.NewReader(data))

	if err != nil {
		return fmt.Errorf("write file err: %w, stdout: %s, stderr: %s", err, stdout, stderr)
	}

	return nil
}
func (i *Isolate) ReadFile(filepath string) ([]byte, error) {
	var outBuf, errBuf bytes.Buffer

	err := i.Run("/bin/cat",
		[]string{filepath},
		Stdout(&outBuf),
		Stderr(&errBuf),
	)

	if err != nil {
		return nil, fmt.Errorf("read file err: %w, stdout: %s, stderr: %s", err, outBuf.String(), errBuf.String())
	}

	return outBuf.Bytes(), nil
}
func (i *Isolate) RemoveFile(paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	var errBuf bytes.Buffer

	err := i.Run("/bin/rm",
		append(paths, "-rf"),
		Stderr(&errBuf),
	)

	if err != nil {
		return fmt.Errorf("rm file err, stderr: %s, %w", errBuf.String(), err)
	}

	return nil
}

func (i *Isolate) runSh(sh string, stdin io.Reader) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer

	err = i.Run("/bin/sh",
		[]string{"-c", sh},
		Env(map[string]string{
			"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			"HOME": "/tmp",
		}),
		Stdout(&outBuf),
		Stderr(&errBuf),
		Stdin(stdin),
	)

	return outBuf.String(), errBuf.String(), err
}

func (i *Isolate) Workdir() string {
	return i.workdir
}

type run struct {
	meta *Meta
	// --share-net
	enableNetwork bool

	timeLimit      time.Duration
	wallTimeLimit  time.Duration
	extraTimeLimit time.Duration

	processes int
	fileSize  int

	env map[string]string

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (r *run) getArgs(workdir string) (args []string) {
	if r.meta != nil {
		args = append(args, fmt.Sprintf("--meta=%s", path.Join(workdir, "meta")))
	}
	if r.enableNetwork {
		args = append(args, "--share-net")
	}
	if r.fileSize > 0 {
		args = append(args, fmt.Sprintf("--fsize=%d", r.fileSize))
	}

	args = append(args,
		fmt.Sprintf("--time=%.2f", r.timeLimit.Seconds()),
		fmt.Sprintf("--extra-time=%.2f", r.extraTimeLimit.Seconds()),
		fmt.Sprintf("--wall-time=%.2f", r.wallTimeLimit.Seconds()),
	)

	if r.processes <= 0 {
		args = append(args, "-p")
	} else {
		args = append(args, fmt.Sprintf("-p%d", r.processes))
	}
	for k, v := range r.env {
		if v == "" {
			args = append(args, fmt.Sprintf("--env=%s", k))
		} else {
			args = append(args, fmt.Sprintf("--env=%s=%s", k, v))
		}
	}

	return
}

type Option func(*run)

func Metadata(meta *Meta) Option {
	return func(r *run) {
		r.meta = meta
	}
}
func Network(b bool) Option {
	return func(r *run) {
		r.enableNetwork = b
	}
}
func Time(t time.Duration) Option {
	return func(r *run) {
		r.timeLimit = t
	}
}
func ExtraTime(t time.Duration) Option {
	return func(r *run) {
		r.extraTimeLimit = t
	}
}
func WallTime(t time.Duration) Option {
	return func(r *run) {
		r.wallTimeLimit = t
	}
}
func Processes(num int) Option {
	if num < 0 {
		num = 0
	}
	return func(r *run) {
		r.processes = num
	}
}
func FileSize(kb int) Option {
	if kb < 0 {
		kb = defaultFileSize
	}
	return func(r *run) {
		r.fileSize = kb
	}
}
func Env(kv map[string]string) Option {
	return func(r *run) {
		r.env = kv
	}
}
func Stdin(i io.Reader) Option {
	return func(r *run) {
		r.stdin = i
	}
}
func Stdout(o io.Writer) Option {
	return func(r *run) {
		r.stdout = o
	}
}
func Stderr(e io.Writer) Option {
	return func(r *run) {
		r.stderr = e
	}
}
