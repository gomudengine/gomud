package main

import (
	"os"

	"github.com/GoMudEngine/GoMud/internal/copyover"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func triggerCopyover() error {
	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	serverAlive.Store(false)

	rooms.SaveAllRooms()
	users.SaveAllUsers()
	plugins.Save()

	return copyover.Execute(binaryPath, os.Args[1:])
}
