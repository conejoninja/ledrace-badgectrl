package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/touch/resistive"

	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"

	"tinygo.org/x/drivers/ili9341"
)

const (
	BLACK = iota
	PLAYER1
	PLAYER2
	PLAYER3
	PLAYER4
	BACKGROUND
	WHITE
	STEPL
	STEPR
)

var display = ili9341.NewParallel(
	machine.LCD_DATA0,
	machine.TFT_WR,
	machine.TFT_DC,
	machine.TFT_CS,
	machine.TFT_RESET,
	machine.TFT_RD,
)
var colors = []color.RGBA{
	color.RGBA{0, 0, 0, 255},       // BLACK
	color.RGBA{255, 0, 0, 255},     // PLAYER 1
	color.RGBA{0, 255, 0, 255},     // PLAYER 2
	color.RGBA{255, 255, 0, 255},   // PLAYER 3
	color.RGBA{0, 0, 255, 255},     // PLAYER 4
	color.RGBA{0,0,0, 255},    // BACKGROUND
	color.RGBA{255, 255, 255, 255}, // WHITE
	color.RGBA{102, 255, 51, 255},  // STEPL
	color.RGBA{255, 153, 51, 255},  // STEPR
}
var player int
var oldSpeed int16

var needlePoint = [91][2]int16{{64, 0}, {63, 1}, {63, 2}, {63, 3}, {63, 4}, {63, 5}, {63, 6}, {63, 7}, {63, 8}, {63, 10}, {63, 11}, {62, 12}, {62, 13}, {62, 14}, {62, 15}, {61, 16}, {61, 17}, {61, 18}, {60, 19}, {60, 20}, {60, 21}, {59, 22}, {59, 23}, {58, 25}, {58, 26}, {58, 27}, {57, 28}, {57, 29}, {56, 30}, {55, 31}, {55, 31}, {54, 32}, {54, 33}, {53, 34}, {53, 35}, {52, 36}, {51, 37}, {51, 38}, {50, 39}, {49, 40}, {49, 41}, {48, 41}, {47, 42}, {46, 43}, {46, 44}, {45, 45}, {44, 46}, {43, 46}, {42, 47}, {41, 48}, {41, 49}, {40, 49}, {39, 50}, {38, 51}, {37, 51}, {36, 52}, {35, 53}, {34, 53}, {33, 54}, {32, 54}, {32, 55}, {31, 55}, {30, 56}, {29, 57}, {28, 57}, {27, 58}, {26, 58}, {25, 58}, {23, 59}, {22, 59}, {21, 60}, {20, 60}, {19, 60}, {18, 61}, {17, 61}, {16, 61}, {15, 62}, {14, 62}, {13, 62}, {12, 62}, {11, 63}, {10, 63}, {8, 63}, {7, 63}, {6, 63}, {5, 63}, {4, 63}, {3, 63}, {2, 63}, {1, 63}, {0, 64}}

func main() {
	time.Sleep(3 * time.Second)
	machine.InitADC()
	resistiveTouch.Configure(&resistive.FourWireConfig{
		YP: machine.TOUCH_YD, // y+
		YM: machine.TOUCH_YU, // y-
		XP: machine.TOUCH_XR, // x+
		XM: machine.TOUCH_XL, // x-
	})

	machine.TFT_BACKLIGHT.Configure(machine.PinConfig{machine.PinOutput})

	display.Configure(ili9341.Config{})
	display.SetRotation(ili9341.Rotation270)

	display.FillScreen(colors[PLAYER2])
	machine.TFT_BACKLIGHT.High()

	player = menu()

	configureWifi(player)

	resetDisplay()
	// Both progress bar are 0-100 (0 started lap or race, 100 lap or race completed)
	// resetLapBar reset the lap bar for a new lap
	progressLapBar(80)
	progressRaceBar(60)
	// speedGaugeNeedle is 0-250
	speedGaugeNeedle(0, colors[PLAYER1])

	// STEPS are true|false if they are activated or not
	stepL(true)
	stepR(true)

	for {
		// CODE THAT READ SENSOR AND SEND MQTT MSG
		Send([]byte("1"))
		time.Sleep(100 * time.Millisecond)
	}

}

func resetDisplay() {
	display.FillScreen(colors[BACKGROUND])

	// GAUGE
	speedGauge()
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 67, 136, []byte("SPEED"), colors[WHITE])

	// STEP L
	tinydraw.Rectangle(display, 180, 100, 60, 60, colors[WHITE])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 184, 98, []byte("LEFT"), colors[WHITE])

	// STEP R
	tinydraw.Rectangle(display, 244, 100, 60, 60, colors[WHITE])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 248, 98, []byte("RIGHT"), colors[WHITE])

	// LAP PROGRESS BAR
	tinydraw.Rectangle(display, 8, 178, 304, 18, colors[WHITE])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 12, 175, []byte("LAP"), colors[WHITE])
	// RACE PROGRESS BAR
	tinydraw.Rectangle(display, 8, 218, 304, 18, colors[WHITE])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 12, 215, []byte("RACE"), colors[WHITE])
}

func progressLapBar(progress int16) {
	if progress > 300 {
		progress = 300
	}
	if progress < 0 {
		progress = 0
	}
	display.FillRectangle(10, 180, progress, 14, colors[player])
}

func resetLapBar() {
	display.FillRectangle(10, 180, 300, 14, colors[BACKGROUND])
}

func progressRaceBar(progress int16) {
	if progress > 300 {
		progress = 300
	}
	if progress < 0 {
		progress = 0
	}
	display.FillRectangle(10, 220, progress, 14, colors[player])
}

func speedGauge() {
	tinydraw.FilledCircle(display, 80, 90, 70, colors[WHITE])
	tinydraw.FilledCircle(display, 80, 90, 66, colors[BACKGROUND])
	tinydraw.FilledTriangle(display, 80, 90, 0, 160, 160, 160, colors[BACKGROUND])
}

func speedGaugeNeedle(speed int16, c color.RGBA) {
	speed -= 35
	if speed < 0 {
		speed -= 2 * speed
		tinydraw.Line(display, 80-needlePoint[speed][0], 90+needlePoint[speed][1], 79, 89, c)
		tinydraw.Line(display, 80-needlePoint[speed][0], 90+needlePoint[speed][1], 80, 90, c)
		tinydraw.Line(display, 80-needlePoint[speed][0], 90+needlePoint[speed][1], 81, 91, c)
	} else if speed >= 0 && speed <= 90 {
		tinydraw.Line(display, 80-needlePoint[speed][0], 90-needlePoint[speed][1], 79, 91, c)
		tinydraw.Line(display, 80-needlePoint[speed][0], 90-needlePoint[speed][1], 80, 90, c)
		tinydraw.Line(display, 80-needlePoint[speed][0], 90-needlePoint[speed][1], 81, 89, c)
	} else if speed > 90 && speed <= 180 {
		speed = 180 - speed
		tinydraw.Line(display, 80+needlePoint[speed][0], 90-needlePoint[speed][1], 79, 89, c)
		tinydraw.Line(display, 80+needlePoint[speed][0], 90-needlePoint[speed][1], 80, 90, c)
		tinydraw.Line(display, 80+needlePoint[speed][0], 90-needlePoint[speed][1], 81, 91, c)
	} else {
		if speed > 250 {
			speed = 250
		}
		speed -= 180
		tinydraw.Line(display, 80+needlePoint[speed][0], 90+needlePoint[speed][1], 79, 91, c)
		tinydraw.Line(display, 80+needlePoint[speed][0], 90+needlePoint[speed][1], 80, 90, c)
		tinydraw.Line(display, 80+needlePoint[speed][0], 90+needlePoint[speed][1], 81, 89, c)
	}
}

func stepL(enabled bool) {
	if enabled {
		display.FillRectangle(182, 102, 56, 56, colors[STEPL])
	} else {
		display.FillRectangle(182, 102, 56, 56, colors[BACKGROUND])
	}
}

func stepR(enabled bool) {
	if enabled {
		display.FillRectangle(246, 102, 56, 56, colors[STEPR])
	} else {
		display.FillRectangle(246, 102, 56, 56, colors[BACKGROUND])
	}
}
