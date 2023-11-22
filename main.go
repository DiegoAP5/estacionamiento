package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"concurrentec2/scenes"
)

const (
	screenWidth  = 900
	screenHeight = 700
)

func main() {
	rand.Seed(time.Now().UnixNano())
	game := scenas.NewGame()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Estacionamiento")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
