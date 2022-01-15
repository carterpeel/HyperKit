package dockerutil

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hyperkit/core/util"
	"os/exec"
	"path/filepath"
)

var (
	dockerPath string
)

func init() {
	var err error
	if dockerPath, err = CheckDocker(); err != nil {
		log.Fatalf("could not find docker: %v\n", err)
	}
}

func CheckDocker() (path string, err error) {
	if path, err = exec.LookPath("docker"); err != nil {
		return "", exec.ErrNotFound
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("error converting Docker path %q to absolute: %w", path, err)
	}
	return absPath, nil
}

func PauseContainer() error {
	return util.FormatExecOutputToError(exec.Command(dockerPath, LedFxContainerPauseArgs...).CombinedOutput())
}
func ResumeContainer() error {
	return util.FormatExecOutputToError(exec.Command(dockerPath, LedFxContainerUnpauseArgs...).CombinedOutput())
}

func StartContainer() error {
	return util.FormatExecOutputToError(exec.Command(dockerPath, LedFxContainerStartArgs...).CombinedOutput())
}

func RestartContainer() error {
	return util.FormatExecOutputToError(exec.Command(dockerPath, LedFxContainerRestartArgs...).CombinedOutput())
}

func CreateContainerIfNotExist() error {
	if !ContainerExists() {
		return util.FormatExecOutputToError(exec.Command(dockerPath, LedFxContainerCreateArgs...).CombinedOutput())
	}
	return nil
}

func ContainerExists() bool {
	cmd := exec.Command(dockerPath, LedFxContainerInspectArgs...)
	_ = cmd.Start()
	_ = cmd.Wait()
	return cmd.ProcessState.ExitCode() == 0
}

func ContainerState() State {
	out, err := exec.Command(dockerPath, LedFxContainerGetStateArgs...).Output()
	if err != nil {
		log.Errorf("Error getting container state: %v\n", err)
		return StateUnknown
	}
	out = bytes.ReplaceAll(out, []byte("\n"), nil)

	log.Infof("LedFX Container state: %s\n", string(out))
	switch string(out) { //nolint:typecheck
	case "'paused'":
		return StatePaused
	case "'running'":
		return StateOnline
	case "'exited'":
		return StateOffline
	}
	return StateUnknown
}
