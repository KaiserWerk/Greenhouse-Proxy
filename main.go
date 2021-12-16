package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/huin/goserial"
)

type Measurement struct {
	AirTemperature float64 `json:"air_temperature"`
	Humidity       float64 `json:"humidity"`
	WaterLevel     float64 `json:"water_level"`
}

var m Measurement

func main() {
	// Find the device that represents the Arduino serial connection.
	c := &goserial.Config{Name: `\\.\COM4`, Baud: 9600}
	s, err := goserial.OpenPort(c)
	if err != nil {
		log.Fatalf("could not open port: %s", err.Error())
	}

	// When connecting to an older revision Arduino, you need to wait
	// a little while it resets.
	time.Sleep(1 * time.Second)

	br := bufio.NewReader(s)
	for {
		b, _, err := br.ReadLine()
		if err != nil {
			fmt.Println("could not read line:", err.Error())
			break
		}

		//log.Printf("received data: %s\n", b)
		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Printf("could not decode JSON '%s': %s", b, err.Error())
			break
		}

		log.Printf("Lufttemperatur: %.1f°C, Luftfeuchte: %.1f%%, Wasserstand: %.1f%%\n", m.AirTemperature, m.Humidity, m.WaterLevel)
	}
}
