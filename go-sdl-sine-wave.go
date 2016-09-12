package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"os"
	"sync"
)

type Point struct {
	x int
	y int
}

type Color struct {
	r uint8
	g uint8
	b uint8
	a uint8
}

type Pixel struct {
	point Point
	color Color
}

var winTitle string = "Go SDL2 Sine Wave"
var winWidth, winHeight int = 1280, 720

var lastFrameTick uint32 = 0

var origGridPoints []sdl.Point = createGrid(winWidth, winHeight)
var gridPoints []sdl.Point = createGrid(winWidth, winHeight)
var drawMode = 0
var updateMode = 0

func RoundToInt32(a float64) int32 {
	if a < 0 {
		return int32(a - 0.5)
	}

	return int32(a + 0.5)
}

func draw(window *sdl.Window, renderer *sdl.Renderer) {
	const fps int = 90
	const ticksPerFrame uint32 = uint32(1000 / fps)
	bgColor := Color{r: 0, b: 0, g: 0, a: 255}
	var currentFrameTick uint32 = sdl.GetTicks()
	if lastFrameTick == 0 {
		lastFrameTick = sdl.GetTicks()
	}

	renderer.SetDrawColor(
		bgColor.r,
		bgColor.g,
		bgColor.b,
		bgColor.a,
	)

	renderer.Clear()

	if updateMode == 0 {
		updateGrid_1()
	} else {
		updateGrid()
	}

	if drawMode == 0 {
		drawGrid(gridPoints, renderer)
	} else {
		drawGrid_1(gridPoints, renderer)
	}

	renderer.Present()

	if (currentFrameTick - lastFrameTick) >= ticksPerFrame {
		lastFrameTick = currentFrameTick
	} else {
		var elapsedTicks uint32 = currentFrameTick - lastFrameTick
		sdl.Delay(ticksPerFrame - elapsedTicks)
	}
}

var amp float64 = 50
var tx float32 = math.Pi / 9
var ty float32 = math.Pi / 4
var xFreq float32 = 1
var yFreq float32 = 1
var xFreqDir float32 = 1
var yFreqDir float32 = 1
var lastUpdateTicks uint32 = 0

func sineWaveDistortXY(x int32, y int32, w int, h int) (int32, int32) {
	var distortedX int32 = x
	var distortedY int32 = y
	var updateTime uint32 = 10
	var currentTicks uint32 = sdl.GetTicks()
	if lastUpdateTicks == 0 {
		lastUpdateTicks = sdl.GetTicks()
	}

	var normalizedX float32 = float32(x) / float32(w)
	var normalizedY float32 = float32(y) / float32(h)

	var xOffset = int32(amp * (math.Sin(float64(xFreq*normalizedY+yFreq*normalizedX+2*math.Pi*tx)) * 0.5))
	var yOffset = int32(amp * (math.Sin(float64(xFreq*normalizedY+yFreq*normalizedX+2*math.Pi*ty)) * 0.5))

	//fmt.Println(xOffset, yOffset)
	distortedX += xOffset
	distortedY += yOffset

	if (currentTicks - lastUpdateTicks) >= updateTime {
		xFreq += (0.1) * xFreqDir
		if xFreq > 25 || xFreq < 1 {
			xFreqDir *= -1
		}
		yFreq += (0.1) * yFreqDir
		if yFreq > 30 || yFreq < 1 {
			yFreqDir *= -1
		}

		lastUpdateTicks = sdl.GetTicks()
	}

	return distortedX, distortedY
}

func createGrid(w int, h int) []sdl.Point {
	var cols int = 50
	var rows int = 50
	var pixelCount = (cols-1)*h + (rows-1)*w
	var points = make([]sdl.Point, pixelCount)
	var cellWidth int32 = RoundToInt32(float64(w) / float64(cols))
	var cellHeight int32 = RoundToInt32(float64(h) / float64(rows))
	var index = 0

	// Create Columns
	for i := 1; i < cols; i++ {
		x := i * int(cellWidth)
		for j := 0; j < h; j++ {
			points[index] = sdl.Point{X: int32(x), Y: int32(j)}
			index++
		}
	}

	// Create Rows
	for i := 1; i < cols; i++ {
		y := i * int(cellHeight)
		for j := 0; j < w; j++ {
			points[index] = sdl.Point{X: int32(j), Y: int32(y)}
			index++
		}
	}

	return points
}

func updateGrid_1() {
	for i, point := range origGridPoints {
		newX, newY := sineWaveDistortXY(point.X, point.Y, winWidth, winHeight)
		gridPoints[i] = sdl.Point{X: newX, Y: newY}
	}
}

func updateGrid() {
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}
	for i, point := range origGridPoints {
		wg.Add(1)
		go func(i int, point sdl.Point) {
			newX, newY := sineWaveDistortXY(point.X, point.Y, winWidth, winHeight)
			mutex.Lock()
			gridPoints[i] = sdl.Point{X: newX, Y: newY}
			mutex.Unlock()
			wg.Done()
			fmt.Println("done")
		}(i, point)
	}
	fmt.Println("Waiting")
	wg.Wait()
	fmt.Println("All done")
}

func drawGrid(gridPoints []sdl.Point, renderer *sdl.Renderer) {
	var gridColor Color = Color{r: 0, g: 255, b: 0, a: 255}
	renderer.SetDrawColor(
		gridColor.r,
		gridColor.g,
		gridColor.b,
		gridColor.a,
	)

	renderer.DrawPoints(gridPoints)
}

func drawGrid_1(gridPoints []sdl.Point, renderer *sdl.Renderer) {
	var gridColor Color = Color{r: 0, g: 255, b: 0, a: 255}
	for _, point := range gridPoints {
		renderer.SetDrawColor(
			gridColor.r,
			gridColor.g,
			gridColor.b,
			gridColor.a,
		)

		renderer.DrawPoint(int(point.X), int(point.Y))
	}
}

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer

	var event sdl.Event
	var fullscreen bool
	var running bool

	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer renderer.Destroy()

	running = true
	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyUpEvent:
				if t.Keysym.Sym == sdl.K_q {
					running = false
				} else if t.Keysym.Sym == sdl.K_f {
					//fmt.Println("Fullscreen not working")
					if !fullscreen {
						// Go fullscreen
						fullscreen = true
						window.SetFullscreen(sdl.WINDOW_FULLSCREEN)
					} else {
						// Leave fullscreen
						fullscreen = false
						window.SetFullscreen(0)
					}
				} else if t.Keysym.Sym == sdl.K_m {
					if drawMode == 0 {
						drawMode = 1
					} else {
						drawMode = 0
					}
				} else if t.Keysym.Sym == sdl.K_u {
					if updateMode == 0 {
						updateMode = 1
					} else {
						updateMode = 0
					}
				}
			}
		}

		draw(window, renderer)
	}

	renderer.Destroy()
	window.Destroy()
	sdl.Quit()
	return 0
}

func main() {
	os.Exit(run())
}
