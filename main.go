package main

import (
	"math"
	"math/rand"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	size = 256
)

var (
	bg, fg          uint32
	out, surf       *sdl.Surface
	sand, sbuf, vis [][]bool
	ord             []int
	brush           int32 = 8
)

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	win, err := sdl.CreateWindow("Sand", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 1024, 1024, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer win.Destroy()

	surf, err = win.GetSurface()
	if err != nil {
		panic(err)
	}

	out, err = sdl.CreateRGBSurface(0, size, size, 24, 0, 0, 0, 0)
	if err != nil {
		panic(err)
	}
	defer out.Free()

	bg = sdl.MapRGB(out.Format, 230, 241, 254)
	fg = sdl.MapRGB(out.Format, 221, 132, 59)

	sand = make([][]bool, size)
	sbuf = make([][]bool, size)
	vis = make([][]bool, size)
	ord = make([]int, size)

	for i := 0; i < size; i++ {
		ord[i] = -1
	}

	for i := 0; i < size; i++ {
		sand[i] = make([]bool, size)
		sbuf[i] = make([]bool, size)
		vis[i] = make([]bool, size)

		var j int
		for j = rand.Intn(size); has(ord, j); j = rand.Intn(size) {
		}

		ord[i] = j
	}

	for y := 10; y < 90; y++ {
		for x := 0; x < size; x++ {
			if rand.Float64() < 0.4 {
				sand[y][x] = true
			}
		}
	}

	sdl.GLSetSwapInterval(1)

	running := true
	for running {
		for evt := sdl.PollEvent(); evt != nil; evt = sdl.PollEvent() {
			switch evt.(type) {
			case *sdl.QuitEvent:
				running = false
				break
			case *sdl.MouseWheelEvent:
				brush += int32(evt.(*sdl.MouseWheelEvent).Y)
				if brush <= 0 {
					brush = 1
				}

				break
			}
		}

		wMx, wMy, ms := sdl.GetMouseState()
		if ms&sdl.BUTTON_LEFT > 0 {
			mx, my := wMx*size/1024, wMy*size/1024

			for i := mx - brush; i < mx+brush; i++ {
				for j := my - brush; j < my+brush; j++ {
					if i < 0 || j < 0 || i >= size || j >= size || (i-mx)*(i-mx)+(j-my)*(j-my) >= brush*brush {
						continue
					}

					sand[j][i] = true
				}
			}
		}

		update()

		out.FillRect(nil, bg)
		render(out)

		out.BlitScaled(nil, surf, nil)
		win.UpdateSurface()
	}
}

func has(xs []int, n int) bool {
	for i := 0; i < len(xs); i++ {
		if xs[i] == n {
			return true
		}
	}

	return false
}

func setPixel(s *sdl.Surface, x, y int, colour uint32) {
	var (
		w      = s.Bounds().Size().X
		offset = s.BytesPerPixel() * (y*w + x)
		pix    = s.Pixels()
	)

	for i := 0; i < s.BytesPerPixel(); i++ {
		pix[offset+i] = byte(colour & 0xFF)
		colour >>= 8
	}
}

func render(s *sdl.Surface) {
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			if sand[y][x] {
				setPixel(s, x, y, fg)
			}
		}
	}

}

func update() {
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			sbuf[y][x] = false
			vis[y][x] = false
		}
	}

	// gravity
	for x := 0; x < size; x++ {
		for y := size - 1; y >= 0; y-- {
			if !sand[y][x] || vis[y][x] {
				continue
			}

			// emulate different speeds - sometimes, don't move by gravity
			if rand.Float64() < 0.4 {
				vis[y][x] = true
				sbuf[y][x] = true
				continue
			}

			for y2 := y; y2 >= 0 && sand[y2][x]; y2-- {
				vis[y2][x] = true

				if y == size-1 || sand[y+1][x] {
					// make no change if it's at the bottom
					sbuf[y2][x] = true
				} else {
					sbuf[y2+1][x] = true
				}
			}
		}
	}

	// slight pressure variations in air
	for x := 0; x < size; x++ {
		for y := 0; y < size-1; y++ {
			if !sbuf[y][x] || sbuf[y+1][x] {
				continue
			}

			r := rand.Float64()
			if r < 0.2 && x > 0 && !sbuf[y][x-1] {
				sbuf[y][x-1] = true
				sbuf[y][x] = false
			} else if r >= 0.8 && x < size-1 && !sbuf[y][x+1] {
				sbuf[y][x+1] = true
				sbuf[y][x] = false
			}
		}
	}

	// collapse towers
	for xx := 0; xx < size; xx++ {
		x := ord[xx]

		for y := 0; y < size-1; y++ {
			if sbuf[y][x] && sbuf[y+1][x] {

				var l, r, target int

				l = 0
				if x > 0 {
					for j := y; j < size && !sbuf[j][x-1]; j++ {
						l++
					}
				}

				r = 0
				if x+1 < size {
					for j := y; j < size && !sbuf[j][x+1]; j++ {
						r++
					}
				}

				max := maxSlopeAt(x, y)
				if l <= max && r <= max {
					continue
				}

				if l == 0 {
					target = 1
				} else if r == 0 {
					target = -1
				} else if rand.Float64() < 0.5 {
					target = 1
				} else {
					target = -1
				}

				if x+target >= 0 && x+target < size {
					sbuf[y][x] = false
					sbuf[y][x+target] = true
				}
			}
		}
	}

	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			sand[y][x] = sbuf[y][x]
		}
	}
}

func maxSlopeAt(x, y int) int {
	return int(10*math.Sqrt(10*math.Cos(float64(x)*153.123-float64(y)*15345.2333)+math.Sin(float64(x)*31223.5121+float64(y)*512563.123)))%2 + 1
}
