package main

import (
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
		babashka.WriteInvokeResponse(message, res)
	}
}
