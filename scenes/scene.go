package escenas

import (
	"concurrentec2/models"
	"image/color"
	"math/rand"
	"sync"
	"time"

	"github.com/oakmound/oak/v4/event"
	"github.com/oakmound/oak/v4/render"
	"github.com/oakmound/oak/v4/scene"
)

var ( 
	parking modelos.Parking
	gateMutex  sync.Mutex
)

func MainScene(ctx *scene.Context) {
	parking = *modelos.NewParking()
	prepareParking(ctx)

	event.GlobalBind(ctx, event.Enter, func(enterPayload event.EnterPayload) event.Response {
		for i := 0; i < 100; i++ {
			go run(ctx)
			sleepRandomDuration(1000, 2000)
		}
		return 0
	})
}

func prepareParking(ctx *scene.Context) {
    backgroundColor := color.RGBA{86, 101, 115, 255}

    screenWidth := 800 
    screenHeight := 600 

    fullScreenRect := render.NewColorBox(screenWidth, screenHeight, backgroundColor)

    render.Draw(fullScreenRect, 0)

    // Dibuja las líneas del estacionamiento
    for _, spot := range parking.GetSlots() {
        area := spot.GetArea()
        areaX1 := area.Min.X()
        areaY1 := area.Min.Y()
        areaX2 := area.Max.X()
        areaY2 := area.Max.Y()

        topLine := render.NewLine(areaX1, areaY1, areaX2, areaY1, color.White)
        render.Draw(topLine, 0)

        leftLine := render.NewLine(areaX1, areaY1, areaX1, areaY2, color.White)
        render.Draw(leftLine, 0)

        bottomLine := render.NewLine(areaX1, areaY2, areaX2, areaY2, color.White)
        render.Draw(bottomLine, 0)
    }

    limiteIzquierdo := render.NewLine(100, 125, 100, float64(screenHeight-150), color.White)
    limiteDerecho := render.NewLine(float64(screenWidth-300), 125, float64(screenWidth-300), float64(screenHeight-150), color.White)
    limiteSuperior := render.NewLine(100, 125, float64(screenWidth-350), 125, color.White)
    limiteInferior := render.NewLine(100, float64(screenHeight-150), float64(screenWidth-300), float64(screenHeight-150), color.White)

    render.Draw(limiteIzquierdo, 0)
    render.Draw(limiteDerecho, 0)
    render.Draw(limiteSuperior, 0)
    render.Draw(limiteInferior, 0)

    entrada := render.NewLine(100, float64(screenHeight-100), 100, float64(screenHeight), color.White)
    render.Draw(entrada, 0)
}

func run(ctx *scene.Context) {
	c := createCar(ctx)
	parkCar(c)
	exitCar(c)
	removeCar(c)
}

func createCar(ctx *scene.Context) *modelos.Car {
	c := modelos.NewCar(ctx)
	modelos.AddCar(c)
	return c
}

func parkCar(c *modelos.Car) {
	spotAvailable := parking.GetAvailableParkingSlot()
	withMutex(&gateMutex, c.EntryParking)
	c.Park(spotAvailable)
	sleepRandomDuration(4000, 50000)
	c.LeaveSpot()
	parking.MakeParkingSlotAvailable(spotAvailable)
	c.Leave(spotAvailable)
}

func withMutex(m *sync.Mutex, action func()) {
	m.Lock()
	defer m.Unlock()
	action()
}

func exitCar(c *modelos.Car) {
	c.ExitParking()
	c.GoAway()
}

func removeCar(c *modelos.Car) {
	modelos.RemoveCar(c)
	c.Remove()
}

func sleepRandomDuration(min, max int) {
	randDuration := time.Duration(rand.Intn(max-min+1) + min)
	time.Sleep(randDuration * time.Millisecond)
}