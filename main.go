package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func fetch_html() string {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://10.0.0.1/RST_stattbl.htm", nil)
	req.Header.Add("Authorization", "Basic "+basicAuth("admin", "drowssap"))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	text, err := ioutil.ReadAll(resp.Body)
	return string(text)
}

type IOInfo struct {
	Count int32
	Speed int32
}

type Row struct {
	Status     string
	Transfered IOInfo
	Received   IOInfo
	Collisions int16
	Uptime     string
}

type Data struct {
	Uptime     string
	Interfaces map[string]Row
}

type Cell struct {
	isHead bool
	value  string
}

func main() {
	//html := fetch_html()
	//fmt.Println(html)
	//err := ioutil.WriteFile("tableout.html", []byte(html), 0644)
	bytes, _ := ioutil.ReadFile("tableout.html")

	data := Data{Interfaces: make(map[string]Row)}

	rows := strings.Split(string(bytes), "<tr>")
	data.Uptime = strings.Trim(strings.Split(rows[2], "<!>")[1], " ")

	lastBand := struct {
		Transfered IOInfo
		Received   IOInfo
		Collisions int16
	}{}
	for i := 5; i < len(rows); i++ {
		cells := strings.Split(rows[i], "<td")
		values := make([]string, 0)
		for j := 1; j < len(cells); j++ {
			rawCell := strings.SplitN(strings.Split(cells[j], "</td")[0], ">", 2)[1]
			cell := strings.Split(strings.Split(rawCell, ">")[1], "</span")[0]
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

		data.Interfaces[values[0]] = Row{values[1], lastBand.Transfered, lastBand.Received, lastBand.Collisions, values[len(values)-1]}
	}

	//fmt.Printf("%+v\n", data)
	spew.Dump(data)
}
