package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

const CRLF = "\r\n"

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:9999")

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer func() {
		if cerr := listener.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		err = handle(conn)
		if err != nil {
			log.Println(err)
		}

	}
}

func handle(conn net.Conn) error {
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()

	reader := bufio.NewReader(conn)

	const delim = '\n'
	line, err := reader.ReadString(delim)

	if err != nil {
		if err != io.EOF {
			return err
		}
		log.Printf("received: %s\n", line)
		return err
	}
	log.Printf("received: %s\n", line)

	path, err := getPath(line)
	if err != nil {
		return err
	}

	switch path {
	case "/":
		err = writeIndex(conn)
	case "/transactions.csv":
		err = writeTransactionsCsv(conn)
	case "/transactions.json":
		err = writeTransactionsJson(conn)
	case "/transactions.xml":
		err = writeTransactionsXml(conn)
	default:
		err = write404(conn)
	}

	if err != nil {
		return err
	}

	return nil
}

func writeResponse(conn net.Conn, status int, strings []string, file []byte) error {
	writer := bufio.NewWriter(conn)

	_, err := writer.WriteString(fmt.Sprintf("HTTP/1.1 %d", status) + CRLF)
	if err != nil {
		return err
	}

	for i, s := range strings {
		if len(strings) == (i + 1) {
			s += CRLF
		}

		_, err = writer.WriteString(s + CRLF)

		if err != nil {
			return err
		}
	}

	if file != nil {
		_, err = writer.Write(file)
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func write404(conn net.Conn) error {
	return writeResponse(conn, 404, []string{"Connection: close"}, nil)
}

func writeIndex(conn net.Conn) error {
	file, err := includeIndexTemplate()
	if err != nil {
		return err
	}

	return writeResponse(conn, 200, []string{
		"Connection: close",
		fmt.Sprintf("Content-Length: %d", len(file)),
		"Content-Type: text/html; charset=utf-8",
	}, file)

}

func includeIndexTemplate() ([]byte, error) {
	username := "Michael"
	balance := "1 000.50"

	file, err := ioutil.ReadFile("web/template/index.html")
	if err != nil {
		return nil, err
	}

	file = bytes.ReplaceAll(file, []byte("{username}"), []byte(username))
	file = bytes.ReplaceAll(file, []byte("{balance}"), []byte(balance))
	return file, nil
}

func getPath(line string) (string, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return "", errors.New(fmt.Sprintf("invalid request line %s", line))
	}
	return parts[1], nil
}

func writeTransactionsCsv(conn net.Conn) error {
	file, err := ioutil.ReadFile("web/shared/transactions.csv")
	if err != nil {
		return err
	}

	return writeResponse(conn, 200, []string{
		"Content-Type: text/csv",
		fmt.Sprintf("Content-Length: %d", len(file)),
		"Connection: close",
	}, file)
}

func writeTransactionsJson(conn net.Conn) error {
	file, err := ioutil.ReadFile("web/shared/transactions.json")
	if err != nil {
		return err
	}

	return writeResponse(conn, 200, []string{
		"application/json; charset=utf-8",
		fmt.Sprintf("Content-Length: %d", len(file)),
		"Connection: close",
	}, file)
}

func writeTransactionsXml(conn net.Conn) error {
	file, err := ioutil.ReadFile("web/shared/transactions.xml")
	if err != nil {
		return err
	}

	return writeResponse(conn, 200, []string{
		"application/xml",
		fmt.Sprintf("Content-Length: %d", len(file)),
		"Connection: close",
	}, file)
}
