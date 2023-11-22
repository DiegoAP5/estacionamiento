package scenas

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"concurrentec2/models"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	spaceWidth     = 60
	spaceHeight    = 120
	entryWidth     = 100
	numSpaces      = 20
	exitXOffset    = 10
	lambda         = 0.1
	maxYPosition   = 100
	estadoEntrando = "Entrando"
	estadoSaliendo = "Saliendo"
	estadoNinguno  = "Ninguno"
	screenHeight   = 700
	screenWidth    = 900
)

var (
	carImage         *ebiten.Image
	timeSinceLastCar float64
)

func loadCarImage() error {
	var err error
	carImage, _, err = ebitenutil.NewImageFromFile("./assets/images/car.png")
	return err
}

type Game struct {
	Cars                []modelos.Car
	spaces              []modelos.ParkingSpace
	waitingCars         []modelos.Car
	EstadoEntradaSalida string
	semaforoEntrada     chan struct{}
}

func NewGame() *Game {
	g := &Game{
		EstadoEntradaSalida: estadoNinguno,
		semaforoEntrada:     make(chan struct{}, 1),
	}
	g.semaforoEntrada <- struct{}{}

	if err := loadCarImage(); err != nil {
		log.Fatal(err)
	}

	parkingWidth := 10 * spaceWidth
	startX := (screenWidth - parkingWidth) / 2
	startY := (screenHeight - 2*spaceHeight) / 2

	for i := 0; i < numSpaces; i++ {
		x := startX + (i%10)*spaceWidth
		y := startY + (i/10)*spaceHeight
		g.spaces = append(g.spaces, modelos.ParkingSpace{
			X: float64(x),
			Y: float64(y),
		})
	}

	return g
}

func (g *Game) AddCar() {
	entryX := (screenWidth - entryWidth) / 2
	car := modelos.Car{
		Image:  carImage,
		X:      float64(entryX),
		Y:      0,
		Width:  50,
		Height: 100,
	}
	g.Cars = append(g.Cars, car)
}

func (g *Game) Update() error {
	currentTime := time.Now()
	timeSinceLastCar += 1.0 / 20.0

	// Generación de carros
	if timeSinceLastCar >= 2 && len(g.Cars) < 100 {
		select {
		case <-g.semaforoEntrada:

			g.AddCar()
			timeSinceLastCar = 0
			g.EstadoEntradaSalida = estadoEntrando

			go func() {
				time.Sleep(100 * time.Millisecond)
				g.EstadoEntradaSalida = estadoNinguno
				g.semaforoEntrada <- struct{}{}
			}()
		default:

		}
	}

	for i := len(g.Cars) - 1; i >= 0; i-- {
		car := &g.Cars[i]

		if !car.IsParked && car.Y < maxYPosition {
			car.Y += 2
		}

		// Estacionar
		if !car.IsParked && car.Y >= maxYPosition {
			for j := range g.spaces {
				if !g.spaces[j].IsOccupied {
					g.spaces[j].IsOccupied = true
					car.IsParked = true
					car.X = g.spaces[j].X
					car.Y = g.spaces[j].Y
					car.ParkedTime = currentTime
					car.LeaveAfter = time.Duration(5+rand.Intn(6)) * time.Second
					car.SpaceIndex = j
					break
				}
			}
		}

		// Salida
		if car.IsParked && currentTime.Sub(car.ParkedTime) >= car.LeaveAfter {
			select {
			case <-g.semaforoEntrada:
				g.EstadoEntradaSalida = estadoSaliendo
				g.spaces[car.SpaceIndex].IsOccupied = false
				g.Cars = append(g.Cars[:i], g.Cars[i+1:]...)
				go func() {
					time.Sleep(100 * time.Millisecond)
					g.EstadoEntradaSalida = estadoNinguno
					g.semaforoEntrada <- struct{}{}
				}()
			default:

			}
		}
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}


func (g *Game) Draw(screen *ebiten.Image) {
	DrawGame(screen, g)
}

func DrawGame(screen *ebiten.Image, game *Game) {
	drawParkingLot(screen)
	drawCars(screen, game)
	drawGameState(screen, game.EstadoEntradaSalida)
}

func drawParkingLot(screen *ebiten.Image) {
	white := color.White
	parkingWidth := 10 * spaceWidth
	parkingHeight := 2 * spaceHeight
	startX := (screenWidth - parkingWidth) / 2
	startY := (screenHeight - parkingHeight) / 2

	for i := 0; i < 20; i++ {
		x := startX + (i%10)*spaceWidth
		y := startY + (i/10)*spaceHeight
		ebitenutil.DrawLine(screen, float64(x), float64(y), float64(x+spaceWidth), float64(y), white)
		ebitenutil.DrawLine(screen, float64(x), float64(y), float64(x), float64(y+spaceHeight), white)
		ebitenutil.DrawLine(screen, float64(x+spaceWidth), float64(y), float64(x+spaceWidth), float64(y+spaceHeight), white)
		ebitenutil.DrawLine(screen, float64(x), float64(y+spaceHeight), float64(x+spaceWidth), float64(y+spaceHeight), white)
	}

	entryX := (screenWidth - entryWidth) / 2
	ebitenutil.DrawRect(screen, float64(entryX), 0, float64(entryWidth), 20, color.RGBA{0, 255, 0, 255})
}

func drawCars(screen *ebiten.Image, game *Game) {
	for _, car := range game.Cars {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(car.X, car.Y)
		screen.DrawImage(car.Image, opts)
	}
}

func drawGameState(screen *ebiten.Image, estado string) {
	textoEstado := fmt.Sprintf("Estado: %s", estado)
	ebitenutil.DebugPrintAt(screen, textoEstado, screenWidth-150, 10)
}
