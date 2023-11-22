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
	poisonChan chan struct{}
)

func main() {
    oak.AddScene("main", scene.Scene{
        Start: MainScene,
    })

    oak.Init("main", func(c oak.Config) (oak.Config, error) {
        c.BatchLoad = true
        c.Assets.ImagePath = "assets/images"
        return c, nil
    })
}


type Car struct {
    area     floatgeom.Rect2
    entity   *entities.Entity
    mutex    sync.Mutex
    parkSpot *Boxes
}

func sendPoisonAfterDuration(duration time.Duration) {
    time.Sleep(duration)
    close(poisonChan)
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

func (c *Car) AddToQueue() {
	c.Move("down", 95, 1)
}

func (c *Car) EntryParking() {
	c.Move("down", entranceSpot, 1)
}

func (c *Car) ExitParking() {
	c.Move("up", exitSpot, -1)
}


func (c *Car) Park(spot *Boxes) {
	withMutex(&gateMutex, func() {
		c.EntryParking()
	})

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

func (c *Car) Leave(spot *Boxes) {
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

func (c *Car) GoAway() {
	c.Move("up", -20, -1)
}

func (c *Car) safeExecute(action func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	action()
}

func (c *Car) MoveVertically(dy float64) {
	c.safeExecute(func() {
		c.entity.ShiftY(dy)
	})
}

func (c *Car) MoveHorizontally(dx float64) {
	c.safeExecute(func() {
		c.entity.ShiftX(dx)
	})
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

func (c *Car) SetParkSpot(spot *Boxes) {
	c.parkSpot = spot
}

func (c *Car) GetParkSpot() *Boxes {
	return c.parkSpot
}

func (c *Car) Remove() {
	c.safeExecute(func() {
		c.entity.Destroy()
	})
}
func (c *Car) isCollision(direction string) bool {
	cars := GetCars()
	for _, car := range cars {
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
	safeExecute(func() {
		for i, c := range carsList {
			if c == car {
				carsList = append(carsList[:i], carsList[i+1:]...)
				break
			}
		}
	})
}

func GetCars() (cars []*Car) {
	safeExecute(func() {
		cars = carsList
	})
	return cars
}

type Boxes struct {
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
	slots         []*Boxes
	queueCars     []Car
	mutex         sync.Mutex
	availableCond *sync.Cond
}

func NewParkingSlot(x, y, x2, y2 float64, column int) *Boxes {
	directionsForParking := getDirectionForParking(x, y, column)
	directionsForLeaving := getDirectionsForLeaving()
	area := floatgeom.NewRect2(x, y, x2, y2)

	return &Boxes{
		area:                 &area,
		isAvailable:          true,
		directionsForParking: directionsForParking,
		directionsForLeaving: directionsForLeaving,
	}
}

func (p *Boxes) GetArea() *floatgeom.Rect2 {
	return p.area
}

func (p *Boxes) GetDirectionsForParking() []*struct {
	Direction string
	Location  float64
} {
	return p.directionsForParking
}

func (p *Boxes) GetDirectionsForLeaving() []*struct {
	Direction string
	Location  float64
} {
	return p.directionsForLeaving
}

func (p *Boxes) GetIsAvailable() bool {
	return p.isAvailable
}

func (p *Boxes) SetIsAvailable(isAvailable bool) {
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
	slots := make([]*Boxes, 4*5)
	queueCars := make([]Car, 0)
	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			x := 410 - i*90
			y := 210 + j*45
			slots[i*columns+j] = NewParkingSlot(float64(x), float64(y), float64(x+30), float64(y+30), i+1)
		}
	}

	p := &Parking{
		slots:     slots,
		queueCars: queueCars,
	}
	p.availableCond = sync.NewCond(&p.mutex)
	return p
}

func (p *Parking) GetSlots() []*Boxes {
	return p.slots
}

func (p *Parking) safeExecute(action func()) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	action()
}

func (p *Parking) GetAvailableParkingSlot() *Boxes {
    p.mutex.Lock()
    defer p.mutex.Unlock()

    for {
        for _, spot := range p.slots {
            if spot.GetIsAvailable() {
                spot.SetIsAvailable(false)
                return spot
            }
        }
        p.availableCond.Wait()
    }
}

func (p *Parking) findAvailableSpot() *Boxes {
    for {
        p.mutex.Lock()
        for _, spot := range p.slots {
            if spot.GetIsAvailable() {
                p.mutex.Unlock()
                return spot
            }
        }
        p.mutex.Unlock()
    }
}

func (p *Parking) MakeParkingSlotAvailable(spot *Boxes) {
    p.mutex.Lock()
    spot.SetIsAvailable(true)
    p.availableCond.Signal()
    p.mutex.Unlock()
}

func (p *Parking) GetCarsInQueue() (cars []Car) {
	p.safeExecute(func() {
		cars = p.queueCars
	})
	return cars
}

func MainScene(ctx *scene.Context) {
	parking = NewParking()
	crearEstacionamiento(ctx)

	event.GlobalBind(ctx, event.Enter, func(enterPayload event.EnterPayload) event.Response {
		for i := 0; i < 100; i++ {
			go ejecutar(ctx)
			random(1000, 2000)
		}
		return 0
	})
}

func crearEstacionamiento(ctx *scene.Context) {
    backgroundColor := color.RGBA{86, 101, 115, 255}

    screenWidth := 800 
    screenHeight := 600 

    fullScreenRect := render.NewColorBox(screenWidth, screenHeight, backgroundColor)

    render.Draw(fullScreenRect, 0)

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

func ejecutar(ctx *scene.Context) {
	c := crear(ctx)
	estacionar(c)
	salir(c)
	eliminar(c)
}

func crear(ctx *scene.Context) *Car {
	c := NewCar(ctx)
	AddCar(c)
	c.AddToQueue()
	return c
}

func estacionar(c *Car) {
	spotAvailable := parking.GetAvailableParkingSlot()
	withMutex(&gateMutex, c.EntryParking)
	c.Park(spotAvailable)
	random(400, 50000)
	c.LeaveSpot()
	parking.MakeParkingSlotAvailable(spotAvailable)
	c.Leave(spotAvailable)
}

func salir(c *Car) {
	withMutex(&gateMutex, c.ExitParking)
	c.GoAway()
}

func eliminar(c *Car) {
	c.Remove()
	RemoveCar(c)
}

func random(min, max int) {
	randomGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomDuration := time.Duration(randomGen.Intn(max-min+1) + min)
	time.Sleep(time.Millisecond * randomDuration)
}
