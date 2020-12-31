package pod

import (
	"container/list"
	"database/sql"
	"fmt"
	"github.com/babashka/pod-babashka-sqlite3/babashka"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
	"github.com/russolsen/transit"
	"os"
	"strings"
)

func encodeRows(rows *sql.Rows) ([]interface{}, error) {
	cols, err := rows.Columns()
	columns := make([]transit.Keyword, len(cols))
	for i, col := range cols {
		columns[i] = transit.Keyword(col)
	}
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	c := 0
	results := make(map[transit.Keyword]interface{})
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

func encodeResult(result sql.Result) (map[transit.Keyword]int64, error) {
	rowsAffected, err := result.RowsAffected()
	lastInsertedId, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	res := map[transit.Keyword]int64{
		transit.Keyword("rows-affected"): rowsAffected,
		transit.Keyword("last-inserted-id"): lastInsertedId,
	}
	return res, nil
}

func debug(v interface{}) {
	fmt.Fprintf(os.Stderr, "debug: %+v\n", v)
}

func listToSlice(l *list.List) []interface{} {
	var slice []interface{}
	slice = make([]interface{}, l.Len())
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
	var theList *list.List
	theList = value.(*list.List)
	theSlice := listToSlice(theList)

	var db string
	db = theSlice[0].(string)
	// println("db", db)

	var queryArgs []interface{}
	queryArgs = theSlice[1].([]interface{})

	var query string
	query = queryArgs[0].(string)

	return db, query, queryArgs[1:], nil
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

		switch message.Var {
		case "pod.babashka.sqlite3/execute!":
			res, err := conn.Exec(query, args...)
			if err != nil {
				return nil, err
			}

			if json, err := encodeResult(res); err != nil {
				return nil, err
			} else {
				return json, nil
			}
		case "pod.babashka.sqlite3/query!":
			res, err := conn.Query(query, args...)
			if err != nil {
				return nil, err
			}

			if json, err := encodeRows(res); err != nil {
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
