package main

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowTitle  = "Go-SDL2 Render"
	windowWidth  = 1280
	windowHeight = 720
	frameRate    = 90
	cols         = 50
	rows         = 50
	numPoints    = (cols-1)*windowHeight + (rows-1)*windowWidth
	frameTime    = 1000 / frameRate
)

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
		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
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
	cellWidth := math.Ceil(float64(windowWidth) / float64(cols))
	cellHeight := math.Ceil(float64(windowHeight) / float64(rows))
	// Create Columns
	for i := 1; i < cols-1; i++ {
		x := i * int(cellWidth)
		for j := 0; j < windowHeight; j++ {
			gridPoints[index] = sdl.Point{X: int32(x), Y: int32(j)}
			origGridPoints[index] = sdl.Point{X: int32(x), Y: int32(j)}
			index++
		}
	}

	// Create Rows
	for i := 1; i < rows-1; i++ {
		y := i * int(cellHeight)
		for j := 0; j < windowWidth; j++ {
			gridPoints[index] = sdl.Point{X: int32(j), Y: int32(y)}
			origGridPoints[index] = sdl.Point{X: int32(j), Y: int32(y)}
			index++
		}
	}

	running := true
	useConcurrentDistort := false
	var lastFrameTime uint32 = sdl.GetTicks()
	for running {

		sdl.CallQueue <- func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch t := event.(type) {
				case *sdl.QuitEvent:
					running = false
				case *sdl.KeyUpEvent:
					if t.Keysym.Sym == sdl.K_u {
						useConcurrentDistort = !useConcurrentDistort
					}
				}
			}

			renderer.Clear()
			renderer.SetDrawColor(0, 0, 0, 0x20)
			renderer.FillRect(&sdl.Rect{0, 0, windowWidth, windowHeight})

			renderer.SetDrawColor(0x00, 0xff, 0x00, 0xff)
			renderer.DrawPoints(gridPoints)
		}

		concurrentDistort := func() {
			wg := sync.WaitGroup{}
			mutex := &sync.Mutex{}
			for i, point := range origGridPoints {
				wg.Add(1)
				go func(i int, point sdl.Point) {
					newX, newY := sineWaveDistortXY(point.X, point.Y, windowWidth, windowHeight)
					mutex.Lock()
					gridPoints[i] = sdl.Point{X: newX, Y: newY}
					mutex.Unlock()
					wg.Done()
				}(i, point)
			}
			wg.Wait()
		}

		normalDistort := func() {
			for i, point := range origGridPoints {
				newX, newY := sineWaveDistortXY(point.X, point.Y, windowWidth, windowHeight)
				gridPoints[i] = sdl.Point{X: newX, Y: newY}
			}
		}

		if useConcurrentDistort {
			// fmt.Println("Using concurrentDistort")
			concurrentDistort()
		} else {
			// fmt.Println("Using normalDistort")
			normalDistort()
		}

		updateDistortionState()

		sdl.CallQueue <- func() {
			currentFrameTime := sdl.GetTicks()
			renderer.Present()
			if currentFrameTime-lastFrameTime < frameTime {
				sdl.Delay(frameTime - (currentFrameTime - lastFrameTime))
			}
			lastFrameTime = currentFrameTime
		}
	}

	return 0
}

func main() {
	os.Exit(run())
}
