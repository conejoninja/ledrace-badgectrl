package main

import (
	"image/color"

	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"

	"tinygo.org/x/drivers/touch"
	"tinygo.org/x/drivers/touch/resistive"
)

var (
	resistiveTouch = new(resistive.FourWire)
)

const (
	Xmin = 750
	Xmax = 325
	Ymin = 840
	Ymax = 240
)

func menu() int {
	display.FillScreen(color.RGBA{0, 0, 0, 255})

	display.FillRectangle(0, 0, 160, 120, colors[PLAYER1])
	display.FillRectangle(160, 0, 160, 120, colors[PLAYER2])
	display.FillRectangle(0, 120, 160, 120, colors[PLAYER3])
	display.FillRectangle(160, 120, 160, 120, colors[PLAYER4])

	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 60, 100, []byte("SELECT PLAYER."), color.RGBA{0, 0, 0, 255})

	last := touch.Point{}

	// loop and poll for touches, including performing debouncing
	debounce := 0
	for {
		point := resistiveTouch.ReadTouchPoint()
		touch := touch.Point{}
		if point.Z>>6 > 100 {
			touch.X = mapval(point.X>>6, Xmin, Xmax, 0, 240)
			touch.Y = mapval(point.Y>>6, Ymin, Ymax, 0, 320)
			touch.Z = point.Z >> 6 / 100
		} else {
			touch.X = 0
			touch.Y = 0
			touch.Z = 0
		}

		if last.Z != touch.Z {
			debounce = 0
			last = touch
		} else if (touch.X-last.X) > 4 ||
			(touch.Y-last.Y) > 4 ||
			(touch.X-last.X) < -4 ||
			(touch.Y-last.Y) < -4 {
			debounce = 0
			last = touch
		} else if debounce > 1 {
			debounce = 0
			//HandleTouch(last)
			if touch.X < 120 {
				if touch.Y < 160 {
					return PLAYER2
				} else {
					return PLAYER1
				}
			} else {
				if touch.Y < 160 {
					return PLAYER4
				} else {
					return PLAYER3
				}
			}
		} else if touch.Z > 0 {
			debounce++
		} else {
			last = touch
			debounce = 0
		}

	}
	return PLAYER1
}

// based on Arduino's "map" function
func mapval(x int, inMin int, inMax int, outMin int, outMax int) int {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}

func HandleTouch(touch touch.Point) {
	println("touch point:", touch.X, touch.Y, touch.Z)
}
