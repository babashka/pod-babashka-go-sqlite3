package main

import (
	"bytes"
	"github.com/babashka/pod-babashka-sqlite3/babashka"
	"github.com/babashka/pod-babashka-sqlite3/pod"
	"github.com/russolsen/transit"
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

		buf := bytes.NewBufferString("")
		encoder := transit.NewEncoder(buf, false)
		if err := encoder.Encode(res); err != nil {
			babashka.WriteErrorResponse(message, err)
		} else {
			//println("buf", buf.String())
			babashka.WriteInvokeResponse(message, string(buf.String()))
		}
	}
}
