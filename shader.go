package main

import (
	"math"
	"sync"
)

const (
	Width  = 100
	Height = 100
)

// RGB holds a single pixel color.
type RGB struct {
	R, G, B byte
}

// RenderFrame computes a Width×Height frame for the given elapsed time,
// writing results into the provided framebuffer.
func RenderFrame(elapsed, mod1, mod2, mod3 float64, frame *[Height][Width]RGB) {
	rx, ry := float64(Width), float64(Height)
	var wg sync.WaitGroup

	for row := range Height {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for col := range Width {
				fcx := float64(col) + 0.5
				fcy := float64(Height-1-row) + 0.5

				px := (fcx*2.0 - rx) / ry
				py := (fcy*2.0 - ry) / ry

				lval := math.Abs(mod1 - (px*px + py*py))

				scale := (1.0 - lval) / 0.2
				vx := px * scale
				vy := py * scale

				var channels [3]float64

				for iter := 1.0; iter <= mod2; iter++ {
					dvx := math.Cos(vy*iter+elapsed)/iter + mod1
					dvy := math.Cos(vx*iter+iter+elapsed)/iter + mod1
					vx += dvx
					vy += dvy

					brightness := math.Abs(vx-vy) * mod3
					channels[0] += (math.Sin(vx) + 1.0) * brightness
					channels[1] += (math.Sin(vy) + 1.0) * brightness
					channels[2] += (math.Sin(vy) + 1.0) * brightness
				}

				el := math.Exp(-4.0 * lval)
				channels[0] = math.Tanh(math.Exp(py*1.0) * el / channels[0])
				channels[1] = math.Tanh(math.Exp(py*-1.0) * el / channels[1])
				channels[2] = math.Tanh(math.Exp(py*-2.0) * el / channels[2])

				frame[row][col] = RGB{
					R: clampByte(channels[0]),
					G: clampByte(channels[1]),
					B: clampByte(channels[2]),
				}
			}
		}()
	}
	wg.Wait()
}

func clampByte(val float64) byte {
	if val < 0 {
		val = 0
	}
	if val > 1 {
		val = 1
	}
	return byte(val * 255)
}
