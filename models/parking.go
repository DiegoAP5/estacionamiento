package modelos

import (
	"image/color"
	"sync"

	"github.com/oakmound/oak/render"
	"github.com/oakmound/oak/v4/alg/floatgeom"
)

var (
	rows       = 4
	columns    = 5
	parking    *Parking
	gateMutex  sync.Mutex
)

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
			line := render.NewLine(float64(x), float64(y), float64(x+30), float64(y+30), color.White)
			render.Draw(line, 2)
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
