package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Point struct {
	Coordinate []uint8
	Value      uint8
}

const MAJOR_VER = "0"
const X_MARGIN = 20
const Y_MARGIN = 20
const CUBE_SIZE float32 = 0.1
const MIN_BRIGHTNESS int = 10
const MAX_POINTS = 100000
const DEBUG_MODE = true

var zoomSquareOnScreen bool = false
var isZoomed bool = false
var wValue uint8 = 0
var dimension int

func DrawDebugText(pointCount int) {
	// NOTE: don't forget to add newlines
	rl.DrawFPS(10, 10)
	debugText := fmt.Sprintf("Dimension: %v\n", dimension)
	debugText += fmt.Sprintf("Number of points: %v\n", pointCount)
	if dimension == 2 {
		debugText += fmt.Sprintf("Zoomed: %v\n", zoomSquareOnScreen)
	}
	rl.DrawText(debugText, 10, 30, 15, rl.White)
}

func ParseVizFileHeader(lines []string) {
	// format:VERSION maj.min.patch (semver)
	version := strings.Split(lines[0], " ")[1]
	fileMajorVersion := strings.Split(version, ".")[0]
	// check if the major version is the same
	if fileMajorVersion != MAJOR_VER {
		log.Printf("Major version mismatch between %v and %v, probably won't work how you think it will, you have been warned", fileMajorVersion, MAJOR_VER)
	}

	dim, err := strconv.ParseInt(strings.Split(lines[1], " ")[1], 10, 8)
	if err != nil {
		fmt.Println("Failed to parse dimension in file header")
		fmt.Println(err)
	}
	if dim < 1 || dim > 5 {
		fmt.Println("Dimension must be from 2 to 4")
	}
	dimension = int(dim)
}

func ParseVizFilePoints(lines []string) []Point {
	points := make([]Point, len(lines)-2)
	// i := 2 because first two lines are header
	for i := 2; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		coords := strings.Split(lines[i], " ")
		// parse coordinates
		coordinates := make([]uint8, dimension)
		for j := 0; j < dimension; j++ {
			coord, err := strconv.ParseUint(coords[j], 10, 8)
			if err != nil {
				fmt.Println("Failed to parse coordinate on line " + strconv.Itoa(i) + "")
				fmt.Println(err)
			}
			coordinates[j] = uint8(coord)
		}
		// parse value (last element)
		val, err := strconv.ParseUint(coords[dimension], 10, 8)
		if err != nil {
			fmt.Println("Failed to parse value")
			fmt.Println(err)
		}
		points[i-2] = Point{coordinates, uint8(val)}
	}
	// TODO range checks
	return points
}

func FilterPoints(points []Point) []Point {
	// if the value is less than 20, delete it
	// TODO make this adjustable
	filteredPoints := make([]Point, 0)
	for _, point := range points {
		if int(point.Value) > MIN_BRIGHTNESS {
			filteredPoints = append(filteredPoints, point)
		}
	}
	// sort the points by value
	sort.Slice(filteredPoints, func(a, b int) bool {
		return filteredPoints[a].Value > filteredPoints[b].Value
	})
	// turn the values into colours
	// return the top maxPoints points
	//
	// remove empty points
	for i := 0; i < len(filteredPoints); i++ {
		if len(filteredPoints[i].Coordinate) == 0 {
			filteredPoints = append(filteredPoints[:i], filteredPoints[i+1:]...)
		}
	}
	return filteredPoints[:int(math.Min(float64(MAX_POINTS), float64(len(filteredPoints))))]
}

func ReadLines(filename string) []string {
	f, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Failed to open file")
		fmt.Println(err)
	}
	return strings.Split(string(f), "\n")
}

func Draw2D(points *[]Point) {
	// we're basically just drawing a 256x256 image
	// Find the correct pixel size that will fill the window
	pixelSize := 4
	var zoomStart rl.Vector2 = rl.Vector2{X: 0, Y: 0}
	var zoomEnd rl.Vector2 = rl.Vector2{X: 0, Y: 0}
	// TODO: center the image
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// draw the points
		for _, point := range *points {
			x := X_MARGIN + (int32(point.Coordinate[0]) * int32(pixelSize)) + int32(rl.GetScreenWidth()/2) - int32(256*int32(pixelSize)/2)
			y := Y_MARGIN + (int32(point.Coordinate[1]) * int32(pixelSize)) + int32(rl.GetScreenHeight()/2) - int32(256*int32(pixelSize)/2)
			if isZoomed {
				// adjust them to be in the zoomed area
				windowWidth := math.Abs(float64(zoomEnd.X - zoomStart.X))
				windowHeight := math.Abs(float64(zoomEnd.Y - zoomStart.Y))
				// adjust the x and y to be relative to the zoom start
				x -= int32(zoomStart.X)
				y -= int32(zoomStart.Y)
				// scale them to be in the zoomed area
				x = int32(float64(x) * (256 / windowWidth))
				y = int32(float64(y) * (256 / windowHeight))
				// adjust them to be relative to the screen
				x += int32(zoomStart.X)
				y += int32(zoomStart.Y)
				pixelSize = int(float64(pixelSize) * (256 / windowWidth))
			}
			// centre this in the screen
			rl.DrawRectangle(
				x,                // X
				y,                // Y
				int32(pixelSize), // Width
				int32(pixelSize), // Height
				rl.NewColor(255, 255, 255, point.Value),
			)
			pixelSize = 4
		}
		// if the user stops right clicking, zoom in and set zoom end to their mouse pos
		zoomSquareOnScreen = rl.IsKeyDown(rl.KeyQ)
		if rl.IsKeyPressed(rl.KeyQ) {
			zoomStart = rl.GetMousePosition()
			isZoomed = false
		}
		if rl.IsKeyReleased(rl.KeyQ) {
			zoomEnd = rl.GetMousePosition()
			if zoomStart != zoomEnd {
				isZoomed = true
			}
		}
		if zoomSquareOnScreen {
			rl.DrawRectangleLines(
				int32(zoomStart.X),
				int32(zoomStart.Y),
				int32(rl.GetMousePosition().X)-int32(zoomStart.X),
				int32(rl.GetMousePosition().Y)-int32(zoomStart.Y),
				rl.Red)
		}
		if DEBUG_MODE {
			DrawDebugText(len(*points))
		}
		rl.EndDrawing()
	}
}

func Draw3D(points *[]Point) {
	// Init camera
	camera := rl.Camera3D{}
	camera.Position = rl.NewVector3(10.0, 10.0, 10.0)
	camera.Target = rl.NewVector3(CUBE_SIZE*128, CUBE_SIZE*128, CUBE_SIZE*128)
	camera.Up = rl.NewVector3(0.0, -1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	for !rl.WindowShouldClose() {
		rl.HideCursor()
		rl.ClearBackground(rl.Black)
		rl.UpdateCamera(&camera, rl.CameraFree)
		rl.BeginDrawing()
		rl.BeginMode3D(camera)
		for _, point := range *points {
			rl.DrawCube(
				rl.NewVector3(float32(point.Coordinate[0])*CUBE_SIZE, float32(point.Coordinate[1])*CUBE_SIZE, float32(point.Coordinate[2])*CUBE_SIZE),
				CUBE_SIZE,
				CUBE_SIZE,
				CUBE_SIZE,
				rl.NewColor(point.Value, point.Value, point.Value, point.Value),
			)
		}
		rl.EndMode3D()
		rl.DrawFPS(10, 10)
		// draw debug info some text
		if DEBUG_MODE {
			DrawDebugText(len(*points))
		}
		rl.EndDrawing()
	}
}

func Draw4D(points *[]Point) {
	// TODO fix this at parse-time, I think it's something to do with the newline at the end
	for _, point := range *points {
		if point.Coordinate[3] != wValue {
			continue
		}
		rl.DrawCube(
			rl.NewVector3(float32(point.Coordinate[0])*CUBE_SIZE, float32(point.Coordinate[1])*CUBE_SIZE, float32(point.Coordinate[2])*CUBE_SIZE),
			CUBE_SIZE,
			CUBE_SIZE,
			CUBE_SIZE,
			rl.NewColor(point.Value, point.Value, point.Value, 255),
		)
	}
	wValue = uint8(rg.SliderBar(rl.NewRectangle(20.0, 20.0, 100, 20.0), "W=0", "W=255", 0.0, 0.0, 255.0))
}

func main() {
	// the first argument is the file name
	// TODO proper arg handling
	startRead := time.Now()
	lines := ReadLines(os.Args[1])
	fmt.Println("Read file in", time.Since(startRead))
	ParseVizFileHeader(lines)

	startParse := time.Now()

	points := ParseVizFilePoints(lines)
	points = FilterPoints(points)

	var drawerFunction func(*[]Point)
	// probably too many code paths here, might find a better way to do this
	switch dimension {
	case 2:
		drawerFunction = Draw2D
	case 3:
		drawerFunction = Draw3D
	case 4:
		drawerFunction = Draw4D
	}

	// Raylib initialisation
	rl.InitWindow(0, 0, "binvizualiser")
	rl.SetConfigFlags(rl.FlagBorderlessWindowedMode | rl.FlagWindowMousePassthrough | rl.FlagWindowUndecorated)
	rl.SetWindowState(rl.FlagBorderlessWindowedMode | rl.FlagWindowMousePassthrough | rl.FlagWindowUndecorated)
	rl.ToggleFullscreen()
	fmt.Println("Parsed file in", time.Since(startParse))
	fmt.Println("File:         ", os.Args[1])
	fmt.Println("Dimension:    ", dimension)
	fmt.Println("No. of points:", len(points))
	drawerFunction(&points)
	// TODO: common drawing function / event loop stuff
	// TODO: GUI for configuration of the visualisation (dimension, max points, etc.)
}
