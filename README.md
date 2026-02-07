# halofx

`halofx` is a small Go-based CLI tool for experimenting with video composition using FFmpeg.

This is my first Go project. The goal is mainly to learn Go project structure, FFmpeg integration, and basic video rendering concepts.

---

## What it does

- Renders a video on top of a background
- Supports simple frame / mask based layouts
- Can blur or scale backgrounds
- Uses FFmpeg filter graphs internally

This project is **not production-ready** and is mostly for learning and experimentation.

---

## Requirements

- Go 1.21+
- FFmpeg installed and available in `$PATH`

Check FFmpeg:
```bash
ffmpeg -version
```

## Build
```
git clone https://github.com/devanshu0x/halofx.git
cd halofx
go build
```

## Usage
```
./halofx -i <input file> [flags](optional)
```


