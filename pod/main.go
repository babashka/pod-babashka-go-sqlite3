package pod

import (
	"database/sql"
	_ "encoding/json"
	"github.com/russolsen/transit"
	"fmt"
	"github.com/babashka/pod-babashka-sqlite3/babashka"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
	"strings"
	"os"
	"container/list"
)

type ExecResult struct {
	RowsAffected   int64 `json:"rows-affected"`
	LastInsertedId int64 `json:"last-inserted-id"`
}

func JsonifyRows(rows *sql.Rows) ([]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
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

		if err = rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		for i, value := range values {
			results[columns[i]] = value
		}

		data = append(data, results)
		c++
	}

	return data, nil
}

func JsonifyResult(result sql.Result) (*ExecResult, error) {
	rowsAffected, err := result.RowsAffected()
	lastInsertedId, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &ExecResult{
		RowsAffected:   rowsAffected,
		LastInsertedId: lastInsertedId,
	}, nil
}

func parseQuery(args string) (string, string, []interface{}, error) {
	reader := strings.NewReader(args)
	decoder := transit.NewDecoder(reader)
	value, err := decoder.Decode()

	var theList *list.List
	theList = value.(*list.List)
	front := theList.Front()
	fmt.Fprintf(os.Stderr, "The value read is: %+v\n", front.Value)
	if err != nil {
		return "", "", nil, err
	}
	return "", "", nil, nil
	// var db string

	// var queryArgs []interface{}
	// if err := json.Unmarshal(podArgs[1], &queryArgs); err != nil {
	// 	return "", "", nil, err
	// }

	// var query string
	// query = queryArgs[0].(string)

	// return db, query, queryArgs[1:], nil
}

func makeArgs(query []string) []interface{} {
	args := make([]interface{}, len(query)-1)

	for i := range query[1:] {
		args[i] = query[i+1]
	}

	return args
}

func ProcessMessage(message *babashka.Message) (interface{}, error) {
	switch message.Op {
	case "describe":
		return &babashka.DescribeResponse{
			Format: "transit+json",
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
		db, query, args, err := parseQuery(message.Args)
		if err != nil {
			return nil, err
		}

		conn, err := sql.Open("sqlite3", db)
		if err != nil {
			return nil, err
		}

		defer conn.Close()

		//args := makeArgs(query)

		switch message.Var {
		case "pod.babashka.sqlite3/execute!":
			res, err := conn.Exec(query, args...)
			if err != nil {
				return nil, err
			}

			if json, err := JsonifyResult(res); err != nil {
				return nil, err
			} else {
				return json, nil
			}
		case "pod.babashka.sqlite3/query!":
			res, err := conn.Query(query, args...)
			if err != nil {
				return nil, err
			}

			if json, err := JsonifyRows(res); err != nil {
				return nil, err
			} else {
				return json, nil
			}
		default:
			return nil, fmt.Errorf("Unknown var %s", message.Var)
		}
	default:
		return nil, fmt.Errorf("Unknown op %s", message.Op)
	}
}
