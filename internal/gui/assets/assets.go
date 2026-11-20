package assets

import (
	"embed"

	"fyne.io/fyne/v2"
)

//go:embed icon.png
var iconPNG []byte

var Icon = fyne.NewStaticResource("icon.png", iconPNG)

//go:embed images
var images embed.FS

func Image(name string) (fyne.Resource, error) {
	bytes, err := images.ReadFile(name)

	if err != nil {
		return nil, err
	}

	return fyne.NewStaticResource(name, bytes), nil
}
