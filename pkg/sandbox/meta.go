package sandbox

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Meta struct {
	// the default value -1 means the value is not set
	CSWForced    int
	CSWVoluntary int
	Exitcode     int
	MaxRSS       int
	Time         int
	TimeWall     int

	ExitSig int
	Killed  bool
	Message string
	//RE: run-time error, i.e., exited with a non-zero exit code
	//SG: program died on a signal
	//TO: timed out
	//XX: internal error of the sandbox
	Status string
}

func NewMeta() *Meta {
	return &Meta{
		CSWForced:    -1,
		CSWVoluntary: -1,
		Exitcode:     -1,
		MaxRSS:       -1,
		Time:         -1,
		TimeWall:     -1,
		ExitSig:      -1,
	}
}

func (m *Meta) ReadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("read meta file: %s, err: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		s := strings.Split(line, ":")
		if len(s) != 2 {
			continue
		}
		k := strings.TrimSpace(s[0])
		v := strings.TrimSpace(s[1])
		switch k {
		case "csw-forced":
			m.CSWForced = atoi(v)
		case "csw-voluntary":
		case "exitcode":
		case "max-rss":
		case "time":
		case "time-wall":

		case "exitsig":
		case "killed":
		case "message":
		case "status":

		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("cannot read meta file: %s, err: [%v]", path, err)
	}

	return nil
}

func atoi(str string) int {
	n, err := strconv.Atoi(str)
	if err != nil {
		return -1
	}

	return n
}
