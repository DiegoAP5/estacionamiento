package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	spaceWidth   = 60 
	spaceHeight  = 120
	entryWidth   = 100
	numSpaces    = 20 
	exitXOffset = 10
	lambda = 0.1
	screenWidth  = 900
	screenHeight = 700
	updatesPerCar = 30
	maxYPosition = 100
	estadoEntrando = "Entrando"
    estadoSaliendo = "Saliendo"
    estadoNinguno  = "Ninguno"
)

var (
	carImage *ebiten.Image
	timeSinceLastCar float64 
)

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

type ParkingSpace struct {
	X, Y        float64
	IsOccupied  bool
}

type Game struct {
	cars        []Car
    spaces      []ParkingSpace
	waitingCars []Car
	estadoEntradaSalida string
    semaforoEntrada     chan struct{}
}

func loadCarImage() error {
    var err error
    carImage, _, err = ebitenutil.NewImageFromFile("./assets/images/car.png")
    return err
}

func NewGame() *Game {
	g := &Game{
        estadoEntradaSalida: estadoNinguno,
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
		g.spaces = append(g.spaces, ParkingSpace{
			X: float64(x),
			Y: float64(y),
		})
	}
	
	return g
}

func (g *Game) Update() error {
    currentTime := time.Now()
    timeSinceLastCar += 1.0 / 20.0

    // Generación de carros
    if timeSinceLastCar >= 2 && len(g.cars) < 100 {
        select {
        case <-g.semaforoEntrada:

            g.AddCar()
            timeSinceLastCar = 0
            g.estadoEntradaSalida = estadoEntrando
            
            go func() {
                time.Sleep(100 * time.Millisecond)
                g.estadoEntradaSalida = estadoNinguno
                g.semaforoEntrada <- struct{}{}
            }()
        default:
            
        }
    }

    for i := len(g.cars) - 1; i >= 0; i-- {
        car := &g.cars[i]

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
                g.estadoEntradaSalida = estadoSaliendo
                g.spaces[car.SpaceIndex].IsOccupied = false
                g.cars = append(g.cars[:i], g.cars[i+1:]...)
                go func() {
                    time.Sleep(100 * time.Millisecond)
                    g.estadoEntradaSalida = estadoNinguno
                    g.semaforoEntrada <- struct{}{}
                }()
            default:
                
            }
        }
    }

    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
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

	textoEstado := fmt.Sprintf("Estado: %s", g.estadoEntradaSalida)
    ebitenutil.DebugPrintAt(screen, textoEstado, screenWidth-150, 10)

    entryX := (screenWidth - entryWidth) / 2
    ebitenutil.DrawRect(screen, float64(entryX), 0, float64(entryWidth), 20, color.RGBA{0, 255, 0, 255})

    for _, car := range g.cars {
        opts := &ebiten.DrawImageOptions{}
        opts.GeoM.Translate(car.X, car.Y)
        screen.DrawImage(car.Image, opts)
    }
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) AddCar() {
	
	carImage := ebiten.NewImage(50, 100)
	ebitenutil.DrawRect(carImage, 0, 0, 50, 100, color.RGBA{255, 0, 0, 255})

	entryX := (screenWidth - entryWidth) / 2
	car := Car{
		Image:  carImage,
		X:      float64(entryX),
		Y:      0,
		Width:  50,
		Height: 100,
	}
	g.cars = append(g.cars, car)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Estacionamiento")

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}