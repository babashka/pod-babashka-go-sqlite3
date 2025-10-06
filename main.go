package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	// "reflect"

	"container/list"
	"database/sql"

	"github.com/babashka/pod-babashka-go-sqlite3/babashka"
	"github.com/babashka/transit-go"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library

	"github.com/google/uuid"
)

var syncMap sync.Map

func debug(v interface{}) {
	fmt.Fprintf(os.Stderr, "debug: %+q\n", v)
}

func encodeRows(rows *sql.Rows) ([]interface{}, error) {
	defer rows.Close()

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

func parseQuery(args string) (string, string, string, []interface{}, error) {
	reader := strings.NewReader(args)
	decoder := transit.NewDecoder(reader)
	value, err := decoder.Decode()
	if err != nil {
		return "", "", "", nil, err
	}

	argSlice := listToSlice(value.(*list.List))
	var id string
	var db string

	switch first := argSlice[0].(type) {
	case string:
		db = first
	case map[interface{}]interface{}:
		connId, ok := first["connection"].(string)
		if !ok {
			return "", "", "", nil, errors.New(`the "connection" key in the map must be a string`)
		}
		id = connId

	default:
		return "", "", "", nil, errors.New("the sqlite connection must be a string or a map with a \"connection\" key")
	}

	switch queryArgs := argSlice[1].(type) {
	case string:
		return db, id, queryArgs, make([]interface{}, 0), nil
	case []interface{}:
		return db, id, queryArgs[0].(string), queryArgs[1:], nil
	default:
		return "", "", "", nil, errors.New("unexpected query type, expected a string or a vector")
	}
}

func parseGetConnectionArgs(args string) (string, error) {
	reader := strings.NewReader(args)
	decoder := transit.NewDecoder(reader)
	value, err := decoder.Decode()
	if err != nil {
		return "", err
	}

	argSlice := listToSlice(value.(*list.List))
	db, ok := argSlice[0].(string)
	if !ok {
		return "", errors.New("the sqlite connection must be a string")
	}
	return db, nil
}

func parseCloseConnectionArgs(args string) (string, error) {
	reader := strings.NewReader(args)
	decoder := transit.NewDecoder(reader)
	value, err := decoder.Decode()
	if err != nil {
		return "", err
	}
	argSlice := listToSlice(value.(*list.List))
	var id string
	switch first := argSlice[0].(type) {
	case map[interface{}]interface{}:
		connId, ok := first["connection"].(string)
		if !ok {
			return "", errors.New(`the "connection" key in the map must be a string`)
		}
		id = connId

	default:
		return "", errors.New("the sqlite connection must be a map with a \"connection\" key")
	}

	return id, nil
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

func getConn(db string, connId string) (*sql.DB, bool, error) {
	var conn *sql.DB

	if connId == "" {
		newConn, err := sql.Open("sqlite3", db)
		if err != nil {
			return nil, false, err
		}
		conn = newConn
		return conn, true, nil
	} else {
		cached, ok := syncMap.Load(connId)
		if !ok {
			return nil, false, fmt.Errorf("Invalid connection id: %s", connId)
		}
		conn = cached.(*sql.DB)
		return conn, false, nil
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
							{
								Name: "get-connection",
							},
							{
								Name: "close-connection",
							},
						},
					},
				},
			})
	case "invoke":

		switch message.Var {
		case "pod.babashka.go-sqlite3/execute!":
			db, connId, query, args, err := parseQuery(message.Args)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			conn, shouldDefer, err := getConn(db, connId)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			if shouldDefer {
				defer conn.Close()
			}

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
			db, connId, query, args, err := parseQuery(message.Args)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			conn, shouldDefer, err := getConn(db, connId)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			if shouldDefer {
				defer conn.Close()
			}

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
		case "pod.babashka.go-sqlite3/get-connection":
			db, err := parseGetConnectionArgs(message.Args)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			id := uuid.New().String()
			result := map[string]interface{}{"connection": id}
			conn, err := sql.Open("sqlite3", db)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}

			syncMap.Store(id, conn)
			respond(message, result)
		case "pod.babashka.go-sqlite3/close-connection":
			connId, err := parseCloseConnectionArgs(message.Args)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			cached, ok := syncMap.Load(connId)
			if !ok {
				err := fmt.Errorf("invalid connection id: %s", connId)
				babashka.WriteErrorResponse(message, err)
			}
			syncMap.Delete(connId)
			conn := cached.(*sql.DB)
			err = conn.Close()
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			respond(message, nil)

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
			if err.Error() == "EOF" {
				log.Fatal("Unrecoverable error: EOF")
			}

			debug("Error reading message")
			debug(err)

			continue
		}

		processMessage(message)
	}
}
