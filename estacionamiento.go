package main

import (
	"image/color"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/oakmound/oak/v4"
	"github.com/oakmound/oak/v4/alg/floatgeom"
	"github.com/oakmound/oak/v4/entities"
	"github.com/oakmound/oak/v4/event"
	"github.com/oakmound/oak/v4/render"
	"github.com/oakmound/oak/v4/scene"
)

// Constantes y variables globales
const (
	entranceSpot = 135
	exitSpot     = 145
	speed        = 10
)
var (
	rows       = 4
	columns    = 5
	parking    *Parking
	gateMutex  sync.Mutex
	carsList   []*Car
	carsListMutex sync.Mutex
)

// Definiciones de estructuras
type ParkingSlot struct {
	area                 *floatgeom.Rect2
	isAvailable          bool
	directionsForParking []*struct {
		Direction string
		Location  float64
	}
	directionsForLeaving []*struct {
		Direction string
		Location  float64
	}
}

type Parking struct {
	slots         []*ParkingSlot
	queueCars     []Car
	mutex         sync.Mutex
	availableCond *sync.Cond
}

type Car struct {
	area     floatgeom.Rect2
	entity   *entities.Entity
	mutex    sync.Mutex
	parkSpot *ParkingSlot
}

func main() {
	oak.AddScene("main", scene.Scene{
		Start: mainScene,
	})

	oak.Init("main", func(c oak.Config) (oak.Config, error) {
		c.BatchLoad = true
		c.Assets.ImagePath = "assets/images"
		c.Assets.AudioPath = "assets/audio"
		return c, nil
	})
}

func NewParkingSlot(x, y, x2, y2 float64, column int) *ParkingSlot {
	directionsForParking := getDirectionForParking(x, y, column)
	directionsForLeaving := getDirectionsForLeaving()
	area := floatgeom.NewRect2(x, y, x2, y2)

	return &ParkingSlot{
		area:                 &area,
		isAvailable:          true,
		directionsForParking: directionsForParking,
		directionsForLeaving: directionsForLeaving,
	}
}

func (p *ParkingSlot) GetArea() *floatgeom.Rect2 {
	return p.area
}

func (p *ParkingSlot) GetDirectionsForParking() []*struct {
	Direction string
	Location  float64
} {
	return p.directionsForParking
}

func (p *ParkingSlot) GetDirectionsForLeaving() []*struct {
	Direction string
	Location  float64
} {
	return p.directionsForLeaving
}

func (p *ParkingSlot) GetIsAvailable() bool {
	return p.isAvailable
}

func (p *ParkingSlot) SetIsAvailable(isAvailable bool) {
	p.isAvailable = isAvailable
}

func getDirectionForParking(x, y float64, column int) []*struct {
	Direction string
	Location  float64
} {
	var directions []*struct {
		Direction string
		Location  float64
	}

	leftLocation := map[int]float64{
		1: 445,
		2: 355,
		3: 265,
		4: 175,
	}

	if location, ok := leftLocation[column]; ok {
		directions = append(directions, &struct {
			Direction string
			Location  float64
		}{
			"left", location,
		})
	}

	directions = append(directions, &struct {
		Direction string
		Location  float64
	}{
		"down", y + 5,
	})
	directions = append(directions, &struct {
		Direction string
		Location  float64
	}{
		"left", x + 5,
	})

	return directions
}

func getDirectionsForLeaving() []*struct {
	Direction string
	Location  float64
} {
	var directions []*struct {
		Direction string
		Location  float64
	}

	directions = append(directions, &struct {
		Direction string
		Location  float64
	}{
		"down", 425,
	})
	directions = append(directions, &struct {
		Direction string
		Location  float64
	}{
		"right", 475,
	})
	directions = append(directions, &struct {
		Direction string
		Location  float64
	}{
		"up", 135,
	})

	return directions
}

func NewParking() *Parking {
	slots := make([]*ParkingSlot, rows*columns)
	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			x := 410 - i*90
			y := 210 + j*45
			slots[i*columns+j] = NewParkingSlot(float64(x), float64(y), float64(x+30), float64(y+30), i+1)
		}
	}

	p := &Parking{
		slots: slots,
	}
	p.availableCond = sync.NewCond(&p.mutex)
	return p
}

func (p *Parking) GetSlots() []*ParkingSlot {
	return p.slots
}

func (p *Parking) safeExecute(action func()) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	action()
}

func (p *Parking) GetAvailableParkingSlot() *ParkingSlot {
	var spot *ParkingSlot
	p.safeExecute(func() {
		spot = p.findAvailableSpot()
		if spot != nil {
			spot.SetIsAvailable(false)
		}
	})
	return spot
}

func (p *Parking) findAvailableSpot() *ParkingSlot {
	for _, spot := range p.slots {
		if spot.GetIsAvailable() {
			return spot
		}
	}
	p.availableCond.Wait()
	return p.findAvailableSpot()
}

func (p *Parking) MakeParkingSlotAvailable(spot *ParkingSlot) {
	p.safeExecute(func() {
		spot.SetIsAvailable(true)
		p.availableCond.Signal()
	})
}

func (p *Parking) GetCarsInQueue() (cars []Car) {
	p.safeExecute(func() {
		cars = p.queueCars
	})
	return cars
}

func NewCar(ctx *scene.Context) *Car {
	area := floatgeom.NewRect2WH(445, -20, 32, 32)

	carRender, err := render.LoadSprite("assets/images/car.png")
	if err != nil {
		log.Fatal(err)
	}

	entity := entities.New(
		ctx,
		entities.WithRect(area),
		entities.WithColor(color.RGBA{255, 0, 0, 255}),
		entities.WithRenderable(carRender),
		entities.WithDrawLayers([]int{2, 3}),
	)

	return &Car{
		area:   area,
		entity: entity,
	}
}

func (c *Car) Move(direction string, target float64, step float64) {
	for {
		current := c.getCoordinate(direction)
		if (step > 0 && current >= target) || (step < 0 && current <= target) {
			break
		}
		if !c.isCollision(direction) {
			c.shift(direction, step)
		}
		time.Sleep(speed * time.Millisecond)
	}
}

func (c *Car) shift(direction string, step float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if direction == "left" || direction == "right" {
		c.entity.ShiftX(step)
	} else {
		c.entity.ShiftY(step)
	}
}

func (c *Car) getCoordinate(direction string) float64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if direction == "left" || direction == "right" {
		return c.entity.X()
	}
	return c.entity.Y()
}

func (c *Car) isCollision(direction string) bool {
	carsListMutex.Lock()
	defer carsListMutex.Unlock()
	for _, car := range carsList {
		if car == c {
			continue // Skip self
		}
		switch direction {
		case "left":
			if c.isCollisionLeft(car) {
				return true
			}
		case "right":
			if c.isCollisionRight(car) {
				return true
			}
		case "up":
			if c.isCollisionUp(car) {
				return true
			}
		case "down":
			if c.isCollisionDown(car) {
				return true
			}
		}
	}
	return false
}

func (c *Car) isCollisionLeft(car *Car) bool {
	minDistance := 30.0
	return c.X() > car.X() && c.X()-car.X() < minDistance && c.Y() == car.Y()
}

func (c *Car) isCollisionRight(car *Car) bool {
	minDistance := 30.0
	return c.X() < car.X() && car.X()-c.X() < minDistance && c.Y() == car.Y()
}

func (c *Car) isCollisionUp(car *Car) bool {
	minDistance := 30.0
	return c.Y() > car.Y() && c.Y()-car.Y() < minDistance && c.X() == car.X()
}

func (c *Car) isCollisionDown(car *Car) bool {
	minDistance := 30.0
	return c.Y() < car.Y() && car.Y()-c.Y() < minDistance && c.X() == car.X()
}

func safeExecute(action func()) {
	carsListMutex.Lock()
	defer carsListMutex.Unlock()
	action()
}

func AddCar(car *Car) {
	safeExecute(func() {
		carsList = append(carsList, car)
	})
}

func RemoveCar(car *Car) {
	carsListMutex.Lock()
	for i, c := range carsList {
		if c == car {
			carsList = append(carsList[:i], carsList[i+1:]...)
			break
		}
	}
	carsListMutex.Unlock()
}

func GetCars() []*Car {
	carsListMutex.Lock()
	defer carsListMutex.Unlock()
	copiedCarsList := make([]*Car, len(carsList))
	copy(copiedCarsList, carsList)
	return copiedCarsList
}

func mainScene(ctx *scene.Context) {
	parking = NewParking()
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
    // Define el color de fondo sólido
    backgroundColor := color.RGBA{86, 101, 115, 255} // Gris azulado

    // Asume dimensiones estáticas de la ventana
    screenWidth := 800  // Reemplaza con el ancho real de la ventana
    screenHeight := 600 // Reemplaza con la altura real de la ventana

    // Crea un rectángulo de color que cubra toda la pantalla
    fullScreenRect := render.NewColorBox(screenWidth, screenHeight, backgroundColor)

    // Añade el rectángulo de color al renderizado
    render.Draw(fullScreenRect, 0) // 0 es la capa base

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

    // Dibuja los límites del estacionamiento
    // Estos valores deben ser ajustados según tu diseño específico
    limiteIzquierdo := render.NewLine(100, 125, 100, float64(screenHeight-150), color.White)
    limiteDerecho := render.NewLine(float64(screenWidth-300), 125, float64(screenWidth-300), float64(screenHeight-150), color.White)
    limiteSuperior := render.NewLine(100, 125, float64(screenWidth-350), 125, color.White)
    limiteInferior := render.NewLine(100, float64(screenHeight-150), float64(screenWidth-300), float64(screenHeight-150), color.White)

    render.Draw(limiteIzquierdo, 0)
    render.Draw(limiteDerecho, 0)
    render.Draw(limiteSuperior, 0)
    render.Draw(limiteInferior, 0)

    // Dibuja la entrada del estacionamiento
    entrada := render.NewLine(100, float64(screenHeight-100), 100, float64(screenHeight), color.White)
    render.Draw(entrada, 0)
}


func run(ctx *scene.Context) {
	c := createCar(ctx)
	parkCar(c)
	exitCar(c)
	removeCar(c)
}

func createCar(ctx *scene.Context) *Car {
	c := NewCar(ctx)
	AddCar(c)
	return c
}

func parkCar(c *Car) {
	spotAvailable := parking.GetAvailableParkingSlot()
	c.Park(spotAvailable)
	sleepRandomDuration(40000, 50000)
	c.LeaveSpot()
	parking.MakeParkingSlotAvailable(spotAvailable)
}

func exitCar(c *Car) {
	c.ExitParking()
	c.GoAway()
}

func removeCar(c *Car) {
	RemoveCar(c)
	c.Remove()
}

func sleepRandomDuration(min, max int) {
	randDuration := time.Duration(rand.Intn(max-min+1) + min)
	time.Sleep(randDuration * time.Millisecond)
}

func (c *Car) AddToQueue() {
	c.Move("down", entranceSpot, 1)
}

func (c *Car) EntryParking() {
	c.Move("down", entranceSpot, 1)
}

func (c *Car) ExitParking() {
	c.Move("up", exitSpot, -1)
}

func (c *Car) Park(spot *ParkingSlot) {
	// Asegurarse de que el coche espere su turno para entrar al estacionamiento
	withMutex(&gateMutex, func() {
		c.EntryParking()
	})

	// Mover el coche a su plaza asignada
	directions := spot.GetDirectionsForParking()
	for _, direction := range directions {
		if direction.Direction == "right" || direction.Direction == "down" {
			c.Move(direction.Direction, direction.Location, 1)
		} else {
			c.Move(direction.Direction, direction.Location, -1)
		}
	}

	// Estacionar el coche
	withMutex(&c.mutex, func() {
		c.parkSpot = spot
	})
}

func (c *Car) Leave(spot *ParkingSlot) {
	directions := spot.GetDirectionsForLeaving()
	for _, direction := range directions {
		if direction.Direction == "right" || direction.Direction == "down" {
			c.Move(direction.Direction, direction.Location, 1)
		} else {
			c.Move(direction.Direction, direction.Location, -1)
		}
	}
}

func (c *Car) LeaveSpot() {
	spotX := c.X()
	c.Move("left", spotX-30, -1)
}

func withMutex(m *sync.Mutex, action func()) {
	m.Lock()
	defer m.Unlock()
	action()
}

func (c *Car) GoAway() {
	c.Move("up", -20, -1)
}

func (c *Car) Remove() {
	c.entity.Destroy()
}

func (c *Car) safeExecute(action func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	action()
}

func (c *Car) X() float64 {
	var x float64
	c.safeExecute(func() {
		x = c.entity.X()
	})
	return x
}

func (c *Car) Y() float64 {
	var y float64
	c.safeExecute(func() {
		y = c.entity.Y()
	})
	return y
}

func (c *Car) SetParkSpot(spot *ParkingSlot) {
	c.parkSpot = spot
}

func (c *Car) GetParkSpot() *ParkingSlot {
	return c.parkSpot
}
