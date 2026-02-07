package render

import (
	"bytes"
	"fmt"
	"github.com/devanshu0x/halofx/internal/ffmpeg"
	"os/exec"
	"strconv"
	"strings"
)

type MacOptions struct {
	InputPath      string
	OutputPath     string
	BackgroundPath string
	MaskPath       string
	Width          int
	Height         int
	VideoHeight    int
	VideoWidth     int
	Force          bool
	FramePath      string
}

func RenderMac(options MacOptions) error {
	args := []string{}
	if options.Force {
		args = append(args, "-y")
	} else {
		args = append(args, "-n")
	}

	filter := fmt.Sprintf(
		// background
		"[0:v]scale=%d:%d,format=rgba[bg];"+

			// video
			"[1:v]scale=%d:%d,format=rgba[win];"+

			// anti-aliased mask
			"[2:v]scale=%d:%d,format=rgba,alphaextract,"+
			"gblur=sigma=0.6:steps=1[mask];"+

			// apply rounded mask
			"[win][mask]alphamerge[rounded];"+

			//Frame
			"[3:v]format=rgba[frame];"+

			// composite
			"[bg][frame]overlay=(W-w)/2:(H-h)/2[bgframe];"+
			"[bgframe][rounded]overlay=(W-w)/2:(H-h)/2",
		options.Width,
		options.Height,
		options.VideoWidth,
		options.VideoHeight,
		options.VideoWidth,
		options.VideoHeight,
	)

	args = append(args,
		"-loop", "1", "-i", options.BackgroundPath,
		"-i", options.InputPath,
		"-i", options.MaskPath,
		"-i", options.FramePath,
		"-filter_complex",
		filter,
		"-shortest",
		options.OutputPath,
	)

	return ffmpeg.Run(args...)

}

func GetVideoDimensions(path string) (width, height int, err error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		path,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Split(strings.TrimSpace(out.String()), "x")
	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return w, h, nil
}
