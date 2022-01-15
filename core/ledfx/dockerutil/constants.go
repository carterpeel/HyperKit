package dockerutil

var (
	LedFxContainerComposeFile = `version: '3.3'
services:
    ledfx:
        network_mode: host
        volumes:
            - '/home/pi/ledfx/audio/:/app/audio'
            - '/home/pi/ledfx/ledfx-config/:/app/ledfx-config'
        restart: always
        logging:
            options:
                max-size: 64m
        container_name: ledfx
        image: 'cpeel147/ledfx:latest'`

	LedFxContainerCreateArgs = []string{
		"run",
		"-d",
		"--network=host",
		"--volume=/home/pi/ledfx/audio/:/app/audio/",
		"--volume=/home/pi/ledfx/ledfx-config/:/app/ledfx-config/",
		"--restart=always",
		"--log-opt=max-size=64m",
		"--name=ledfx",
		"cpeel147/ledfx:latest",
	}

	LedFxContainerPauseArgs = []string{
		"container",
		"pause",
		"ledfx",
	}

	LedFxContainerUnpauseArgs = []string{
		"container",
		"unpause",
		"ledfx",
	}

	LedFxContainerRestartArgs = []string{
		"container",
		"restart",
		"ledfx",
	}

	LedFxContainerStopArgs = []string{
		"container",
		"stop",
		"ledfx",
	}

	LedFxContainerStartArgs = []string{
		"container",
		"start",
		"ledfx",
	}

	LedFxContainerGetStateArgs = []string{
		"container",
		"inspect",
		"--format='{{.State.Status}}'",
		"ledfx",
	}

	LedFxContainerInspectArgs = []string{
		"container",
		"inspect",
		"ledfx",
	}
)

type State uint8

const (
	StateOnline State = iota
	StateOffline
	StatePaused
	StateUnknown
)
