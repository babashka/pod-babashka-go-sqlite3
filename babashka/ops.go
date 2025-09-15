package babashka

import (
	"bufio"
	"fmt"
	"github.com/jackpal/bencode-go"
	"os"
)

func debug(v interface{}) {
	fmt.Fprintf(os.Stderr, "debug: %+q\n", v)
}

type Message struct {
	Op   string
	Id   string
	Args string
	Var  string
}

type Namespace struct {
	Name string "name"
	Vars []Var  "vars"
}

type Var struct {
	Name string "name"
	Code string `bencode:"code,omitempty"`
}

type DescribeResponse struct {
	Format     string      "format"
	Namespaces []Namespace "namespaces"
}

type InvokeResponse struct {
	Id     string   "id"
	Value  string   "value" // stringified json response
	Status []string "status"
}

type ErrorResponse struct {
	Id        string   "id"
	Status    []string "status"
	ExMessage string   "ex-message"
	ExData    string   "ex-data,omitempty"
}

func ReadMessage() (*Message, error) {
	reader := bufio.NewReader(os.Stdin)
	message := &Message{}
	if err := bencode.Unmarshal(reader, &message); err != nil {
		return nil, err
	}

	return message, nil
}

func WriteDescribeResponse(describeResponse *DescribeResponse) {
	writeResponse(*describeResponse)
}

func WriteInvokeResponse(inputMessage *Message, value string) error {
	response := InvokeResponse{Id: inputMessage.Id, Status: []string{"done"}, Value: value}

	return writeResponse(response)
}

func WriteErrorResponse(inputMessage *Message, err error) {
	errorMessage := string(err.Error())

	id := "error"
	if inputMessage != nil {
		id = inputMessage.Id
	}

	errorResponse := ErrorResponse{
		Id:        id,
		Status:    []string{"done", "error"},
		ExMessage: errorMessage,
	}
	writeResponse(errorResponse)
}

func writeResponse(response interface{}) error {
	writer := bufio.NewWriter(os.Stdout)
	if err := bencode.Marshal(writer, response); err != nil {
		return err
	}

	writer.Flush()

	return nil
}
