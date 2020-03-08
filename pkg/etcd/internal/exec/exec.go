package exec

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

// Executes command and outputs stdout+stderr and error if any.
// Arguments:
// - cmd  - command to execute
// - args - arguments for command
// - envs - envronment variables
func Cmd(cmd string, args []string, envs []string, logger micrologger.Logger) ([]byte, error) {
	logger.Log("level", "info", "msg", fmt.Sprintf("Executing: %s %v", cmd, args))

	// Create cmd and add environment.
	c := exec.Command(cmd, args...)
	c.Env = append(os.Environ(), envs...)

	// Execute and get output.
	stdOutErr, err := c.CombinedOutput()
	if err != nil {
		logger.Log("level", "error", "msg", "execCmd failed", "reason", fmt.Sprintf("%s", stdOutErr), "err", err)
		return stdOutErr, microerror.Mask(err)
	}
	return stdOutErr, nil
}
