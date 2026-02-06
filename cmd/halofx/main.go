package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/devanshu0x/halofx/internal/mask"
	"github.com/devanshu0x/halofx/internal/render"
	"github.com/devanshu0x/halofx/internal/ui"
)

var version = "dev"

var videoExtensions = []string{".mp4", ".mkv", ".avi", ".mov", ".flv", ".wmv", ".mpeg", ".mpg", ".webm"}

func main() {
	input := flag.String("i", "", "input file")
	output := flag.String("o", "", "output file")
	showVersion := flag.Bool("v", false, "print version")
	forceOwerwrite := flag.Bool("force", false, "force overwrite of output file if it exists")
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

	if *showVersion {
		ui.Info("halofx version: " + version)
		os.Exit(0)
	}
	var inputPath, outputPath string

	if *input == "" {
		ui.Error("Input file is required, specify with -i flag")
		os.Exit(1)
	}
	err := videoFileExists(*input)
	if err != nil {
		ui.Error(fmt.Sprintf("Error with input file: %v", err.Error()))
		os.Exit(1)
	}
	inputPath = *input

	if *output == "" {
		outputPath = fmt.Sprintf("%s_halofx%s", (*input)[:len(*input)-len(filepath.Ext(*input))], filepath.Ext(*input))
	} else {
		err = validateOutputPath(*output)
		if err != nil {
			ui.Error(fmt.Sprintf("Error with output path: %v", err.Error()))
			os.Exit(1)
		}
		if *forceOwerwrite {
			outputPath = *output
		} else {
			_, err := os.Stat(*output)
			if err == nil {
				ui.Error("Output file already exists. Use --force flag to overwrite.")
				os.Exit(1)
			}
			outputPath = *output
		}
	}

	width, height, err := render.GetVideoDimensions(inputPath)
	if err != nil {
		ui.Error("Failed to get video dimensions: " + err.Error())
		os.Exit(1)
	}

	if width == 0 || height == 0 {
		ui.Error("Invalid video dimensions")
		os.Exit(1)
	}
	paddingX := 30
	paddingY := 20
	adjustedWidth, adjustedHeight := fitInside(width, height, 1920-2*paddingX, 1080-2*paddingY)
	tmpFile, err := os.CreateTemp("", "halofx-mask-*.png")
	if err != nil {
		ui.Error("Failed to create temporary file for mask: " + err.Error())
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name())

	err = mask.GenerateRoundedMask(tmpFile.Name(), adjustedWidth, adjustedHeight, 16)
	if err != nil {
		ui.Error("Failed to generate mask: " + err.Error())
		os.Exit(1)
	}

	err = render.RenderMac(render.MacOptions{
		InputPath:      inputPath,
		OutputPath:     outputPath,
		BackgroundPath: "internal/assets/backgrounds/bg-1.jpg",
		MaskPath:       tmpFile.Name(),
		Width:          1920,
		Height:         1080,
		VideoWidth:     adjustedWidth,
		VideoHeight:    adjustedHeight,
		Force:          *forceOwerwrite,
	})

	if err != nil {
		ui.Error("Render failed: " + err.Error())
		os.Exit(1)
	}

	ui.Success("Output written to " + outputPath)

}

func videoFileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("Path is a directory, not a file")
	}
	ext := filepath.Ext(path)
	if slices.Contains(videoExtensions, ext) {
		return nil
	}
	return fmt.Errorf("File is not a supported video format")
}

func validateOutputPath(path string) error {
	if filepath.Base(path) == "." || filepath.Base(path) == string(os.PathSeparator) {
		return fmt.Errorf("Output path cannot be a directory")
	}

	dir := filepath.Dir(path)

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("Output directory does not exist")
	}
	if !info.IsDir() {
		return fmt.Errorf("Output directory is not a valid directory")
	}
	return nil
}


func fitInside(srcWidth, srcHeight, maxWidth, maxHeight int) (newWidth, newHeight int) {
	scaleW := float64(maxWidth) / float64(srcWidth)
	scaleH := float64(maxHeight) / float64(srcHeight)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}

	// never upscale
	if scale > 1 {
		scale = 1
	}

	newWidth = int(float64(srcWidth) * scale)
	newHeight = int(float64(srcHeight) * scale)
	return
}