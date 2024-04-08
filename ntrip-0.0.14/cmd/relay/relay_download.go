package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-gnss/ntrip"
)

func main() {
	casterURL := "ntrip.gsc-europa.eu"
	casterPort := 443
	username := "sekkas"
	password := "Buildspace2023"
	ntripVersion := "Ntrip/2.0"
	mountpoint := "SSRA00EUH0"

	sourceURL := fmt.Sprintf("https://%s:%d/%s", casterURL, casterPort, mountpoint)

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

		// Create a file to save the downloaded data
		outputFile, err := os.Create("downloaded_data.bin")
		if err != nil {
			fmt.Println("Failed to create output file:", err)
			return
		}
		defer outputFile.Close()

		// Create a multi-writer to write to both file and standard output
		multiWriter := io.MultiWriter(os.Stdout, outputFile)

		// Copy the response body (data from the mount point) to the multi-writer
		_, err = io.Copy(multiWriter, resp.Body)
		if err != nil {
			fmt.Println("Failed to save data to output file:", err)
			return
		}

		fmt.Println("Data downloaded and saved to downloaded_data.bin")
		return
	}
}

