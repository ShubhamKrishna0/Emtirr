package main

import (
	"os"
	"os/exec"
)

func main() {
	// Change to backend-go directory and run the actual main
	os.Chdir("backend-go")
	cmd := exec.Command("./main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}