package main

import (
	"encoding/json"
	"fmt"
	"github.com/babashka/pod-babashka-sqlite3/babashka"
	"github.com/babashka/pod-babashka-sqlite3/pod"
	"log"
)

func execQuery(query string, exec bool, args ...string) {
	elts := len(args) + 1
	argsArray := make([]string, elts)
	argsArray[0] = query
	for i := range args {
		argsArray[i+1] = args[i]
	}
	log.Printf("args: %q", argsArray)
	q, err := json.Marshal(argsArray)
	var theVar string
	if exec {
		theVar = "pod.babashka.sqlite3/execute!"
	} else {
		theVar = "pod.babashka.sqlite3/query!"
	}
	message := &babashka.Message{
		Op:  "invoke",
		Var: theVar,
		Id:  "123",
		Args: fmt.Sprintf(`{"db":"/tmp/pod.db",
                        "query": %s}`, q)}
	res, err := pod.ProcessMessage(message)
	if err != nil {
		log.Println("err", err)
	}
	log.Println("res", res)
}

func main() {
	execQuery("select sqlite_version()", false)
	execQuery("select sqlite_source_id()", false)
	execQuery("create table if not exists foo (col1 TEXT, col2 TEXT)", true)
	execQuery("delete from foo", true)
	execQuery("insert into foo values (?,?)", true, "1", "2")
	execQuery("select * from foo", false)
}
