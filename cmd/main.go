package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

func main() {
	// import todays Readings
	// written, err := io.Copy(os.Stdout, port)
	// log.Printf("Copied %v bytes", written)
	err := scanSensor()
	if err != nil {
		log.Fatalf("scanning problem: %v", err)
	}

}

type SensorReading struct {
	When        time.Time
	Co2         int     `json:"CO2"`
	Pressure    float64 `json:"pressure"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

func scanSensor() (err error) {
	port, err := serial.Open(serial.OpenOptions{
		PortName:                "/dev/tty.usbmodem142101",
		BaudRate:                9600,
		DataBits:                8,
		StopBits:                1,
		ParityMode:              serial.PARITY_NONE,
		RTSCTSFlowControl:       false,
		InterCharacterTimeout:   1,
		MinimumReadSize:         1,
		Rs485Enable:             false,
		Rs485RtsHighDuringSend:  false,
		Rs485RtsHighAfterSend:   false,
		Rs485RxDuringTx:         false,
		Rs485DelayRtsBeforeSend: 0,
		Rs485DelayRtsAfterSend:  0,
	})
	if err != nil {
		return fmt.Errorf("Couldn't open serial: %w", err)
	}
	defer port.Close()

	scan := bufio.NewScanner(port)
	var readings SensorReadings

	err = readings.Import()
	if err != nil {
		log.Printf(`¯\_(ツ)_/¯ Didn't find anything to import ¯\_(ツ)_/¯  : %v`+"\n", err)
	}
	lastDay := time.Now().YearDay()
	for scan.Scan() {
		sr := SensorReading{When: time.Now().Local()}
		err := json.Unmarshal(scan.Bytes(), &sr)

		if err != nil {
			log.Printf("Marshall Error: %v", err)
			continue
		}
		if lastDay != sr.When.YearDay() {
			lastDay = sr.When.YearDay()
			readings = nil
		}
		readings = append(readings, sr)
		readings.Dump()
		if err != nil {
			log.Printf("Dump Error: %v", err)
		}
		fmt.Printf(".")

	}
	return nil
}

type SensorReadings []SensorReading

const fnameFormat = "readings-2006-01-02.json"

func (srs SensorReadings) Dump() (err error) {
	bits, err := json.Marshal(srs)
	if err != nil {
		return
	}
	return os.WriteFile(time.Now().Local().Format(fnameFormat), bits, 0700)
}
func (srs *SensorReadings) Import() (err error) {
	bits, err := os.ReadFile(time.Now().Local().Format(fnameFormat))
	if err != nil {
		return err
	}
	return json.Unmarshal(bits, srs)
}

func (sr SensorReading) ToRecord() (rec []string) {
	return []string{
		sr.When.Format(time.RFC3339),
		strconv.Itoa(sr.Co2),
		strconv.FormatFloat(sr.Pressure, 'f', 2, 64),
		strconv.FormatFloat(sr.Temperature, 'f', 2, 64),
		strconv.FormatFloat(celsToFahr(sr.Temperature), 'f', 2, 64),
		strconv.FormatFloat(sr.Humidity, 'f', 2, 64)}
}
func (srs SensorReadings) ToRecords() (recs [][]string) {
	recs = append(recs, []string{"When", "CO2", "Pressure", "Temperature Celsius", "Temperature Fahrenheit", "Humidity"})
	for _, r := range srs {
		recs = append(recs, r.ToRecord())
	}
	return
}

func celsToFahr(cels float64) float64 {
	return cels*1.8 + 32
}
