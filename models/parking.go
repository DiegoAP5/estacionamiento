package modelos

import (
	"sync"

	"github.com/oakmound/oak/v4/alg/floatgeom"
)

var (
	rows       = 4
	columns    = 5
	parking    *Parking
	gateMutex  sync.Mutex
)

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