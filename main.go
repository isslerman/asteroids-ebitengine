package main

import (
	"g-asteroids/goasteroids"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowTitle("Go Asteroids")
	ebiten.SetWindowSize(goasteroids.ScreenWidth, goasteroids.ScreenHeight)

	// We are passing the interface here. It uses 3 methods (Draw, Update and Layout)
	err := ebiten.RunGame(&goasteroids.Game{})
	if err != nil {
		panic(err)
	}
}
