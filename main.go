package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"container/list"
	"database/sql"

	"github.com/babashka/pod-babashka-go-sqlite3/babashka"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
	"github.com/russolsen/transit"
)

func debug(v interface{}) {
	fmt.Fprintf(os.Stderr, "debug: %+q\n", v)
}

func encodeRows(rows *sql.Rows) ([]interface{}, error) {
	cols, err := rows.Columns()
	columns := make([]transit.Keyword, len(cols))
	for i, col := range cols {
		columns[i] = transit.Keyword(col)
	}
	if err != nil {
		return nil, err
	}

	var data []interface{}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		results := make(map[transit.Keyword]interface{})

		if err = rows.Scan(scanArgs...); err != nil {
			debug(err)
			return nil, err
		}

		for i, val := range values {
			col := columns[i]
			results[col] = val
		}

		// debug(results)
		// debug(fmt.Sprintf("%T", results))

		data = append(data, results)
	}

	return data, nil
}

type ExecResult = map[transit.Keyword]int64

func encodeResult(result sql.Result) (ExecResult, error) {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	res := ExecResult{
		transit.Keyword("rows-affected"):    rowsAffected,
		transit.Keyword("last-inserted-id"): lastInsertedId,
	}

	return res, nil
}

func listToSlice(l *list.List) []interface{} {
	slice := make([]interface{}, l.Len())
	cnt := 0
	for e := l.Front(); e != nil; e = e.Next() {
		slice[cnt] = e.Value
		cnt++
	}

	return slice
}

func parseQuery(args string) (string, string, []interface{}, error) {
	reader := strings.NewReader(args)
	decoder := transit.NewDecoder(reader)
	value, err := decoder.Decode()
	if err != nil {
		return "", "", nil, err
	}

	argSlice := listToSlice(value.(*list.List))
	db := argSlice[0].(string)

	switch queryArgs := argSlice[1].(type) {
	case string:
		return db, queryArgs, make([]interface{}, 0), nil
	case []interface{}:
		return db, queryArgs[0].(string), queryArgs[1:], nil
	default:
		return "", "", nil, errors.New("unexpected query type, expected a string or a vector")
	}
}

func makeArgs(query []string) []interface{} {
	args := make([]interface{}, len(query)-1)

	for i := range query[1:] {
		args[i] = query[i+1]
	}

	return args
}

func respond(message *babashka.Message, response interface{}) {
	buf := bytes.NewBufferString("")
	encoder := transit.NewEncoder(buf, false)

	if err := encoder.Encode(response); err != nil {
		babashka.WriteErrorResponse(message, err)
	} else {
		babashka.WriteInvokeResponse(message, string(buf.String()))
	}
}

func processMessage(message *babashka.Message) {
	switch message.Op {
	case "describe":
		babashka.WriteDescribeResponse(
			&babashka.DescribeResponse{
				Format: "transit+json",
				Namespaces: []babashka.Namespace{
					{
						Name: "pod.babashka.go-sqlite3",
						Vars: []babashka.Var{
							{
								Name: "execute!",
							},
							{
								Name: "query",
							},
						},
					},
				},
			})
	case "invoke":
		db, query, args, err := parseQuery(message.Args)
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			return
		}

		conn, err := sql.Open("sqlite3", db)
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			return
		}

		defer conn.Close()

		switch message.Var {
		case "pod.babashka.go-sqlite3/execute!":
			res, err := conn.Exec(query, args...)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}

			if json, err := encodeResult(res); err != nil {
				babashka.WriteErrorResponse(message, err)
			} else {
				respond(message, json)
			}
		case "pod.babashka.go-sqlite3/query":
			res, err := conn.Query(query, args...)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}

			if json, err := encodeRows(res); err != nil {
				babashka.WriteErrorResponse(message, err)
			} else {
				respond(message, json)
			}
		default:
			babashka.WriteErrorResponse(message, fmt.Errorf("Unknown var %s", message.Var))
		}
	default:
		babashka.WriteErrorResponse(message, fmt.Errorf("Unknown op %s", message.Op))
	}
}

func main() {
	for {
		message, err := babashka.ReadMessage()
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			continue
		}

		processMessage(message)
	}
}
