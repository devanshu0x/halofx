package render

import (
	"fmt"

	"github.com/devanshu0x/halofx/internal/ffmpeg"
)

type MacOptions struct {
	InputPath  string
	OutputPath string
	BackgroundPath string
	Width  int
	Height int
	Force bool
}

func RenderMac(options MacOptions) error {
	args := []string{}
	if options.Force {
		args = append(args, "-y")
	}else {
		args = append(args, "-n")
	}

	filter:= fmt.Sprintf(
		"[0:v]scale=%d:%d[bg];"+
			"[1:v]scale=1600:900:force_original_aspect_ratio=decrease[win];"+
			"[bg][win]overlay=(W-w)/2:(H-h)/2",
		options.Width,
		options.Height,
	)


	args = append(args,
		"-loop", "1", "-i", options.BackgroundPath,
		"-i", options.InputPath,
		"-filter_complex",
		// filter string goes here
		filter,
		"-shortest",
		options.OutputPath,
	)

	return ffmpeg.Run(args...)
	
}