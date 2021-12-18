package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/huin/goserial"
)

type Measurement struct {
	AirTemperature float64 `json:"air_temperature"`
	Humidity       float64 `json:"humidity"`
	WaterLevel     float64 `json:"water_level"`
}

const (
	apiRoute = "/api/v1/receive"
)

var (
	err              error
	baseUrl          = "http://127.0.0.1:47336/api/v1/receive"
	key              = ""
	emptyLineCounter int64
	maxEmptyLines    int64 = 100
)

func main() {
	log.Println("Starting up...")
	var (
		m    Measurement
		s    io.ReadWriteCloser
		port string
		cl   = &http.Client{Timeout: 2 * time.Second}
	)

start:

	if u := os.Getenv("GREENHOUSE_PROXY_BASEURL"); u != "" {
		log.Printf("found baseUrl env var, setting to '%s'...\n", u)
		baseUrl = u
	}

	if k := os.Getenv("GREENHOUSE_PROXY_KEY"); k != "" {
		log.Printf("found key env var, setting to '%s'...\n", k)
		key = k
	}

	// NOTE: find the device that represents the Arduino serial connection
	log.Println("Now trying to find open Serial Port...")
	for i := 0; i < 10; i++ {
		if runtime.GOOS == "windows" {
			port = fmt.Sprintf(`\\.\COM%d`, i)
		} else {
			port = fmt.Sprintf(`/dev/ttyUSB%d`, i)
		}
		log.Printf("attempting to open port %s...\n", port)
		c := &goserial.Config{Name: port, Baud: 9600}
		s, err = goserial.OpenPort(c)
		if err == nil {
			log.Printf("now using port %s\n", port)
			break
		}
		log.Printf("could not open port %s: %s\n", port, err.Error())
	}

	// When connecting to an older revision Arduino, you need to wait
	// a little while it resets.
	log.Println("waiting for arduino reset")
	time.Sleep(1 * time.Second)

	log.Println("starting to read from Serial Port...")
	br := bufio.NewReader(s)
	for {
		if emptyLineCounter > maxEmptyLines {
			emptyLineCounter = 0
			break
		}
		b, err := br.ReadBytes('\n')
		if err != nil {
			log.Printf("could not read line: %s\n", err.Error())
			continue
		}

		if len(b) < 2 {
			log.Println("got empty line")
			emptyLineCounter++
			continue
		}

		err = json.Unmarshal(b[:len(b)-2], &m) // strip the '\n'
		if err != nil {
			log.Printf("could not decode JSON '%s': %s\n", b, err.Error())
			continue
		}

		log.Printf("Lufttemperatur: %.1f°C, Luftfeuchte: %.1f%%, Wasserstand: %.1f%%\n", m.AirTemperature, m.Humidity, m.WaterLevel)
		err = sendMeasurement(cl, &m)
		if err != nil {
			log.Printf("could not set POST request with measurements: %s", err.Error())
		}
	}

	goto start
}

func sendMeasurement(cl *http.Client, m *Measurement) error {
	b, err := json.Marshal(*m)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, baseUrl+apiRoute, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("X-Greenhouse-Key", key)

	resp, err := cl.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("expected success status code, got %d", resp.StatusCode)
	}

	return nil
}
