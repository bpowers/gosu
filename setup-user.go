package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/docker/libcontainer/system"
	"github.com/docker/libcontainer/user"
)

// this function comes from libcontainer/init_linux.go
// we don't use that directly because we don't want the whole namespaces package imported here

// SetupUser changes the groups, gid, and uid for the user inside the container
func SetupUser(u string) error {
	// Set up defaults.
	defaultExecUser := user.ExecUser{
		Uid:  syscall.Getuid(),
		Gid:  syscall.Getgid(),
		Home: "/",
	}
	passwdPath, err := user.GetPasswdPath()
	if err != nil {
		return err
	}
	groupPath, err := user.GetGroupPath()
	if err != nil {
		return err
	}
	execUser, err := user.GetExecUserPath(u, &defaultExecUser, passwdPath, groupPath)
	if err != nil {
		return fmt.Errorf("get supplementary groups %s", err)
	}
	if err := syscall.Setgroups(execUser.Sgids); err != nil {
		return fmt.Errorf("setgroups %s", err)
	}
	if err := system.Setgid(execUser.Gid); err != nil {
		return fmt.Errorf("setgid %s", err)
	}
	if err := system.Setuid(execUser.Uid); err != nil {
		return fmt.Errorf("setuid %s", err)
	}
	// if we didn't get HOME already, set it based on the user's HOME
	if envHome := os.Getenv("HOME"); envHome == "" {
		if err := os.Setenv("HOME", execUser.Home); err != nil {
			return fmt.Errorf("set HOME %s", err)
		}
	}
	return nil
}
