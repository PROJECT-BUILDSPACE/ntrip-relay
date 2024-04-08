package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-gnss/ntrip"
)

var (
	reader, writer = io.Pipe()
	mountpoint     = "SSRA00EUH0" // Set default mountpoint
)

func main() {
	casterURL := "ntrip.gsc-europa.eu"
	casterPort := 443
	username := "sekkas"
	password := "Buildspace2023"
	ntripVersion := "Ntrip/2.0"

	sourceURL := fmt.Sprintf("https://%s:%d/%s", casterURL, casterPort, mountpoint)
	listenAddr := ":8080" // Change the port as needed

	go serveLocal(listenAddr)

	for {
		client, _ := ntrip.NewClientRequest(sourceURL)

		// Encode Basic Auth credentials
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
		client.Header.Set("Authorization", auth)

		// Set NTRIP version in the HTTP headers
		client.Header.Set("Ntrip-Version", ntripVersion)

		resp, err := http.DefaultClient.Do(client)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("client failed to connect", resp, err)

			// Log request details for debugging
			fmt.Printf("Request: %v\n", client)

			// Log response headers for debugging
			fmt.Printf("Response Headers: %v\n", resp.Header)

			time.Sleep(time.Second * 2) // Sleep for 2 seconds before retrying
			continue
		}

		fmt.Println("client connected")
		data := make([]byte, 4096)
		br, err := resp.Body.Read(data)
		for ; err == nil; br, err = resp.Body.Read(data) {
			writer.Write(data[:br])
		}

		fmt.Println("client connection died", err)
	}
}

// Serve the NTRIP data locally on the specified address
func serveLocal(addr string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received request:", r.Method, r.URL.Path)

		if r.Method == "GET" && r.URL.Path == "/" {
			// If it's a GET request to the root URL, return the mountpoint
			w.Write([]byte("BS2023"))
			return
		}

		if r.URL.Path == "/BS2023" {
			// If accessing the "/BS2023" endpoint, stream the data
			io.Copy(w, reader)
			return
		}

		// Otherwise, return 404 Not Found
		http.NotFound(w, r)
	})

	fmt.Println("Local server listening on", addr)
	http.ListenAndServe(addr, nil)
}

