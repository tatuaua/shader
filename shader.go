package main

import (
	"math"
	"sync"
)

const (
	Width  = 50
	Height = 50
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

				var o [4]float64

				for i := 1.0; i <= mod2; i++ {
					dvx := math.Cos(vy*i+t)/i + mod1
					dvy := math.Cos(vx*i+i+t)/i + mod1
					vx += dvx
					vy += dvy

					s := math.Abs(vx-vy) * mod3
					o[0] += (math.Sin(vx) + 1.0) * s
					o[1] += (math.Sin(vy) + 1.0) * s
					o[2] += (math.Sin(vy) + 1.0) * s
					o[3] += (math.Sin(vx) + 1.0) * s
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
	var wg sync.WaitGroup

	for y := range Height {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			for x := range Width {
				fcx := float64(x) + 0.5
				fcy := float64(Height-1-row) + 0.5
				fcz := 0.5

				// Normalize direction: normalize(FC.rgb*2 - r.xyy)
				dx := fcx*2 - rx
				dy := fcy*2 - ry
				dz := fcz*2 - rx
				dlen := math.Sqrt(dx*dx + dy*dy + dz*dz)
				dx /= dlen
				dy /= dlen
				dz /= dlen

				var o [4]float64
				var z float64

				for i := 0.0; i < 60; i++ {
					// c = p = z * normalize(...)
					cx, cy, cz := z*dx, z*dy, z*dz
					px, py, pz := cx, cy, cz

					// p.z -= t / 0.2
					pz -= t / 0.2
					// c.z += 9
					cz += 9

					// Inner loop: accumulate sinusoidal displacement
					for f := 1.0; f <= 7; f++ {
						sx := math.Sin(px*f+z*0.2+1/l(cx, cy, cz, t)) / f
						sy := math.Sin(py*f+z*0.2+1/l(cx, cy, cz, t)) / f
						sz := math.Sin(pz*f+z*0.2+1/l(cx, cy, cz, t)) / f
						// p += sin(...).yzx / f
						px += sy
						py += sz
						pz += sx
					}

					// l = length(c + 4*sin(t + vec3(0,8,4)))
					lval := l(cx, cy, cz, t)

					// f = 8 - length((c+p).xy)
					cpx := cx + px
					cpy := cy + py
					fval := 8 - math.Sqrt(cpx*cpx+cpy*cpy)

					// min(max(f, -f*0.2), l)
					minVal := math.Min(math.Max(fval, -fval*0.2), lval)

					// z += f = 0.01 + minVal/7
					step := 0.01 + minVal/7
					z += step

					// o += vec4(5,1,l,1) / l / l / f / z
					if lval > 0.001 && step > 0.001 && z > 0.001 {
						denom := lval * lval * step * z
						o[0] += 5 / denom
						o[1] += 1 / denom
						o[2] += lval / denom
						o[3] += 1 / denom
					}
				}

				// o = tanh(o / 300)
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

// l computes length(c + 4*sin(t + vec3(0,8,4)))
func l(cx, cy, cz, t float64) float64 {
	sx := cx + 4*math.Sin(t+0)
	sy := cy + 4*math.Sin(t+8)
	sz := cz + 4*math.Sin(t+4)
	return math.Sqrt(sx*sx + sy*sy + sz*sz)
}
