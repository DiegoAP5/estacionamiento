package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"math/rand"
	"sync"
	"time"
	"math"
)

const (
	numCajones   = 20
	maxVehiculos = 100
	width        = 800
	height       = 480
)

var (
	mutex          sync.Mutex
	semEntrada     = make(chan bool, 1) // semáforo para la entrada/salida
	cajones        = make([]bool, numCajones)
	vehiculos      = make([]*Vehiculo, 0)
	poissonRate    = 10.0
	content        = container.NewWithoutLayout()
)

type Vehiculo struct {
	ID        int
	Color     color.Color
	X, Y      float32
	Rectangle *canvas.Rectangle
}

func generarVehiculosPoisson() {
	rate := float64(1) / poissonRate
	for len(vehiculos) < maxVehiculos {
		sleepTime := -rate * math.Log(1-rand.Float64())
		time.Sleep(time.Duration(sleepTime) * time.Second)

		v := &Vehiculo{
			ID:    len(vehiculos) + 1,
			Color: color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255},
		}
		mutex.Lock()
		vehiculos = append(vehiculos, v)
		mutex.Unlock()

		go func(v *Vehiculo) {
			semEntrada <- true
			entrarEstacionamiento(v)
			salirEstacionamiento(v)
			<-semEntrada
		}(v)
	}
}

func entrarEstacionamiento(v *Vehiculo) {
	mutex.Lock()
	for i := 0; i < numCajones; i++ {
		if !cajones[i] {
			cajones[i] = true
			v.X = float32((i % 10) * 60 + 50)
			v.Y = float32((i / 10) * 160 + 100)
			v.Rectangle = canvas.NewRectangle(v.Color)
			v.Rectangle.Move(fyne.NewPos(float32(float64(v.X)), float32(float64(v.Y))))
			v.Rectangle.Resize(fyne.NewSize(50, 80))
			content.Add(v.Rectangle)
			break
		}
	}
	mutex.Unlock()
	time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
}

func salirEstacionamiento(v *Vehiculo) {
	mutex.Lock()
	for i := 0; i < numCajones; i++ {
		if cajones[i] && v.X == float32(i%10*60+50) && v.Y == float32(i/10*160+100) {
			cajones[i] = false
			v.Rectangle.Hide()
			break
		}
	}
	mutex.Unlock()
}

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())
	myWindow := myApp.NewWindow("Estacionamiento")
	myWindow.Resize(fyne.NewSize(width, height))

	go generarVehiculosPoisson()

	entrada := canvas.NewRectangle(color.RGBA{0x99, 0x00, 0x00, 0xff})
	entrada.Move(fyne.NewPos(width/2-25, height/2-40))
	entrada.Resize(fyne.NewSize(50, 80))
	content.Add(entrada)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
