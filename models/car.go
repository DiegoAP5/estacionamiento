package modelos

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var CarImage *ebiten.Image

type Car struct {
	Image         *ebiten.Image
	X, Y          float64
	Width, Height int
	IsParked      bool
	ParkedTime    time.Time
	LeaveAfter    time.Duration
	SpaceIndex    int
	IsLeaving	  bool
}