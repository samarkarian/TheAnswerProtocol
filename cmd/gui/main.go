package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func main() {
	application := app.New()
	mainWindow := application.NewWindow("The Answer Protocol")

	background := canvas.NewRectangle(color.NRGBA{R: 66, G: 66, B: 66, A: 255})

	content := container.NewWithoutLayout(background)

	mainWindow.Resize(fyne.NewSize(1280, 720))
	mainWindow.SetContent(content)
	mainWindow.ShowAndRun()
}
