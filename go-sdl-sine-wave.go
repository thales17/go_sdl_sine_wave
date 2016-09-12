package main

import (
	"fmt"
	"math"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowTitle  = "Go-SDL2 Render"
	windowWidth  = 1280
	windowHeight = 720
	frameRate    = 60

	cols      = 50
	rows      = 50
	numPoints = (cols-1)*windowHeight + (rows-1)*windowWidth
)

func roundToInt32(a float64) int32 {
	if a < 0 {
		return int32(a - 0.5)
	}

	return int32(a + 0.5)
}

var tx float32 = math.Pi / 9
var ty float32 = math.Pi / 4
var xFreq float32 = 1
var yFreq float32 = 1
var xFreqDir float32 = 1
var yFreqDir float32 = 1

//Called Once Per Frame
func updateDistortionState() {
	xFreq += (0.1) * xFreqDir
	if xFreq > 25 || xFreq < 1 {
		xFreqDir *= -1
	}
	yFreq += (0.1) * yFreqDir
	if yFreq > 30 || yFreq < 1 {
		yFreqDir *= -1
	}
}

func sineWaveDistortXY(x int32, y int32, w int, h int) (int32, int32) {
	var normalizedX = float32(x) / float32(w)
	var normalizedY = float32(y) / float32(h)

	var xOffset = int32(50 * (math.Sin(float64(xFreq*normalizedY+yFreq*normalizedX+2*math.Pi*tx)) * 0.5))
	var yOffset = int32(50 * (math.Sin(float64(xFreq*normalizedY+yFreq*normalizedX+2*math.Pi*ty)) * 0.5))

	return x + xOffset, y + yOffset
}

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var err error

	sdl.CallQueue <- func() {
		window, err = sdl.CreateWindow(windowTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, windowWidth, windowHeight, sdl.WINDOW_OPENGL)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer func() {
		sdl.CallQueue <- func() {
			window.Destroy()
		}
	}()

	sdl.CallQueue <- func() {
		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	}
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer func() {
		sdl.CallQueue <- func() {
			renderer.Destroy()
		}
	}()

	sdl.CallQueue <- func() {
		renderer.Clear()
	}
	var gridPoints = make([]sdl.Point, numPoints)
	var origGridPoints = make([]sdl.Point, numPoints)
	index := 0
	cellWidth := roundToInt32(float64(windowWidth) / float64(cols))
	cellHeight := roundToInt32(float64(windowHeight) / float64(rows))
	// Create Columns
	for i := 1; i < cols; i++ {
		x := i * int(cellWidth)
		for j := 0; j < windowHeight; j++ {
			gridPoints[index] = sdl.Point{X: int32(x), Y: int32(j)}
			index++
		}
	}

	// Create Rows
	for i := 1; i < cols; i++ {
		y := i * int(cellHeight)
		for j := 0; j < windowWidth; j++ {
			gridPoints[index] = sdl.Point{X: int32(j), Y: int32(y)}
			index++
		}
	}

	running := true
	for running {
		sdl.CallQueue <- func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch event.(type) {
				case *sdl.QuitEvent:
					running = false
				}
			}

			renderer.Clear()
			renderer.SetDrawColor(0, 0, 0, 0x20)
			renderer.FillRect(&sdl.Rect{0, 0, windowWidth, windowHeight})

			renderer.SetDrawColor(0x00, 0xff, 0x00, 0xff)
			renderer.DrawPoints(gridPoints)
		}

		// wg := sync.WaitGroup{}
		// mutex := &sync.Mutex{}
		// for i, point := range origGridPoints {
		// 	wg.Add(1)
		// 	go func(i int, point sdl.Point) {
		// 		newX, newY := sineWaveDistortXY(point.X, point.Y, windowWidth, windowHeight)
		// 		fmt.Println(newX, newY)
		// 		mutex.Lock()
		// 		gridPoints[i] = sdl.Point{X: newX, Y: newY}
		// 		mutex.Unlock()
		// 		wg.Done()
		// 	}(i, point)
		// }
		// wg.Wait()

		updateDistortionState()

		sdl.CallQueue <- func() {
			renderer.Present()
			sdl.Delay(1000 / frameRate)
		}
	}

	return 0
}

func main() {
	os.Exit(run())
}
