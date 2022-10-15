package sandbox

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/vincent-vinf/code-validator/pkg/util/log"
)

const (
	MaxID = 999

	defaultFileSize             = 1024 * 1204 // 1GB
	defaultFileMode fs.FileMode = 0660
)

var (
	logger = log.GetLogger()
)

type Sandbox interface {
	Init() error
	Run(cmd string, args []string, opts ...Option) error
	Clean() error

	WriteFile(filepath string, data []byte) error
	ReadFile(filepath string) ([]byte, error)

	GetID() int
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
	if err := exec.Command("isolate", "--version").Run(); err != nil {
		return fmt.Errorf("is isolate installed? err: %w", err)
	}

	if data, err := exec.Command("isolate", fmt.Sprintf("-b %d", i.id), "--init").Output(); err != nil {
		return fmt.Errorf("init box(%d) err: %w", i.id, err)
	} else {
		i.workdir = strings.TrimSpace(string(data))
	}
	logger.Debug("workdir: ", i.workdir)

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
		timeLimit:      time.Second * 10,
		wallTimeLimit:  time.Second * 20,
		extraTimeLimit: time.Second * 20,
		// unlimited processes
		processes: 0,
	}
	for _, opt := range opts {
		opt(r)
	}
	gArgs := r.getArgs()
	gArgs = append(gArgs, fmt.Sprintf("-b %d", i.id), "-s", "--dir=/etc=/etc:noexec", "--run", "--", cmd)
	gArgs = append(gArgs, args...)
	logger.Debug(gArgs)

	c := exec.Command("isolate", gArgs...)
	c.Stdin = r.stdin
	c.Stdout = r.stdout
	c.Stderr = r.stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("run cmd(%s) in box(%d) err: %w", cmd, i.id, err)
	}

	return nil
}

func (i *Isolate) GetID() int {
	return i.id
}

func (i *Isolate) WriteFile(filepath string, data []byte) error {
	if err := i.initFile(filepath); err != nil {
		return err
	}

	return ioutil.WriteFile(i.pathConvert(filepath), data, defaultFileMode)
}
func (i *Isolate) ReadFile(filepath string) ([]byte, error) {
	return ioutil.ReadFile(i.pathConvert(filepath))
}
func (i *Isolate) pathConvert(filepath string) string {
	if path.IsAbs(filepath) {
		s := strings.Split(filepath, "/")
		switch s[1] {
		case "tmp":
			return path.Join(i.workdir, "tmp", filepath)
		default:
			return path.Join(i.workdir, filepath)
		}
	}

	return path.Join(i.workdir, "box", filepath)
}
func (i *Isolate) initFile(filepath string) error {
	filepath = path.Clean(filepath)
	var outBuf, errBuf bytes.Buffer

	err := i.Run("/bin/sh",
		[]string{"-c", fmt.Sprintf("mkdir -p %s && touch %s", path.Dir(filepath), filepath)},
		Env(map[string]string{
			"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		}),
		Stdout(&outBuf),
		Stderr(&errBuf),
	)
	if err != nil {
		return fmt.Errorf("init file err, stdout: %s, stderr: %s, %w", outBuf.String(), errBuf.String(), err)
	}

	return nil
}

type run struct {
	metadata string
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

func (r *run) getArgs() (args []string) {
	if r.metadata != "" {
		args = append(args, "--meta="+r.metadata)
	}
	if r.enableNetwork {
		args = append(args, "--share-net")
	}
	if r.fileSize > 0 {
		args = append(args, fmt.Sprintf("--fsize=%d", r.fileSize))
	}

	args = append(args,
		fmt.Sprintf("--time=%2f", r.timeLimit.Seconds()),
		fmt.Sprintf("--extra-time=%2f", r.extraTimeLimit.Seconds()),
		fmt.Sprintf("--wall-time=%2f", r.wallTimeLimit.Seconds()),
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

func Metadata(file string) Option {
	return func(r *run) {
		r.metadata = file
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
