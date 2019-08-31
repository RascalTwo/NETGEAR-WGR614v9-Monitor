package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// IOInfo - Packet count and reported speed of IO
type IOInfo struct {
	Count int32
	Speed int32
}

// Row - Single row of interface data
type Row struct {
	Status     string
	Transfered IOInfo
	Received   IOInfo
	Collisions int16
	Uptime     string
}

// Data - Snapshot of all data
type Data struct {
	When       time.Time
	Uptime     string
	Interfaces map[string]Row
}

func fetchHTML(hostpath string, username string, password string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", hostpath+"RST_stattbl.htm", nil)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func parseHTML(html string) (Data, error) {
	data := Data{When: time.Now(), Interfaces: make(map[string]Row)}

	rows := strings.Split(html, "<tr>")
	data.Uptime = strings.Trim(strings.Split(rows[2], "<!>")[1], " ")

	lastBand := struct {
		Transfered IOInfo
		Received   IOInfo
		Collisions int16
	}{}
	for _, row := range rows[5:] {
		cells := strings.Split(row, "<td")
		values := make([]string, 0)
		for _, cell := range cells[1:] {
			rawCell := strings.SplitN(strings.Split(cell, "</td")[0], ">", 2)[1]
			cell := strings.Trim(strings.Split(strings.Split(rawCell, ">")[1], "</span")[0], " ")
			values = append(values, cell)
		}
		if len(values) == 8 {
			value, _ := strconv.ParseInt(values[2], 10, 32)
			transInt := int32(value)
			value, _ = strconv.ParseInt(values[5], 10, 32)
			trans := IOInfo{transInt, int32(value)}

			value, _ = strconv.ParseInt(values[3], 10, 32)
			recvInt := int32(value)
			value, _ = strconv.ParseInt(values[6], 10, 32)
			recv := IOInfo{recvInt, int32(value)}

			value, _ = strconv.ParseInt(values[4], 10, 16)
			collisions := int16(value)

			// TODO - Don't ignore errors

			lastBand.Transfered = trans
			lastBand.Received = recv
			lastBand.Collisions = collisions
			if strings.Contains(values[0], "LAN") {
				lastBand.Transfered.Count /= 4
				lastBand.Transfered.Speed /= 4
				lastBand.Received.Count /= 4
				lastBand.Received.Speed /= 4
				lastBand.Collisions /= 4
			}
		}

		uptime := values[len(values)-1]
		if uptime == "--" {
			uptime = ""
		}
		data.Interfaces[values[0]] = Row{values[1], lastBand.Transfered, lastBand.Received, lastBand.Collisions, uptime}
	}

	return data, nil
}

func cacheOrFetch(filename string, fetch func() ([]byte, error)) (string, error) {
	if _, err := os.Stat(filename); os.IsExist(err) {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(bytes), nil

	}
	bytes, err := fetch()
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func main() {
	fetchAuthedHTML := func() ([]byte, error) {
		return fetchHTML("http://10.0.0.1/", "admin", "drowssap")
	}
	html, err := cacheOrFetch("tableout.html", fetchAuthedHTML)
	if err != nil {
		log.Fatal(err)
	}

	data, _ := parseHTML(html)
	spew.Dump(data)
}
