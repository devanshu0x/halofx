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
}

func RenderMac(options MacOptions) error {
	args := []string{}
	if options.Force {
		args = append(args, "-y")
	} else {
		args = append(args, "-n")
	}

	filter := fmt.Sprintf(
		"[0:v]scale=%d:%d[bg];"+
			"[1:v]scale=%d:%d,format=rgba[win];"+
			"[2:v]format=ya8,alphaextract,format=gray[mask];"+
			"[win][mask]alphamerge[rounded];"+
			"[bg][rounded]overlay=(W-w)/2:(H-h)/2",
		options.Width,
		options.Height,
		options.VideoWidth,
		options.VideoHeight,
	)

	args = append(args,
		"-loop", "1", "-i", options.BackgroundPath,
		"-i", options.InputPath,
		"-i", options.MaskPath,
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
