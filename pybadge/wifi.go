package main

import (
	"image/color"
	"machine"
	"math/rand"
	"strconv"
	"time"

	"tinygo.org/x/drivers/wifinina"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"

	"tinygo.org/x/drivers/net/mqtt"
)

const ssid = "YOURSSID"
const pass = "YOURPASS"

//const server = "ssl://test.mosquitto.org:8883"
const server = "tcp://test.mosquitto.org:1883"

const TRACKLENGTH = 300
const LAPS = 3

var (
	uart = machine.UART1
	tx   = machine.UART_TX_PIN
	rx   = machine.UART_RX_PIN
	spi  = machine.SPI0

	// this is the ESP chip that has the WIFININA firmware flashed on it
	adaptor = &wifinina.Device{
		SPI:   spi,
		CS:    machine.D13,
		ACK:   machine.D11,
		GPIO0: machine.D10,
		RESET: machine.D12,
	}

	console = machine.UART0

	cl      mqtt.Client
	topicTx = "tinygo/tx"
	topicRx = "tinygo/rx"
	payload []byte
	enabled bool
)

func updateTrackInfo(client mqtt.Client, msg mqtt.Message) {
	// this code causes a hardfault    ¿?
	ba := msg.Payload()
	if len(ba) != 4 {
		return
	}
	var speed int16
	speed |= int16(ba[0])
	speed |= int16(ba[1]) << 8

	var progress int16
	progress |= int16(ba[2])
	progress |= int16(ba[3]) << 8

	speedGaugeNeedle(speed, colors[BLACK])
	speedGaugeNeedle(speed, colors[player])
	oldSpeed = speed

	progressLapBar(float32(progress % TRACKLENGTH))
	progressRaceBar(float32(progress) / (LAPS * TRACKLENGTH))

}

func configureWifi(player int) {
	display.FillScreen(color.RGBA{0, 0, 0, 255})

	topicTx = "player" + strconv.Itoa(player) + "/tx"
	topicRx = "player" + strconv.Itoa(player) + "/rx"

	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})
	rand.Seed(time.Now().UnixNano())

	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		MOSI:      machine.SPI0_MOSI_PIN,
		MISO:      machine.SPI0_MISO_PIN,
		SCK:       machine.SPI0_SCK_PIN,
	})

	// Init esp8266/esp32
	adaptor.Configure()

	connectToAP()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server).SetClientID("tinygo-racer-" + strconv.Itoa(player))

	println("Connecting to MQTT broker at", server)
	cl = mqtt.NewClient(opts)
	if token := cl.Connect(); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
		tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 40, []byte(token.Error().Error()), color.RGBA{255, 0, 0, 255})
	}

	// subscribe
	token := cl.Subscribe(topicRx, 0, updateTrackInfo)
	token.Wait()
	if token.Error() != nil {
		tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 50, []byte(token.Error().Error()), color.RGBA{255, 0, 0, 255})
		failMessage(token.Error().Error())
	}

	enabled = true

	go sendLoop()

	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 70, []byte("Done."), color.RGBA{0, 255, 0, 255})
	println("Done.")
}

// connect to access point
func connectToAP() {
	time.Sleep(2 * time.Second)
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 90, []byte("Connecting to '"+ssid+"'"), color.RGBA{255, 255, 255, 255})
	println("Connecting to " + ssid)
	adaptor.SetPassphrase(ssid, pass)
	for st, _ := adaptor.GetConnectionStatus(); st != wifinina.StatusConnected; {
		tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 100, []byte(st.String()), color.RGBA{255, 0, 0, 255})
		println("Connection status: " + st.String())
		time.Sleep(1000 * time.Millisecond)
		st, _ = adaptor.GetConnectionStatus()
	}
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 110, []byte("Connected :D"), color.RGBA{0, 255, 0, 255})
	println("Connected.")
	time.Sleep(2 * time.Second)
	ip, _, _, err := adaptor.GetIP()
	for ; err != nil; ip, _, _, err = adaptor.GetIP() {
		tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 120, []byte(err.Error()), color.RGBA{255, 0, 0, 255})
		println("IP", err.Error())
		time.Sleep(1 * time.Second)
	}
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 120, []byte("IP: "+ip.String()), color.RGBA{0, 255, 0, 255})
	println(ip.String())
}

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}

func sendLoop() {
	retries := uint8(0)
	var token mqtt.Token

	for {
		if enabled {
			if retries == 0 {
				println("Publishing MQTT message...", string(payload))
				token = cl.Publish(topicTx, 0, false, payload)
				token.Wait()
			}
			if retries > 0 || token.Error() != nil {
				if retries < 10 {
					token = cl.Connect()
					if token.Wait() && token.Error() != nil {
						retries++
						println("NOT CONNECTED TO MQTT (sendLoop)")
					} else {
						retries = 0
					}
				} else {
					enabled = false
				}
			}
			payload = []byte("none")
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func Send(mqttpayload []byte) {
	payload = mqttpayload
}
