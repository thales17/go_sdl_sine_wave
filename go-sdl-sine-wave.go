package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"math/rand"
	"os"
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

	/*for i := 0; i < 50000; i++ {
		drawPixel(randomPixel(winWidth, winHeight), renderer)
	}*/

	//oldDrawGrid(winWidth, winHeight, renderer)

	updateGrid()
	drawGrid(gridPoints, renderer)

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

func (p *Point) sineWaveDistortPoint(w int, h int) {
	var updateTime uint32 = 10
	var currentTicks uint32 = sdl.GetTicks()
	if lastUpdateTicks == 0 {
		lastUpdateTicks = sdl.GetTicks()
	}

	var normalizedX float32 = float32(p.x) / float32(w)
	var normalizedY float32 = float32(p.y) / float32(h)

	var xOffset = int(amp * (math.Sin(float64(xFreq*normalizedY+yFreq*normalizedX+2*math.Pi*tx)) * 0.5))
	var yOffset = int(amp * (math.Sin(float64(xFreq*normalizedY+yFreq*normalizedX+2*math.Pi*ty)) * 0.5))

	//fmt.Println(xOffset, yOffset)
	p.x += xOffset
	p.y += yOffset

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
}

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
	var cols int = 90
	var rows int = 90
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

func updateGrid() {
	for i, point := range origGridPoints {
		newX, newY := sineWaveDistortXY(point.X, point.Y, winWidth, winHeight)
		gridPoints[i] = sdl.Point{X: newX, Y: newY}
	}
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

func oldDrawGrid(w int, h int, renderer *sdl.Renderer) {
	var cols int = 30
	var rows int = 30
	var cellWidth int32 = RoundToInt32(float64(w) / float64(cols))
	var cellHeight int32 = RoundToInt32(float64(h) / float64(rows))
	var gridColor Color = Color{r: 0, g: 255, b: 0, a: 255}

	// Draw Columns
	for i := 1; i < cols; i++ {
		x := i * int(cellWidth)
		for j := 0; j < h; j++ {
			p := Point{x: x, y: j}
			p.sineWaveDistortPoint(w, h)
			drawPixel(Pixel{point: p, color: gridColor}, renderer)
		}
	}

	for i := 1; i < cols; i++ {
		y := i * int(cellHeight)
		for j := 0; j < w; j++ {
			p := Point{x: j, y: y}
			p.sineWaveDistortPoint(w, h)
			drawPixel(Pixel{point: p, color: gridColor}, renderer)
		}
	}
}

func drawPixel(pixel Pixel, renderer *sdl.Renderer) {
	renderer.SetDrawColor(
		pixel.color.r,
		pixel.color.g,
		pixel.color.b,
		pixel.color.a,
	)

	renderer.DrawPoint(pixel.point.x, pixel.point.y)
}

func randomPixel(w int, h int) Pixel {
	var pixel Pixel
	pixel.point.x = rand.Intn(w)
	pixel.point.y = rand.Intn(h)
	pixel.color.r = uint8(rand.Intn(255))
	pixel.color.g = uint8(rand.Intn(255))
	pixel.color.b = uint8(rand.Intn(255))
	pixel.color.a = uint8(rand.Intn(255))

	return pixel
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
