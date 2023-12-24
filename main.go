package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Point struct {
	Coordinate []uint8
	Value      uint8
}

var dimension int

const xMargin = 20
const yMargin = 20

var screenWidth int32 = 256*3 + xMargin*2
var screenHeight int32 = 256*3 + yMargin*2

const majorVersion = "0"

func ParseVizFileHeader(lines []string) {
	// format:VERSION maj.min.patch (semver)
	version := strings.Split(lines[0], " ")[1]
	fileMajorVersion := strings.Split(version, ".")[0]
	// check if the major version is the same
	if fileMajorVersion != majorVersion {
		log.Printf("Major version mismatch between %v and %v, probably won't work how you think it will, you have been warned", fileMajorVersion, majorVersion)
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

func ReadLines(filename string) []string {
	f, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Failed to open file")
		fmt.Println(err)
	}
	return strings.Split(string(f), "\n")
}

func Eventloop2D(drawerFunction func([]Point), points []Point) {
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		drawerFunction(points)
		rl.EndDrawing()
	}
}

func Draw2DPoints(points []Point) {
	// we're basically just drawing a 256x256 image
	// Find the correct pixel size that will fill the window
	pixelSize := math.Min(float64(screenWidth-xMargin*2), float64(screenWidth-yMargin*2)) / 256

	// draw the points
	for _, point := range points {
		if len(point.Coordinate) == 0 {
			continue
		}
		rl.DrawRectangle(
			xMargin+(int32(point.Coordinate[0])*int32(pixelSize)), // X
			yMargin+(int32(point.Coordinate[1])*int32(pixelSize)), // Y
			int32(pixelSize),                    // Width
			int32(pixelSize),                    // Height
			rl.NewColor(0, 255, 0, point.Value)) // Value
	}
}

func Eventloop3D(drawerFunction func([]Point), points []Point) {
	// init camera
	camera := rl.Camera3D{}
	camera.Position = rl.NewVector3(10.0, 10.0, 10.0)
	camera.Target = rl.NewVector3(0.0, 0.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective
	for !rl.WindowShouldClose() {
		rl.ClearBackground(rl.Black)
		rl.UpdateCamera(&camera, rl.CameraFree)
		rl.BeginDrawing()
		rl.BeginMode3D(camera)
		drawerFunction(points)
		rl.EndMode3D()
		rl.EndDrawing()
	}
}

func Draw3DPoints(points []Point) {
	for _, point := range points {
		// TODO fix this at parse-time, I think it's something to do with the newline at the end
		if len(point.Coordinate) == 0 {
			continue
		}
		// TODO: might want to make the cube size adjustable
		rl.DrawCube(
			rl.NewVector3(float32(point.Coordinate[0])/10, float32(point.Coordinate[1])/10, float32(point.Coordinate[2])/10),
			0.1,
			0.1,
			0.1,
			rl.NewColor(0, 255, 0, point.Value))
	}
}

func Draw4DPoints(points []Point) {
	// like drawing 3D but with a slider for the W axis (0..255) and it just draws the points with W = slider value
	// probably...
	fmt.Println("4D")
}

func main() {
	// the first argument is the file name
	// TODO proper arg handling
	lines := ReadLines(os.Args[1])
	ParseVizFileHeader(lines)
	points := ParseVizFilePoints(lines)
	fmt.Println("File:         ", os.Args[1])
	fmt.Println("Dimension:    ", dimension)
	fmt.Println("No. of points:", len(points))

	var drawerFunction func([]Point)
	var eventLoop func(func([]Point), []Point)
	// probably too many code paths here, might find a better way to do this
	switch dimension {
	case 2:
		eventLoop = Eventloop2D
		drawerFunction = Draw2DPoints
	case 3:
		eventLoop = Eventloop3D
		drawerFunction = Draw3DPoints
	case 4:
		eventLoop = Eventloop3D
		drawerFunction = Draw4DPoints
	default:
		fmt.Println("what the fuck")
	}
	// Raylib initialisation
	rl.InitWindow(screenWidth, screenHeight, "binvizualiser")
	eventLoop(drawerFunction, points)
}
