package ffmpeg

import (
	"os"
	"os/exec"
)

func Run(args ...string) error {
	// implementation goes here
	cmd:=exec.Command("ffmpeg", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}