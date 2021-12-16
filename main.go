package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/huin/goserial"
)

type Measurement struct {
	AirTemperature float64 `json:"air_temperature"`
	Humidity       float64 `json:"humidity"`
	WaterLevel     float64 `json:"water_level"`
}

var (
	m  Measurement
	cl = &http.Client{Timeout: 2 * time.Second}
)

func main() {
	// Find the device that represents the Arduino serial connection.
	c := &goserial.Config{Name: `\\.\COM5`, Baud: 9600}
	s, err := goserial.OpenPort(c)
	if err != nil {
		log.Fatalf("could not open port: %s", err.Error())
	}

	// When connecting to an older revision Arduino, you need to wait
	// a little while it resets.
	time.Sleep(1 * time.Second)

	br := bufio.NewReader(s)
	for {
		b, err := br.ReadBytes('\n')
		if err != nil {
			log.Printf("could not read line: %s\n", err.Error())
			continue
		}

		if len(b) < 2 {
			log.Println("got empty line")
			continue
		}

		err = json.Unmarshal(b[:len(b)-2], &m)
		if err != nil {
			log.Printf("could not decode JSON '%s': %s\n", b, err.Error())
			continue
		}

		log.Printf("Lufttemperatur: %.1f°C, Luftfeuchte: %.1f%%, Wasserstand: %.1f%%\n", m.AirTemperature, m.Humidity, m.WaterLevel)
	}
}
