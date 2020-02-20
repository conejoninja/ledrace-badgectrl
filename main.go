package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/shifter"
	"tinygo.org/x/drivers/st7735"
	"tinygo.org/x/drivers/ws2812"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
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

var display st7735.Device
var buttons shifter.Device
var leds ws2812.Device
var colors []color.RGBA
var player uint8
var needlePoint = [91][2]int16{{30, 0}, {29, 0}, {29, 1}, {29, 1}, {29, 2}, {29, 2}, {29, 3}, {29, 3}, {29, 4}, {29, 4}, {29, 5}, {29, 5}, {29, 6}, {29, 6}, {29, 7}, {28, 7}, {28, 8}, {28, 8}, {28, 9}, {28, 9}, {28, 10}, {28, 10}, {27, 11}, {27, 11}, {27, 12}, {27, 12}, {26, 13}, {26, 13}, {26, 14}, {26, 14}, {25, 14}, {25, 15}, {25, 15}, {25, 16}, {24, 16}, {24, 17}, {24, 17}, {23, 18}, {23, 18}, {23, 18}, {22, 19}, {22, 19}, {22, 20}, {21, 20}, {21, 20}, {21, 21}, {20, 21}, {20, 21}, {20, 22}, {19, 22}, {19, 22}, {18, 23}, {18, 23}, {18, 23}, {17, 24}, {17, 24}, {16, 24}, {16, 25}, {15, 25}, {15, 25}, {15, 25}, {14, 26}, {14, 26}, {13, 26}, {13, 26}, {12, 27}, {12, 27}, {11, 27}, {11, 27}, {10, 28}, {10, 28}, {9, 28}, {9, 28}, {8, 28}, {8, 28}, {7, 28}, {7, 29}, {6, 29}, {6, 29}, {5, 29}, {5, 29}, {4, 29}, {4, 29}, {3, 29}, {3, 29}, {2, 29}, {2, 29}, {1, 29}, {1, 29}, {0, 29}, {0, 30}}

func main() {
	machine.SPI1.Configure(machine.SPIConfig{
		SCK:       machine.SPI1_SCK_PIN,
		MOSI:      machine.SPI1_MOSI_PIN,
		MISO:      machine.SPI1_MISO_PIN,
		Frequency: 8000000,
	})

	display = st7735.New(machine.SPI1, machine.TFT_RST, machine.TFT_DC, machine.TFT_CS, machine.TFT_LITE)
	display.Configure(st7735.Config{
		Rotation: st7735.ROTATION_90,
	})

	buttons = shifter.New(shifter.EIGHT_BITS, machine.BUTTON_LATCH, machine.BUTTON_CLK, machine.BUTTON_OUT)
	buttons.Configure()

	neo := machine.NEOPIXELS
	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})
	leds = ws2812.New(neo)

	colors = []color.RGBA{
		color.RGBA{0, 0, 0, 255},       // BLACK
		color.RGBA{255, 0, 0, 255},     // PLAYER 1
		color.RGBA{0, 255, 0, 255},     // PLAYER 2
		color.RGBA{255, 255, 0, 255},   // PLAYER 3
		color.RGBA{0, 0, 255, 255},     // PLAYER 4
		color.RGBA{50, 50, 50, 255},    // BACKGROUND
		color.RGBA{255, 255, 255, 255}, // WHITE
		color.RGBA{102, 255, 51, 255},  // STEPL
		color.RGBA{255, 153, 51, 255},  // STEPR
	}

	player = 1

	resetDisplay()
	// Both progress bar are 0-100 (0 started lap or race, 100 lap or race completed)
	// resetLapBar reset the lap bar for a new lap
	progressLapBar(80)
	progressRaceBar(60)
	// speedGaugeNeedle is 0-250
	speedGaugeNeedle(0, colors[BACKGROUND])

	// STEPS are true|false if they are activated or not
	stepL(true)
	stepR(true)

	var oldSpeed, speed, delta int16
	delta = 1

	for {
		speedGaugeNeedle(oldSpeed, colors[BACKGROUND])
		speedGaugeNeedle(speed, colors[PLAYER1])
		oldSpeed = speed
		speed += delta
		if speed >= 250 {
			delta = -1
		}
		if speed <= 0 {
			delta = 1
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func resetDisplay() {
	display.FillScreen(colors[BACKGROUND])

	// GAUGE
	speedGauge()
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 37, 76, []byte("SPEED"), colors[WHITE])

	// STEP L
	tinydraw.Rectangle(&display, 108, 30, 20, 20, colors[WHITE])
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 116, 26, []byte("L"), colors[WHITE])

	// STEP R
	tinydraw.Rectangle(&display, 132, 30, 20, 20, colors[WHITE])
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 140, 26, []byte("R"), colors[WHITE])

	// LAP PROGRESS BAR
	tinydraw.Rectangle(&display, 8, 88, 144, 8, colors[WHITE])
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 12, 86, []byte("LAP"), colors[WHITE])
	// RACE PROGRESS BAR
	tinydraw.Rectangle(&display, 8, 108, 144, 8, colors[WHITE])
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 12, 106, []byte("RACE"), colors[WHITE])
}

func progressLapBar(progress float32) {
	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}
	display.FillRectangle(10, 90, int16(1.4*progress), 4, colors[player])
}

func resetLapBar() {
	display.FillRectangle(10, 90, 140, 4, colors[BACKGROUND])
}

func progressRaceBar(progress float32) {
	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}
	display.FillRectangle(10, 110, int16(1.4*progress), 4, colors[player])
}

func speedGauge() {
	tinydraw.FilledCircle(&display, 50, 50, 40, colors[WHITE])
	tinydraw.FilledCircle(&display, 50, 50, 38, colors[BACKGROUND])
	tinydraw.FilledTriangle(&display, 50, 50, 0, 90, 100, 90, colors[BACKGROUND])
}

func speedGaugeNeedle(speed int16, c color.RGBA) {
	speed -= 35
	if speed < 0 {
		speed -= 2 * speed
		tinydraw.Line(&display, 50-needlePoint[speed][0], 50+needlePoint[speed][1], 50, 50, c)
		tinydraw.Line(&display, 50-needlePoint[speed][0], 51+needlePoint[speed][1], 50, 51, c)
	} else if speed >= 0 && speed <= 90 {
		tinydraw.Line(&display, 50-needlePoint[speed][0], 50-needlePoint[speed][1], 50, 50, c)
		tinydraw.Line(&display, 50-needlePoint[speed][0], 51-needlePoint[speed][1], 50, 51, c)
	} else if speed > 90 && speed <= 180 {
		speed = 180 - speed
		tinydraw.Line(&display, 50+needlePoint[speed][0], 50-needlePoint[speed][1], 50, 50, c)
		tinydraw.Line(&display, 50+needlePoint[speed][0], 51-needlePoint[speed][1], 50, 51, c)
	} else {
		if speed > 250 {
			speed = 250
		}
		speed -= 180
		tinydraw.Line(&display, 50+needlePoint[speed][0], 50+needlePoint[speed][1], 50, 50, c)
		tinydraw.Line(&display, 50+needlePoint[speed][0], 51+needlePoint[speed][1], 50, 51, c)
	}
}

func stepL(enabled bool) {
	if enabled {
		display.FillRectangle(110, 32, 16, 16, colors[STEPL])
	} else {
		display.FillRectangle(110, 32, 16, 16, colors[BACKGROUND])
	}
}

func stepR(enabled bool) {
	if enabled {
		display.FillRectangle(134, 32, 16, 16, colors[STEPR])
	} else {
		display.FillRectangle(134, 32, 16, 16, colors[BACKGROUND])
	}
}

func getRainbowRGB(i uint8) color.RGBA {
	if i < 85 {
		return color.RGBA{i * 3, 255 - i*3, 0, 255}
	} else if i < 170 {
		i -= 85
		return color.RGBA{255 - i*3, 0, i * 3, 255}
	}
	i -= 170
	return color.RGBA{0, i * 3, 255 - i*3, 255}
}
