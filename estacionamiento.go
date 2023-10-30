package main

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	capacity = 20
)

var (
	mu                sync.Mutex
	vehiclesInParking int
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Estacionamiento")

	vehiclesInParking = 0
	parkingSpaces := make(chan struct{}, capacity)

	// Interface elements
	parkingStatus := widget.NewLabel(fmt.Sprintf("Vehículos en el estacionamiento: %d", vehiclesInParking))
	enterButton := widget.NewButton("Entrar", func() {
		go enterParking(parkingSpaces, parkingStatus)
	})
	exitButton := widget.NewButton("Salir", func() {
		go exitParking(parkingSpaces, parkingStatus)
	})

	content := container.NewVBox(
		parkingStatus,
		enterButton,
		exitButton,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func enterParking(parkingSpaces chan struct{}, parkingStatus *widget.Label) {
	parkingSpaces <- struct{}{}
	mu.Lock()
	vehiclesInParking++
	parkingStatus.SetText(fmt.Sprintf("Vehículos en el estacionamiento: %d", vehiclesInParking))
	mu.Unlock()

	// Simula el tiempo en el estacionamiento
	time.Sleep(time.Duration(1+time.Now().UnixNano()%5) * time.Second)

	<-parkingSpaces
	mu.Lock()
	vehiclesInParking--
	parkingStatus.SetText(fmt.Sprintf("Vehículos en el estacionamiento: %d", vehiclesInParking))
	mu.Unlock()
}

func exitParking(parkingSpaces chan struct{}, parkingStatus *widget.Label) {
	// Simula el tiempo de salida
	time.Sleep(time.Duration(1+time.Now().UnixNano()%5) * time.Second)
	parkingSpaces <- struct{}{}
	mu.Lock()
	vehiclesInParking--
	parkingStatus.SetText(fmt.Sprintf("Vehículos en el estacionamiento: %d", vehiclesInParking))
	mu.Unlock()
}
