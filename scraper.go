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
	Count int32 `json:"count"`
	Speed int32 `json:"speed"`
}

// Row - Single row of interface data
type Row struct {
	Status      string `json:"status"`
	Transmitted IOInfo `json:"transmitted"`
	Received    IOInfo `json:"received"`
	Collisions  int16  `json:"collisions"`
	Uptime      string `json:"uptime"`
}

// FullFrame - Snapshot of all data
type FullFrame struct {
	When       time.Time      `json:"when"`
	Uptime     string         `json:"uptime"`
	Interfaces map[string]Row `json:"interfaces"`
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

func isLAN(values []string) bool {
	return strings.HasPrefix(values[0], "LAN")
}

func isActive(values []string) bool {
	return !strings.Contains(strings.ToLower(values[1]), "link down")
}

func parseHTML(html string) (FullFrame, error) {
	data := FullFrame{When: time.Now(), Interfaces: make(map[string]Row)}

	// TODO - Look into combining rows regex with rowvalues regex
	rows := findAllSubmatchGroups(`(?s)<tr>(.*?)</tr>`, html, 1)
	data.Uptime = regexp.MustCompile(`(?s)<!>\s*(.*?)\s*?<!>`).FindStringSubmatch(rows[1])[1]

	rowValues := make([][]string, 0)
	for _, rawRow := range rows[3:] {
		// TODO - Strip uptime strings
		rowValues = append(rowValues, findAllSubmatchGroups(`(?s)<td.*?>.*?<span.*?>(.*?)</span>`, rawRow, 1))
	}

	lastIO := struct {
		Transmitted IOInfo
		Received    IOInfo
		Collisions  int16
	}{}
	activeLANs := make(map[string]bool, 0)
	for _, values := range rowValues {
		if !isLAN(values) || !isActive(values) {
			continue
		}
		activeLANs[values[0]] = true
	}
	for _, values := range rowValues {
		if len(values) == 8 {
			value, _ := strconv.ParseInt(values[2], 10, 32)
			transInt := int32(value)
			value, _ = strconv.ParseInt(values[5], 10, 32)
			lastIO.Transmitted = IOInfo{transInt, int32(value)}

			value, _ = strconv.ParseInt(values[3], 10, 32)
			recvInt := int32(value)
			value, _ = strconv.ParseInt(values[6], 10, 32)
			lastIO.Received = IOInfo{recvInt, int32(value)}

			value, _ = strconv.ParseInt(values[4], 10, 16)
			lastIO.Collisions = int16(value)

			// TODO - Don't ignore errors

			if activeLANs[values[0]] {
				lastIO.Transmitted.Count /= int32(len(activeLANs))
				lastIO.Transmitted.Speed /= int32(len(activeLANs))
				lastIO.Received.Count /= int32(len(activeLANs))
				lastIO.Received.Speed /= int32(len(activeLANs))
				lastIO.Collisions /= int16(len(activeLANs))
			}
		}

		// TODO - Ensure uptime and other fields are stripped
		uptime := values[len(values)-1]
		if uptime == "--" {
			uptime = ""
		}

		if !isActive(values) {
			data.Interfaces[values[0]] = Row{values[1], IOInfo{}, IOInfo{}, 0, uptime}
		} else {
			data.Interfaces[values[0]] = Row{values[1], lastIO.Transmitted, lastIO.Received, lastIO.Collisions, uptime}
		}
	}

	return data, nil
}

func cacheOrFetch(filename string, fetch func() ([]byte, error)) ([]byte, error) {
	if _, err := os.Stat(filename); os.IsExist(err) {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return []byte{}, err
		}
		return bytes, nil

	}
	bytes, err := fetch()
	if err != nil {
		return []byte{}, err
	}

	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func collectData(ticker *time.Ticker, provideData func(data FullFrame), ip *string, username *string, password *string) {
	for {
		select {
		case <-ticker.C:
			if *ip == "" || *username == "" || *password == "" {
				continue
			}
			bytes, err := fetchHTML(fmt.Sprintf("http://%s/", *ip), *username, *password)
			if err != nil {
				log.Fatal(err)
			}

			data, _ := parseHTML(string(bytes))
			provideData(data)
		}
	}
}
