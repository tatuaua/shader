package shader

import "math"

const (
	Width  = 50
	Height = 50
)

// RGB holds a single pixel color.
type RGB struct {
	R, G, B byte
}

// RenderFrame computes a Width×Height frame for the given time t.
func RenderFrame(t float64) []RGB {
	frame := make([]RGB, Width*Height)

	rx, ry := float64(Width), float64(Height)

	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			fcx := float64(x) + 0.5
			fcy := float64(Height-1-y) + 0.5

			px := (fcx*2.0 - rx) / ry
			py := (fcy*2.0 - ry) / ry

			lval := math.Abs(0.7 - (px*px + py*py))

			scale := (1.0 - lval) / 0.2
			vx := px * scale
			vy := py * scale

			var o [4]float64

			for i := 1.0; i <= 8.0; i++ {
				dvx := math.Cos(vy*i+t)/i + 0.7
				dvy := math.Cos(vx*i+i+t)/i + 0.7
				vx += dvx
				vy += dvy

				s := math.Abs(vx-vy) * 0.2
				o[0] += (math.Sin(vx) + 1.0) * s
				o[1] += (math.Sin(vy) + 1.0) * s
				o[2] += (math.Sin(vy) + 1.0) * s
				o[3] += (math.Sin(vx) + 1.0) * s
			}

			el := math.Exp(-4.0 * lval)
			o[0] = math.Tanh(math.Exp(py*1.0) * el / o[0])
			o[1] = math.Tanh(math.Exp(py*-1.0) * el / o[1])
			o[2] = math.Tanh(math.Exp(py*-2.0) * el / o[2])

			frame[y*Width+x] = RGB{
				R: clampByte(o[0]),
				G: clampByte(o[1]),
				B: clampByte(o[2]),
			}
		}
	}

	return frame
}

func clampByte(v float64) byte {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return byte(v * 255)
}

func WrapSlice(flat []RGB) [][]RGB {
	grid := make([][]RGB, Height)
	for i := range grid {
		grid[i] = flat[i*Width : (i+1)*Width]
	}
	return grid
}
