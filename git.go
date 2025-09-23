package main

import "os/exec"

func git(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
