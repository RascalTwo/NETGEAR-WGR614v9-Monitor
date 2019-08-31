package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
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

func findAllSubmatchGroups(str string, target string, groupIndex int) []string {
	results := make([]string, 0)
	for _, v := range regexp.MustCompile(str).FindAllStringSubmatch(target, -1) {
		results = append(results, v[groupIndex])
	}
	return results
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

	rows := findAllSubmatchGroups(`(?s)<tr>(.*?)</tr>`, html, 1)
	data.Uptime = regexp.MustCompile(`(?s)<!>\s*(.*?)\s*?<!>`).FindStringSubmatch(rows[1])[1]

	lastIO := struct {
		Transfered IOInfo
		Received   IOInfo
		Collisions int16
	}{}
	for _, row := range rows[4:] {
		values := findAllSubmatchGroups(`(?s)<td.*?>.*?<span.*?>(.*?)</span>`, row, 1)
		if len(values) == 8 {
			value, _ := strconv.ParseInt(values[2], 10, 32)
			transInt := int32(value)
			value, _ = strconv.ParseInt(values[5], 10, 32)
			lastIO.Transfered = IOInfo{transInt, int32(value)}

			value, _ = strconv.ParseInt(values[3], 10, 32)
			recvInt := int32(value)
			value, _ = strconv.ParseInt(values[6], 10, 32)
			lastIO.Received = IOInfo{recvInt, int32(value)}

			value, _ = strconv.ParseInt(values[4], 10, 16)
			lastIO.Collisions = int16(value)

			// TODO - Don't ignore errors

			if strings.Contains(values[0], "LAN") {
				lastIO.Transfered.Count /= 4
				lastIO.Transfered.Speed /= 4
				lastIO.Received.Count /= 4
				lastIO.Received.Speed /= 4
				lastIO.Collisions /= 4
			}
		}

		uptime := values[len(values)-1]
		if uptime == "--" {
			uptime = ""
		}
		data.Interfaces[values[0]] = Row{values[1], lastIO.Transfered, lastIO.Received, lastIO.Collisions, uptime}
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

func collectData(rate <-chan time.Time) {
	for range rate {
		html, err := cacheOrFetch("tableout.html", func() ([]byte, error) {
			return fetchHTML("http://10.0.0.1/", "admin", "drowssap")
		})
		if err != nil {
			log.Fatal(err)
		}

		data, _ := parseHTML(html)
		allData[data.When] = data
		fmt.Println(data.When)
		fmt.Println(data)
	}
}

var allData = make(map[time.Time]Data)

func main() {
	rate := time.NewTicker(1 * time.Second)
	go collectData(rate.C)

	fmt.Scanln()
	rate.Stop()
	fmt.Scanln()
}
