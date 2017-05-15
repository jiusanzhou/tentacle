package util

import (
	"strings"
	"os/exec"
	"errors"
)

func DoCommand(cmdstr string) ([]byte, error) {
	keys := strings.Split(cmdstr, " ")
	if len(keys) < 1 {
		return nil, errors.New("No command offered.")
	}
	cmd := keys[0]
	args := keys[1:]
	c := exec.Command(cmd, args...)
	err := c.Run()
	o, _ := c.Output()
	return o, err
}
