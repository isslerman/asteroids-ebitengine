package goasteroids

import (
	"g-asteroids/assets"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type GameOverScene struct {
	game        *GameScene
	meteors     map[int]*Meteor
	meteorCount int
	stars       []*Star
}

func (o *GameOverScene) Draw(screen *ebiten.Image) {
	// Draw some stars
	for _, s := range o.stars {
		s.Draw(screen)
	}

	// Draw meteors.
	for _, m := range o.meteors {
		m.Draw(screen)
	}

	textToDraw := "Game Over"
	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, ScreenHeight/2+100)
	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.TitleFont,
		Size:   48,
	}, op)
}

func (o *GameOverScene) Update(state *State) error {
	// Spawn meteors.
	if len(o.meteors) < 10 {
		m := NewMeteor(0.25, &GameScene{}, len(o.meteors)-1)
		o.meteorCount++
		o.meteors[o.meteorCount] = m
	}

	// Update meteors
	for _, m := range o.meteors {
		m.Update()
	}

	// Check to see if spacebar is pressed.
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		o.game.Reset()
		state.SceneManager.GoToScene(o.game)
	}

	// Check to see if q is pressed.
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		// Exit with no error
		os.Exit(0)
	}

	return nil
}
