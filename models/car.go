package modelos

import (
	"image/color"
	"log"
	"sync"
	"time"

	"github.com/oakmound/oak/v4/alg/floatgeom"
	"github.com/oakmound/oak/v4/entities"
	"github.com/oakmound/oak/v4/scene"
	"github.com/oakmound/oak/v4/render"
)

const (
	speed        = 10
	entranceSpot = 135
	exitSpot     = 145
)

var (
	carsList   []*Car
	carsListMutex sync.Mutex
	poisonChan chan struct{}
)

type Car struct {
    area     floatgeom.Rect2
    entity   *entities.Entity
    mutex    sync.Mutex
    parkSpot *ParkingSlot
}

func NewCar(ctx *scene.Context) *Car {
	area := floatgeom.NewRect2WH(445, -20, 32, 32)

	carRender, err := render.LoadSprite("./assets/images/car.png")
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
		} else {
			break // Si hay una colisión, detiene el movimiento
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
			continue
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

func sendPoisonAfterDuration(duration time.Duration) {
    time.Sleep(duration)
    close(poisonChan) // Enviar señal de "poison" cerrando el canal
}

func (c *Car) AddToQueue() {
	c.Move("down", 95, 1)
}

func (c *Car) EntryParking() {
	c.Move("down", entranceSpot, 1)
}

func (c *Car) ExitParking() {
	if c.parkSpot != nil {
		directions := c.parkSpot.GetDirectionsForLeaving()
		for _, direction := range directions {
			var step float64
			if direction.Direction == "right" || direction.Direction == "down" {
				step = 1
			} else {
				step = -1
			}
			c.Move(direction.Direction, direction.Location, step)
		}
	}
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

func (c *Car) LeaveSpot() {
	spotX := c.X()
	c.Move("left", spotX-30, -1)
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

func withMutex(m *sync.Mutex, action func()) {
	m.Lock()
	defer m.Unlock()
	action()
}