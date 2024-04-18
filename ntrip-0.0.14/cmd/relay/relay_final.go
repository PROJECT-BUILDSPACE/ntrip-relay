package main

import (
	"fmt"
	"io"
	"net/http"
)

var (
	casterURL    = "ntrip.gsc-europa.eu"
	casterPort   = 443
	username     = "sekkas"
	password     = "Buildspace2023"
	ntripVersion = "Ntrip/2.0S" // Update to NTRIP version 2.0S
)

func main() {
	sourceURL := fmt.Sprintf("https://%s:%d", casterURL, casterPort)
	listenAddr := ":8080" // Change the port as needed

	// Start the local NTRIP server
	go serveLocalNTRIP(listenAddr, sourceURL)

	// Keep the main function running
	select {}
}

// Serve the NTRIP data locally as an NTRIP caster
func serveLocalNTRIP(addr string, sourceURL string) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Received connection from:", req.RemoteAddr)
		fmt.Println("Request URL:", req.URL.Path)

		// Extract the mountpoint from the request URL
		mountpoint := req.URL.Path

		// Construct the URL of the remote NTRIP caster for the specific mountpoint
		remoteURL := fmt.Sprintf("%s%s", sourceURL, mountpoint)

		// Make a request to the remote NTRIP caster
		client := &http.Client{}
		reqToCaster, err := http.NewRequest("GET", remoteURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println("Error creating request to remote NTRIP caster:", err)
			return
		}

		// Set Basic Auth credentials
		reqToCaster.SetBasicAuth(username, password)

		// Log the request details
		fmt.Println("Request to remote NTRIP caster details:")
		fmt.Println("  URL:", remoteURL)
		fmt.Println("  Username:", username)
		fmt.Println("  Password:", password) // Print the actual password
		fmt.Println("  NTRIP Version:", ntripVersion)

		// Include other headers from the original request (if any)
		for header, values := range req.Header {
			for _, value := range values {
				reqToCaster.Header.Add(header, value)
			}
		}

		// Log the headers being sent to the remote NTRIP caster
		fmt.Println("  Other Headers:")
		for header, values := range reqToCaster.Header {
			for _, value := range values {
				fmt.Printf("    %s: %s\n", header, value)
			}
		}

		resp, err := client.Do(reqToCaster)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println("Error making request to remote NTRIP caster:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println("Forwarding request to remote NTRIP caster:", remoteURL)

		// Log the response status code
		fmt.Println("Response Status:", resp.Status)

		// Copy data from the response body to the local NTRIP client
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			fmt.Println("Error copying response data to local NTRIP client:", err)
			return
		}
	})

	// Start the HTTP server
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Failed to start HTTP server:", err)
		return
	}

	fmt.Println("Local NTRIP server listening on", addr)
}
