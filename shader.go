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

// RenderFrame computes a Width×Height frame for the given time t,
// writing results into the provided framebuffer.
func RenderFrame(t, mod1, mod2, mod3 float64, frame *[Height][Width]RGB) {
	rx, ry := float64(Width), float64(Height)
	var wg sync.WaitGroup

	for y := range Height {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			for x := range Width {
				fcx := float64(x) + 0.5
				fcy := float64(Height-1-y) + 0.5

				px := (fcx*2.0 - rx) / ry
				py := (fcy*2.0 - ry) / ry

				lval := math.Abs(mod1 - (px*px + py*py))

				scale := (1.0 - lval) / 0.2
				vx := px * scale
				vy := py * scale

				var o [3]float64

				for i := 1.0; i <= mod2; i++ {
					dvx := math.Cos(vy*i+t)/i + mod1
					dvy := math.Cos(vx*i+i+t)/i + mod1
					vx += dvx
					vy += dvy

					s := math.Abs(vx-vy) * mod3
					o[0] += (math.Sin(vx) + 1.0) * s
					o[1] += (math.Sin(vy) + 1.0) * s
					o[2] += (math.Sin(vy) + 1.0) * s
				}

				el := math.Exp(-4.0 * lval)
				o[0] = math.Tanh(math.Exp(py*1.0) * el / o[0])
				o[1] = math.Tanh(math.Exp(py*-1.0) * el / o[1])
				o[2] = math.Tanh(math.Exp(py*-2.0) * el / o[2])

				frame[y][x] = RGB{
					R: clampByte(o[0]),
					G: clampByte(o[1]),
					B: clampByte(o[2]),
				}
			}
		}(y)
	}
	wg.Wait()
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

// RenderFrame2 computes a Width×Height frame for the raymarching shader,
// writing results into the provided framebuffer.
func RenderFrame2(t float64, frame *[Height][Width]RGB) {
	rx, ry := float64(Width), float64(Height)

	// Pre-compute frame-constant values
	sinT0 := 4 * math.Sin(t)
	sinT8 := 4 * math.Sin(t+8)
	sinT4 := 4 * math.Sin(t+4)
	tDiv02 := t / 0.2
	dz := 1.0 - rx // fcz*2 - rx is constant (fcz = 0.5)

	var wg sync.WaitGroup

	for y := range Height {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			dy := (float64(Height-1-row)+0.5)*2 - ry
			dy2dz2 := dy*dy + dz*dz

			for x := range Width {
				dx := (float64(x)+0.5)*2 - rx
				dlen := math.Sqrt(dx*dx + dy2dz2)
				ndx := dx / dlen
				ndy := dy / dlen
				ndz := dz / dlen

				var o [3]float64
				var z float64

				for i := 0.0; i < 60; i++ {
					cx, cy, cz := z*ndx, z*ndy, z*ndz
					px, py, pz := cx, cy, cz

					pz -= tDiv02
					cz += 9

					// Compute l(cx, cy, cz, t) once
					lsx := cx + sinT0
					lsy := cy + sinT8
					lsz := cz + sinT4
					lval := math.Sqrt(lsx*lsx + lsy*lsy + lsz*lsz)

					// Pre-compute inner-loop invariant
					base := z*0.2 + 1.0/lval

					for f := 1.0; f <= 7; f++ {
						invF := 1.0 / f
						sx := math.Sin(px*f+base) * invF
						sy := math.Sin(py*f+base) * invF
						sz := math.Sin(pz*f+base) * invF
						px += sy
						py += sz
						pz += sx
					}

					cpx := cx + px
					cpy := cy + py
					fval := 8 - math.Sqrt(cpx*cpx+cpy*cpy)

					minVal := math.Min(math.Max(fval, -fval*0.2), lval)

					step := 0.01 + minVal/7
					z += step

					if lval > 0.001 && step > 0.001 && z > 0.001 {
						denom := lval * lval * step * z
						o[0] += 5 / denom
						o[1] += 1 / denom
						o[2] += lval / denom
					}
				}

				o[0] = math.Tanh(o[0] / 300)
				o[1] = math.Tanh(o[1] / 300)
				o[2] = math.Tanh(o[2] / 300)

				frame[row][x] = RGB{
					R: clampByte(o[0]),
					G: clampByte(o[1]),
					B: clampByte(o[2]),
				}
			}
		}(y)
	}
	wg.Wait()
}
