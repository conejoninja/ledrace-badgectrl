package main

import (
	"fmt"
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
const server = "ssl://test.mosquitto.org:8883"

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
)

func subHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("[%s]  ", msg.Topic())
	fmt.Printf("%s\r\n", msg.Payload())
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
	token := cl.Subscribe(topicRx, 0, subHandler)
	token.Wait()
	if token.Error() != nil {
		tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 50, []byte(token.Error().Error()), color.RGBA{255, 0, 0, 255})
		failMessage(token.Error().Error())
	}

	go publishing()

	/*// Right now this code is never reached. Need a way to trigger it...
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 60, []byte("Disconnecting MQTT..."), color.RGBA{255, 0, 0, 255})
	println("Disconnecting MQTT...")
	cl.Disconnect(100)    */

	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 70, []byte("Done."), color.RGBA{0, 255, 0, 255})
	println("Done.")
}

func publishing() {
	for {
		println("Publishing MQTT message...")
		data := []byte("{\"e\":[{ \"n\":\"hello\", \"v\":101 }]}")
		token := cl.Publish(topicTx, 0, false, data)
		token.Wait()
		if token.Error() != nil {
			println(token.Error().Error())
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

// connect to access point
func connectToAP() {
	time.Sleep(2 * time.Second)
	tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 90, []byte("Connecting to '"+ssid+"'"), color.RGBA{255, 255, 255, 255})
	println("Connecting to " + ssid)
	adaptor.SetPassphrase(ssid, pass)
	k := 0
	for st, _ := adaptor.GetConnectionStatus(); st != wifinina.StatusConnected; {
		tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 0, 100, []byte(st.String()), color.RGBA{255, 0,0, 255})
		println("Connection status: " + st.String())
		time.Sleep(1000 * time.Millisecond)
		st, _ = adaptor.GetConnectionStatus()
		k++
		if k > 4 {
			break
		}
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
