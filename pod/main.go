package pod

import (
        "database/sql"
        "encoding/json"
        "fmt"
        "github.com/babashka/pod-babashka-sqlite3/babashka"
        _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
        _ "log"
)

type Query struct {
        Database string   `json:"db"`
        Query    []string `json:"query"`
}

type ExecResult struct {
        RowsAffected   int64 `json:"rows-affected"`
        LastInsertedId int64 `json:"last-inserted-id"`
}

// type Response struct {
//      Type  string  `json:"type"`
//      Path  string  `json:"path"`
//      Dest  *string `json:"dest,omitempty"`
//      Error *string `json:"error,omitempty"`
// }

// type WatcherInfo struct {
//      WatcherId int `json:"watcher/id"`
// }

func JsonifyRows(rows *sql.Rows) []interface{} {
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
        var data []interface{}

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

                data = append(data, results)
                c++
        }

        return data
}

func JsonifyResult(result sql.Result) (ExecResult, error) {
        rowsAffected, err := result.RowsAffected()
        lastInsertedId, err := result.LastInsertId()
        return ExecResult{
                RowsAffected:   rowsAffected,
                LastInsertedId: lastInsertedId,
        }, err
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
						{
                                                        Name: "query!",
                                                },
                                        },
                                },
                        },
                }, nil
        case "invoke":
                switch message.Var {
                case "pod.babashka.sqlite3/execute!":

                        podArgs := []json.RawMessage{}
                        if err := json.Unmarshal([]byte(message.Args), &podArgs); err != nil {
                                return nil, err
                        }
                        var db string
                        if err := json.Unmarshal(podArgs[0],&db); err != nil {
                                return nil, err
                        }

                        var query []string
                        if err := json.Unmarshal(podArgs[1],&query); err != nil {
                                return nil, err
                        }

                        q := Query{
                                Database: db,
                                Query: query,
                        }

                        // log.Printf("json args %q", q)
                        conn, _ := sql.Open("sqlite3", q.Database)
                        defer conn.Close()
                        args := make([]interface{}, len(q.Query)-1)
                        for i := range q.Query[1:] {
                                args[i] = q.Query[i+1]
                        }
                        res, err := conn.Exec(q.Query[0], args...)
                        if err != nil {
                                return nil, err
                        }
                        json, err := JsonifyResult(res)
                        // log.Println("json", json)
                        return json, nil
                case "pod.babashka.sqlite3/query!":
			podArgs := []json.RawMessage{}
                        if err := json.Unmarshal([]byte(message.Args), &podArgs); err != nil {
                                return nil, err
                        }
                        var db string
                        if err := json.Unmarshal(podArgs[0],&db); err != nil {
                                return nil, err
                        }

                        var query []string
                        if err := json.Unmarshal(podArgs[1],&query); err != nil {
                                return nil, err
                        }

                        q := Query{
                                Database: db,
                                Query: query,
                        }

                        // log.Printf("json args %q", q)
                        conn, _ := sql.Open("sqlite3", q.Database)
                        defer conn.Close()
                        args := make([]interface{}, len(q.Query)-1)
                        for i := range q.Query[1:] {
                                args[i] = q.Query[i+1]
                        }
                        res, err := conn.Query(q.Query[0], args...)
                        if err != nil {
                                return nil, err
                        }
                        json := JsonifyRows(res)
                        // log.Println("json", json)
                        return json, nil
                default:
                        return nil, fmt.Errorf("Unknown var %s", message.Var)
                }
        default:
                return nil, fmt.Errorf("Unknown op %s", message.Op)
        }

}
