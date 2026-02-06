package ui

import (
	"github.com/fatih/color"
)

func Info(msg string){
	color.Green(msg)
}

func Error(msg string){
	color.Red(msg)
}	

func Warning(msg string){
	color.Yellow(msg)
}

func Success(msg string){
	color.Green(msg)
}