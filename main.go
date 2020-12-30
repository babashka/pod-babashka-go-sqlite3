package main

import (
	"encoding/json"

	"github.com/babashka/pod-babashka-sqlite3/babashka"
	"github.com/babashka/pod-babashka-sqlite3/pod"
)

func main() {
	for {
		message, err := babashka.ReadMessage()
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			continue
		}

		res, err := pod.ProcessMessage(message)
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			continue
		}

		describeRes, ok := res.(*babashka.DescribeResponse)
		if ok {
			babashka.WriteDescribeResponse(describeRes)
			continue
		}

		if json, err := json.Marshal(res); err != nil {
			babashka.WriteErrorResponse(message, err)
		} else {
			babashka.WriteInvokeResponse(message, string(json))
		}
	}
}
