package pod

import (
        "database/sql"
        "encoding/json"
        "log"
        _ "os"
        "fmt"
        "strings"
        "github.com/babashka/pod-babashka-sqlite3/babashka"
        _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Opts struct {
        DelayMs   uint64 `json:"delay-ms"`
        Recursive bool   `json:"recursive"`
}

type Response struct {
        Type  string  `json:"type"`
        Path  string  `json:"path"`
        Dest  *string `json:"dest,omitempty"`
        Error *string `json:"error,omitempty"`
}

type WatcherInfo struct {
        WatcherId int `json:"watcher/id"`
}

func Jsonify(rows *sql.Rows) ([]string) {
        columns, err := rows.Columns()
        if err != nil {
                panic(err.Error())
        }

        values := make([]interface{}, len(columns))

        scanArgs := make([]interface{}, len(values))
        for i := range values {
                scanArgs[i] = &values[i]
        }

        c := 0
        results := make(map[string]interface{})
        data := []string{}

        for rows.Next() {
                if c > 0 {
                        data = append(data, ",")
                }

                err = rows.Scan(scanArgs...)
                if err != nil {
                        panic(err.Error())
                }

                for i, value := range values {
                        results[columns[i]] = value
                        // switch value.(type) {
                        // case nil:
                        //      results[columns[i]] = nil

                        // case []byte:
                        //      s := string(value.([]byte))
                        //      x, err := strconv.Atoi(s)

                        //      if err != nil {
                        //              results[columns[i]] = s
                        //      } else {
                        //              results[columns[i]] = x
                        //      }


                        // default:
                        //      results[columns[i]] = value
                        // }
                }

                b, _ := json.Marshal(results)
                data = append(data, strings.TrimSpace(string(b)))
                c++
        }

        return data
}


func ProcessMessage(message *babashka.Message) (interface{}, error) {
        switch message.Op {
        case "describe":
                return &babashka.DescribeResponse{
                        Format: "json",
                        Namespaces: []babashka.Namespace{
                                {
                                        Name: "pod.babashka.sqlite3",
                                        Vars: []babashka.Var{
                                                {
                                                        Name: "execute!",
                                                },
                                        },
                                },
                        },
                }, nil
        case "invoke":
                switch message.Var {
		case "pod.babashka.sqlite3/execute!":
			args := []string{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}
			log.Println("args", args)
			return nil, nil
		default:
                        return nil, fmt.Errorf("Unknown var %s", message.Var)
                }
        default:
                return nil, fmt.Errorf("Unknown op %s", message.Op)
        }

}
