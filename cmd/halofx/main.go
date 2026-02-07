package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	_"image/jpeg"
	_"image/png"
	"image/color"
	"os"
	"path/filepath"
	"slices"

	"github.com/devanshu0x/halofx/internal/assets"
	"github.com/devanshu0x/halofx/internal/mask"
	"github.com/devanshu0x/halofx/internal/render"
	"github.com/devanshu0x/halofx/internal/ui"
)

var version = "dev"

var videoExtensions = []string{".mp4", ".mkv", ".avi", ".mov", ".flv", ".wmv", ".mpeg", ".mpg", ".webm"}

const (
	DefaultWidth          = 1920
	DefaultHeight         = 1080
	DefaultRadius         = 16
	DefaultBG             = "1"
	DefaultFrameThickness = 12
)

type Config struct {
	InputPath      string
	OutputPath     string
	PaddingX       int
	PaddingY       int
	Force          bool
	ShowVersion    bool
	Help           bool
	ShowFrame      bool
	FrameWidth     int
	BackgroundPath string
}

func main() {
	cfg := parseFlags()

	if cfg.ShowVersion || cfg.Help {
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

	maskFile, err := os.CreateTemp("", "halofx-mask-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(maskFile.Name())

	if err := mask.GenerateRoundedMask(
		maskFile.Name(),
		adjustedWidth,
		adjustedHeight,
		DefaultRadius,
		color.NRGBA{0, 0, 0, 255},
	); err != nil {
		return err
	}

	frameFile, err := os.CreateTemp("", "halofx-frame-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(frameFile.Name())

	if err := mask.GenerateRoundedMask(
		frameFile.Name(),
		adjustedWidth+2*cfg.FrameWidth,
		adjustedHeight+2*cfg.FrameWidth,
		DefaultRadius,
		color.NRGBA{255, 255, 255, 40},
	); err != nil {
		return err
	}

	bgData, err := resolveBackground(cfg.BackgroundPath)
	if err != nil {
		ui.Error("Invalid background, using default")
		bgData, _ = resolveBackground(DefaultBG)
	}
	ui.Info(fmt.Sprintf("bg size: %d bytes", len(bgData)))

	bgPath, err := writeTempBackground(bgData)
	if err != nil {
		return err
	}
	defer os.Remove(bgPath)

	if err := render.RenderMac(render.MacOptions{
		InputPath:      cfg.InputPath,
		OutputPath:     outputPath,
		BackgroundPath: bgPath,
		MaskPath:       maskFile.Name(),
		Width:          DefaultWidth,
		Height:         DefaultHeight,
		VideoWidth:     adjustedWidth,
		VideoHeight:    adjustedHeight,
		Force:          cfg.Force,
		FramePath:      frameFile.Name(),
	}); err != nil {
		return err
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
	noFrame := flag.Bool("no-frame", false, "disable frame around video")
	frameWidth := flag.Int("frame-width", DefaultFrameThickness, "frame width in pixels")
	backgroundPath := flag.String("bg", DefaultBG, "1-3 for built-in backgrounds or image path")
	help := flag.Bool("h", false, "show help")

	flag.Usage = func() {
		ui.Info("Usage:")
		ui.Info("  halofx -i <input> -o <output>(optional) [flags]")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *noFrame {
		*frameWidth = 0
	}

	return Config{
		InputPath:      *input,
		OutputPath:     *output,
		PaddingX:       *paddingX,
		PaddingY:       *paddingY,
		Force:          *force,
		ShowVersion:    *showVersion,
		Help:           *help,
		ShowFrame:      !*noFrame,
		FrameWidth:     *frameWidth,
		BackgroundPath: *backgroundPath,
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
			return "", fmt.Errorf("output file already exists (use --force)")
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
		return fmt.Errorf("path is a directory")
	}
	if !slices.Contains(videoExtensions, filepath.Ext(path)) {
		return fmt.Errorf("unsupported video format")
	}
	return nil
}

func validateOutputPath(path string) error {
	info, err := os.Stat(filepath.Dir(path))
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid output directory")
	}
	return nil
}

func fitInside(srcW, srcH, maxW, maxH int) (int, int) {
	scale := float64(maxW) / float64(srcW)
	if h := float64(maxH) / float64(srcH); h < scale {
		scale = h
	}
	if scale > 1 {
		scale = 1
	}
	return int(float64(srcW) * scale), int(float64(srcH) * scale)
}

func validateImageFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return nil, fmt.Errorf("invalid file")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("not an image")
	}

	switch format {
	case "jpeg", "png", "gif":
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported image format")
	}
}

func resolveBackground(bg string) ([]byte, error) {
	switch bg {
	case "1":
		return assets.Backgrounds.ReadFile("backgrounds/bg1.jpg")
	case "2":
		return assets.Backgrounds.ReadFile("backgrounds/bg2.jpg")
	case "3":
		return assets.Backgrounds.ReadFile("backgrounds/bg3.jpg")
	default:
		return validateImageFile(bg)
	}
}

func writeTempBackground(data []byte) (string, error) {
	f, err := os.CreateTemp("", "halofx-bg-*.jpg")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return "", err
	}
	f.Close()
	return f.Name(), nil
}
