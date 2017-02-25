package routelistener

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RouteListener struct {
	HTTPResponse *http.Response
}

type RouteUpdate struct {
	Event string
	Data  *EventData
}

type EventData struct {
	RouterGroupID string `json:"router_group_guid"`
	Address       string `json:"backend_ip"`
	InternalPort  int    `json:"backend_port"`
	ExternalPort  int    `json:"port"`
}

func (ru *RouteUpdate) Parse(s string) {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		vals := strings.Split(line, "|")

		if vals[0] == "event" {
			ru.Event = vals[1]
		} else if vals[0] == "data" {
			dec := json.NewDecoder(strings.NewReader(vals[1]))
			eventData := &EventData{}

			err := dec.Decode(eventData)
			if err != nil {
				fmt.Println(err)
			}
			ru.Data = eventData
		}
	}
}

func (rl RouteListener) Listen(ch chan RouteUpdate) {
	defer rl.HTTPResponse.Body.Close()
	reader := bufio.NewReader(rl.HTTPResponse.Body)

	var err error

	for err == nil {
		var (
			isNewPacket = false
			// isPrefix    = false
			packet []byte
		)

		for !isNewPacket {
			line, err := reader.ReadBytes('\n')

			if err != nil {
				fmt.Println(err)
			}

			if len(line) == 1 {
				isNewPacket = true
			} else {
				line = bytes.Replace(line, []byte(":"), []byte("|"), 1)
				packet = append(packet, line...)
			}
		}

		ru := RouteUpdate{}
		ru.Parse(string(packet))

		ch <- ru
	}
}

func New(resp *http.Response) *RouteListener {

	rl := RouteListener{}
	rl.HTTPResponse = resp

	return &rl
}
