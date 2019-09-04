package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APIMessage are messages the server responds to clients with
type APIMessage struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	FullFrame interface{} `json:"data,omitempty"`
}

var username = ""
var password = ""
var ip = ""
var rate = int32(1000)
var interval = *time.NewTicker(time.Duration(rate) * time.Millisecond)

var latest = FullFrame{}

func sendJSON(w http.ResponseWriter, success bool, message string, data interface{}) {
	bytes, _ := json.Marshal(APIMessage{success, message, data})

	fmt.Fprintf(w, string(bytes))
}

func api(w http.ResponseWriter, req *http.Request) {
	sendJSON(w, true, "", latest)
}

func update(w http.ResponseWriter, req *http.Request) {
	type UpdateRequest struct {
		IP       string `json:"ip"`
		Username string `json:"username"`
		Password string `json:"password"`
		Rate     int32  `json:"rate"`
	}

	data := UpdateRequest{Rate: -1}
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		sendJSON(w, false, fmt.Sprintf("Error parsing JSON: %v", err), nil)
		return
	}

	if data.Username != "" {
		username = data.Username
	}
	if data.Password != "" {
		password = data.Password
	}
	if data.IP != "" {
		ip = data.IP
	}
	if data.Rate != -1 {
		rate = data.Rate
		interval.Stop()
		interval = *time.NewTicker(time.Duration(rate) * time.Millisecond)
	}

	sendJSON(w, true, "Things changed", UpdateRequest{ip, username, password, rate})
}

func main() {
	go collectData(&interval, func(data FullFrame) {
		fmt.Println(data.When)
		latest = data
	}, &ip, &username, &password)

	http.Handle("/", http.FileServer(http.Dir("./dist")))
	http.HandleFunc("/api", api)
	http.HandleFunc("/update", update)
	go http.ListenAndServe(":5959", nil)

	fmt.Println("Enter any key to stop data collection...")
	fmt.Scanln()
	interval.Stop()

	fmt.Printf("FullFrame collection halted.\nEnter any key to shutdown server and exit...")
	fmt.Scanln()
}
