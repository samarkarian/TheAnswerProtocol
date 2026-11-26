package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/samarkarian/TheAnswerProtocol/internal/gui/assets"
)

func Display() {
	application := app.New()
	application.SetIcon(assets.Icon)
	mainWindow := application.NewWindow("The Answer Protocol")
	mainWindow.SetFullScreen(true)
	mainWindow.SetPadded(false)

	background := canvas.NewRectangle(color.NRGBA{R: 60, G: 60, B: 60, A: 255})
	background.Resize(fyne.NewSize(1280, 720))

	content := container.NewWithoutLayout(background)
	testImage, err := assets.Image("images/test.png")

	if err == nil {
		image := canvas.NewImageFromResource(testImage)
		image.Resize(fyne.NewSize(128, 128))
		content.Add(image)
	} else {
		fmt.Println("error loading image.")
	}

	mainWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		if ke.Name == fyne.KeyEscape {
			mainWindow.Close()
		}
	})
	mainWindow.SetContent(content)
	mainWindow.ShowAndRun()
}
