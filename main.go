package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const oneMB = 1024 * 1024
const oneGB = 1024 * oneMB
const responseSize = 2 * oneGB

const serverAddr = "localhost:9999"

type Message struct {
	Id      int    `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	HandleHttp(w, r)
}

func HandleHttp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")

	w.(http.Flusher).Flush()

	for i := 1; ; i++ {
		select {
		case <-r.Context().Done():
			return
		default:
			data := fmt.Sprintf("Message type 1 %d\n", i)

			_, err := fmt.Fprint(w, data)
			if err != nil {
				fmt.Println("Error writing data:", err)
				return
			}
			fmt.Println("Sent message", i)

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			} else {
				fmt.Println("Flush not supported!")
			}

			time.Sleep(3 * time.Second)
		}
	}
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=example.txt")
	w.Header().Set("Content-Type", "application/octet-stream")
	datach := make(chan int)
	go func() {
		for i := 0; ; i++ {
			time.Sleep(1 * time.Second)
			datach <- i
		}
	}()

	for i := range datach {
		msg := fmt.Sprintf("data received type 2 %d", i)
		_, err := fmt.Fprint(w, msg)
		if err != nil {
			fmt.Println("Error writing data:", err)
			return
		}

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

}

func HandleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")

	i := 1
	for {
		select {
		case <-r.Context().Done():
			fmt.Printf("Stream closed by client %s\n", r.Host)
			return
		default:
			msg := Message{Message: fmt.Sprintf("data received type 1 %d", i)}
			bytes, err := json.Marshal(msg)
			if err != nil {
				return
			}
			bytes = append(bytes, '\n')
			_, err = w.Write(bytes)
			if err != nil {
				return
			}
			w.(http.Flusher).Flush()

			i++
			fmt.Printf("data sent:%d  time\n", i)
			time.Sleep(1 * time.Second)
		}
	}
}

func Status(w http.ResponseWriter, r *http.Request) {
	m := Message{
		Message: fmt.Sprintf("ok"),
	}
	buff, err := json.Marshal(m)
	if err != nil {
		return
	}
	w.Write(buff)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter a port to dial tcp conn: ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)
	_, err := strconv.Atoi(port)
	if err != nil {
		fmt.Printf("Invalid port - Error: %v\n", err)
		return
	}
	port = ":" + port
	http.HandleFunc("/", Status)
	fmt.Printf("Starting :%s/\n", port)
	http.HandleFunc("/receiver", HandleStream)
	fmt.Printf("Starting :%s/receiver\n", port)
	http.HandleFunc("/file", HandleDownload)
	fmt.Printf("Starting :%s/file\n", port)
	http.ListenAndServe(port, nil)
	fmt.Printf("Started on %s\n", port)
}
