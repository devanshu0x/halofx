package main

import (
	"flag"
	"fmt"
	"image/color"

	// "image/color"
	"os"
	"path/filepath"
	"slices"

	"github.com/devanshu0x/halofx/internal/mask"
	"github.com/devanshu0x/halofx/internal/render"
	"github.com/devanshu0x/halofx/internal/ui"
)

var version = "dev"

var videoExtensions = []string{".mp4", ".mkv", ".avi", ".mov", ".flv", ".wmv", ".mpeg", ".mpg", ".webm"}

const (
	DefaultWidth  = 1920
	DefaultHeight = 1080
	DefaultRadius = 16
	DefaultBG     = "internal/assets/backgrounds/bg-1.jpg"
	DefaultFrameThickness = 12
)

type Config struct {
	InputPath   string
	OutputPath  string
	PaddingX    int
	PaddingY    int
	Force       bool
	ShowVersion bool
}



func main() {
	cfg := parseFlags()

	if cfg.ShowVersion {
		ui.Info("halofx version: " + version)
		return
	}

	if err := run(cfg); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func run(cfg Config) error {
	if cfg.InputPath == "" {
		return fmt.Errorf("input file is required, specify with -i flag")
	}

	if err := videoFileExists(cfg.InputPath); err != nil {
		return fmt.Errorf("error with input file: %w", err)
	}

	outputPath, err := resolveOutputPath(cfg.InputPath, cfg.OutputPath, cfg.Force)
	if err != nil {
		return err
	}

	videoWidth, videoHeight, err := render.GetVideoDimensions(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("failed to get video dimensions: %w", err)
	}
	if videoWidth == 0 || videoHeight == 0 {
		return fmt.Errorf("invalid video dimensions")
	}

	adjustedWidth, adjustedHeight := fitInside(
		videoWidth,
		videoHeight,
		DefaultWidth-2*cfg.PaddingX,
		DefaultHeight-2*cfg.PaddingY,
	)

	tmpFile, err := os.CreateTemp("", "halofx-mask-*.png")
	if err != nil {
		return fmt.Errorf("failed to create temporary mask file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := mask.GenerateRoundedMask(
		tmpFile.Name(),
		adjustedWidth,
		adjustedHeight,
		DefaultRadius,
		color.NRGBA{0, 0, 0, 255},
	); err != nil {
		return fmt.Errorf("failed to generate mask: %w", err)
	}

	frameFile, err := os.CreateTemp("", "halofx-frame-*.png")
	if err != nil {
		return fmt.Errorf("failed to create temporary frame file: %w", err)
	}
	defer os.Remove(frameFile.Name())

	if err := mask.GenerateRoundedMask(
		frameFile.Name(),
		adjustedWidth+2*DefaultFrameThickness,
		adjustedHeight+2*DefaultFrameThickness,
		DefaultRadius,
		color.NRGBA{255, 255, 255, 40},
	); err != nil {
		return fmt.Errorf("failed to generate frame: %w", err)
	}

	if err := render.RenderMac(render.MacOptions{
		InputPath:      cfg.InputPath,
		OutputPath:     outputPath,
		BackgroundPath: DefaultBG,
		MaskPath:       tmpFile.Name(),
		Width:          DefaultWidth,
		Height:         DefaultHeight,
		VideoWidth:     adjustedWidth,
		VideoHeight:    adjustedHeight,
		Force:          cfg.Force,
		FramePath: 	frameFile.Name(),
	}); err != nil {
		return fmt.Errorf("render failed: %w", err)
	}

	ui.Success("Output written to " + outputPath)
	return nil
}

func parseFlags() Config {
	input := flag.String("i", "", "input file")
	output := flag.String("o", "", "output file")
	showVersion := flag.Bool("v", false, "print version")
	force := flag.Bool("force", false, "force overwrite of output file")
	paddingX := flag.Int("px", 50, "horizontal padding around video in pixels")
	paddingY := flag.Int("py", 40, "vertical padding around video in pixels")

	flag.Usage = func() {
		ui.Info("Usage:")
		ui.Info("  halofx -i <input> -o <output>(optional) [flags]")
		ui.Info("")
		ui.Info("By default output file will be created in the same directory as input file with name <input>_halofx.<ext>")
		ui.Info("")
		ui.Info("Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	return Config{
		InputPath:   *input,
		OutputPath: *output,
		PaddingX:    *paddingX,
		PaddingY:    *paddingY,
		Force:       *force,
		ShowVersion: *showVersion,
	}
}

func resolveOutputPath(input, output string, force bool) (string, error) {
	if output == "" {
		ext := filepath.Ext(input)
		return fmt.Sprintf("%s_halofx%s", input[:len(input)-len(ext)], ext), nil
	}

	if err := validateOutputPath(output); err != nil {
		return "", err
	}

	if !force {
		if _, err := os.Stat(output); err == nil {
			return "", fmt.Errorf("output file already exists (use --force to overwrite)")
		}
	}

	return output, nil
}

func videoFileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}
	if !slices.Contains(videoExtensions, filepath.Ext(path)) {
		return fmt.Errorf("file is not a supported video format")
	}
	return nil
}

func validateOutputPath(path string) error {
	dir := filepath.Dir(path)

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("output directory does not exist")
	}
	if !info.IsDir() {
		return fmt.Errorf("output directory is not a directory")
	}
	return nil
}

func fitInside(srcWidth, srcHeight, maxWidth, maxHeight int) (int, int) {
	scaleW := float64(maxWidth) / float64(srcWidth)
	scaleH := float64(maxHeight) / float64(srcHeight)

	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}

	if scale > 1 {
		scale = 1 
	}

	return int(float64(srcWidth) * scale), int(float64(srcHeight) * scale)
}