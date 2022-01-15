package ledfx

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"hyperkit/core/ledfx/dockerutil" //nolint:typecheck
	"strings"
	"time"
)

type Controller struct {
	dockerPath string
}

func NewController() (ctl *Controller, err error) {
	if err := dockerutil.CreateContainerIfNotExist(); err != nil {
		return nil, fmt.Errorf("error creating container if nonexistent: %w", err)
	}

	switch dockerutil.ContainerState() {
	case dockerutil.StatePaused:
		return ctl, nil
	case dockerutil.StateOffline:
		if err := dockerutil.StartContainer(); err != nil {
			return nil, fmt.Errorf("error starting container: %w", err)
		}
		defer func() {
			time.Sleep(3 * time.Second)
			if err := ctl.Pause(); err != nil {
				log.Errorf("Error pausing container: %v\n", err)
			}
		}()
	case dockerutil.StateOnline:
		if err := ctl.Pause(); err != nil {
			return nil, fmt.Errorf("error pausing container: %w", err)
		}
	case dockerutil.StateUnknown:
		return nil, fmt.Errorf("unknown container state")
	}

	return ctl, nil
}

func (ctl *Controller) Pause() error {
	if err := dockerutil.PauseContainer(); err != nil {
		if strings.Contains(err.Error(), "is already paused") {
			return nil
		}
		return fmt.Errorf("error pausing container: %w", err)
	}
	return nil
}

func (ctl *Controller) Resume() error {
	if err := dockerutil.ResumeContainer(); err != nil {
		if strings.Contains(err.Error(), "is not paused") {
			return nil
		}
		return fmt.Errorf("error resuming container: %w", err)
	}
	return nil
}
