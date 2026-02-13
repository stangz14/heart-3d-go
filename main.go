package main

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

var colors = []string{
	"\033[31m", "\033[91m", "\033[95m", "\033[35m",
}

type Point3D struct {
	x, y, z float64
}

var heartPoints []Point3D

const (
	width  = 60
	height = 30
)

func init() {
	for t := 0.0; t < 2*math.Pi; t += 0.1 {
		for z := -10.0; z <= 10.0; z += 0.5 {
			x := 16 * math.Pow(math.Sin(t), 3)
			y := -(13*math.Cos(t) - 5*math.Cos(2*t) - 2*math.Cos(3*t) - math.Cos(4*t))
			scale := 0.6
			r := (15 - math.Abs(z)) / 15
			heartPoints = append(heartPoints, Point3D{
				x: x * scale * r,
				y: y * scale * r,
				z: z * scale,
			})
		}
	}
}

func renderFrame(angle float64) string {
	buffer := make([][]string, height)
	for i := range buffer {
		buffer[i] = make([]string, width)
		for j := range buffer[i] {
			buffer[i][j] = " "
		}
	}

	zBuffer := make([][]float64, height)
	for i := range zBuffer {
		zBuffer[i] = make([]float64, width)
		for j := range zBuffer[i] {
			zBuffer[i][j] = -math.MaxFloat64
		}
	}

	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	for _, p := range heartPoints {
		rotX := p.x*cosA - p.z*sinA
		rotZ := p.x*sinA + p.z*cosA
		rotY := p.y

		screenX := int(width/2 + rotX)
		screenY := int(height/2 + rotY*0.5)

		if screenX >= 0 && screenX < width && screenY >= 0 && screenY < height {
			if rotZ > zBuffer[screenY][screenX] {
				zBuffer[screenY][screenX] = rotZ
				char := "*"
				if rotZ > 5 {
					char = "@"
				} else if rotZ < -5 {
					char = "."
				}
				buffer[screenY][screenX] = char
			}
		}
	}

	var sb strings.Builder
	for _, row := range buffer {
		for _, char := range row {
			sb.WriteString(char)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Connection", "keep-alive")

		angle := 0.0
		colorIndex := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				frame := renderFrame(angle)
				output := fmt.Sprintf("\033[2J\033[H\033[3J%s%s\033[0m", colors[colorIndex], frame)

				fmt.Fprint(w, output)
				flusher.Flush()

				angle += 0.2
				if angle > 2*math.Pi {
					angle -= 2 * math.Pi
				}

				if int(angle*10)%5 == 0 {
					colorIndex = (colorIndex + 1) % len(colors)
				}
			}
		}
	})

	fmt.Println("Rotating Heart server is running ")
	http.ListenAndServe(":8080", nil)
}
