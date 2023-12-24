package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
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

func Draw2DPoints(points []Point) {
	// we're basically just drawing a 256x256 image
	// Find the correct pixel size that will fill the window
	pixelSize := math.Min(float64(rl.GetRenderWidth()-xMargin*2), float64(rl.GetRenderHeight()-yMargin*2)) / 256

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

func Draw3DPoints(points []Point) {
	// TODO
	fmt.Println("3D")
}

func Draw4DPoints(points []Point) {
	// TODO
	fmt.Println("4D")
}

func main() {
	// the first argument is the file name
	// TODO proper arg handling
	lines := ReadLines(os.Args[1])
	ParseVizFileHeader(lines)
	points := ParseVizFilePoints(lines)
	fmt.Printf("%v\n", points[0])
	fmt.Println(len(points))
	// Raylib initialisation
	var drawerFunction func([]Point)
	switch dimension {
	case 2:
		drawerFunction = Draw2DPoints
	case 3:
		drawerFunction = Draw3DPoints
	case 4:
		drawerFunction = Draw4DPoints
	default:
		fmt.Println("what the fuck")
	}
	rl.InitWindow(screenWidth, screenHeight, "binvizualiser")
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		drawerFunction(points)
		rl.EndDrawing()
	}
}
