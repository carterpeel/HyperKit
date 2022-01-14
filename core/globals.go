package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/gorilla/websocket"
	syslogger "log"
)

var (
	socket      *websocket.Conn
	bridge      *accessory.Bridge
	hyperCube   *accessory.Outlet
	miscHandler *MiscHandler
	config      *Config
	log         *syslogger.Logger
)
