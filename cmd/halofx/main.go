package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"github.com/devanshu0x/halofx/internal/ui"
)

var version = "dev"

var videoExtensions = []string{".mp4", ".mkv", ".avi", ".mov", ".flv", ".wmv", ".mpeg", ".mpg", ".webm"}

func main() {
	input := flag.String("i", "", "input file")
	output := flag.String("o", "", "output file")
	showVersion := flag.Bool("v", false, "print version")
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
	if *input == "" {
		ui.Error("Input file is required, specify with -i flag")
		os.Exit(1)
	}
	err := videoFileExists(*input)
	if err != nil {
		ui.Error(fmt.Sprintf("Error with input file: %v", err.Error()))
		os.Exit(1)
	}

	var outputPath string
	if *output == "" {
		outputPath = fmt.Sprintf("%s_halofx%s", (*input)[:len(*input)-len(filepath.Ext(*input))], filepath.Ext(*input))
	} else {
		err = validateOutputPath(*output)
		if err != nil {
			ui.Error(fmt.Sprintf("Error with output path: %v", err.Error()))
			os.Exit(1)
		}
		outputPath = *output
	}

	ui.Info(outputPath)

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
