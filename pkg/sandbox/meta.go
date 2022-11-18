package sandbox

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Meta struct {
	// int: the default value -1 means the value is not set
	CSWForced    int `json:"CSWForced"`
	CSWVoluntary int `json:"CSWVoluntary"`
	ExitCode     int `json:"exitCode"`
	MaxRSS       int `json:"maxRSS"`

	Time     float64 `json:"time"`
	TimeWall float64 `json:"timeWall"`

	ExitSig int    `json:"exitSig"`
	Killed  bool   `json:"killed"`
	Message string `json:"message"`
	//RE: run-time error, i.e., exited with a non-zero exit code
	//SG: program died on a signal
	//TO: timed out
	//XX: internal error of the sandbox
	Status string `json:"status"`
}

func NewMeta() *Meta {
	return &Meta{
		CSWForced:    -1,
		CSWVoluntary: -1,
		ExitCode:     -1,
		MaxRSS:       -1,
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
			m.CSWVoluntary = atoi(v)
		case "exitcode":
			m.ExitCode = atoi(v)
		case "max-rss":
			m.MaxRSS = atoi(v)
		case "time":
			float, err := strconv.ParseFloat(v, 64)
			if err != nil {
				float = 0
			}
			m.Time = float
		case "time-wall":
			float, err := strconv.ParseFloat(v, 64)
			if err != nil {
				float = 0
			}
			m.TimeWall = float
		case "exitsig":
			m.ExitSig = atoi(v)
		case "killed":
			n := atoi(v)
			m.Killed = n == 1
		case "message":
			m.Message = v
		case "status":
			m.Status = v
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
