package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"time"

	"github.com/lucperkins/rek"
	"github.com/tadvi/log"
)

const timeout = 15

var (
	esp      string
	lowTemp  int
	heatTime int
)

type ESPMessage struct {
	Message     string `json:"message"`
	Temperature int    `json:"temperature"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Hardware    string `json:"hardware"`
	Connected   bool   `json:"connected"`
}

func init() {
	flag.StringVar(&esp, "esp", "ESP_4E2ABA", "ESP8266 device address")

	flag.IntVar(&lowTemp, "low-temp", 21, "low temprature to turn heater on")
	flag.IntVar(&heatTime, "heat-time", 12, "heat time")
}

func main() {
	flag.Parse()

	esp = "http://" + esp

	for {
		tm := time.Now().Local()

		if tm.Minute() == 0 {

			res, err := rek.Get(esp+"/temperature", rek.Timeout(timeout*time.Second))
			if err != nil {
				log.Errorf("Error getting temperature: %v", err)
				continue
			}
			if res.StatusCode() >= 300 {
				log.Errorf("Error getting temperature: %v", err, res.StatusCode())
				continue
			}
			defer res.Body().Close()

			var msg ESPMessage
			bs, err := ioutil.ReadAll(res.Body())
			if err != nil {
				log.Errorf("Error reading request: %v", err)
				continue
			}

			if err := json.Unmarshal(bs, &msg); err != nil {
				log.Errorf("Error parsing request: %v", err)
				log.Errorf("Dump: %+v", msg)
				continue
			}

			log.Debugf("Dump: %+v", msg)

			if msg.Temperature <= lowTemp {
				log.Infof("Heater ON")

				res, err := rek.Get(esp+"/digital/8/1", rek.Timeout(timeout*time.Second))
				if err != nil {
					log.Errorf("Error turning on the heater: %v", err)
					continue
				}
				res.Body().Close()

				time.Sleep(time.Minute * time.Duration(heatTime))

				res, err = rek.Get(esp+"/digital/8/0", rek.Timeout(timeout*time.Second))
				if err != nil {
					log.Errorf("Error turning off the heater: %v", err)
					continue
				}
				res.Body().Close()
				log.Infof("Heater OFF")
			}
		}

		time.Sleep(time.Minute)
	}
}
