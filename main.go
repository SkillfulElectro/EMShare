package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

const (
	startPort = 1000
	endPort   = 60000
)

func main() {
	// Parse command-line arguments
	sendPath := flag.String("send", "", "Path to the file to send")
	receivePath := flag.String("receive", "", "Path to save the received file")
	flag.Parse()

	if *sendPath != "" {
		startServerToSend(*sendPath)
	} else if *receivePath != "" {
		startServerToReceive(*receivePath)
	} else {
		fmt.Println("Start the app this way:")
		fmt.Println("app -send /path/to/file")
		fmt.Println("or")
		fmt.Println("app -receive /path/to/save/the/file")
	}
}

// startServerToSend sets up an HTTP server to serve a file and find an available port
func startServerToSend(filePath string) {
	port, err := findAvailablePort()
	if err != nil {
		fmt.Println("failed to send")
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	})

	url := fmt.Sprintf("http://%s:%d/%s", getLANIPAddress(), port, filepath.Base(filePath))
	fmt.Printf("File available at: %s\n", url)

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	fmt.Println("Starting server...")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("failed to send")
		}
	}()


}

// startServerToReceive sets up an HTTP server to receive a file and serve an HTML page
func startServerToReceive(savePath string) {
	port, err := findAvailablePort()
	if err != nil {
		fmt.Println("error")
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/upload" {
			file, _, err := r.FormFile("file")
			if err != nil {
				fmt.Println("error")
				http.Error(w, "Unable to process file", http.StatusBadRequest)
				return
			}
			defer file.Close()

			outFile, err := os.Create(savePath)
			if err != nil {
				fmt.Println("error Unable to save file" , err)
				http.Error(w, "Unable to save file", http.StatusInternalServerError)
				return
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, file)
			if err != nil {
				fmt.Println("error while saving file" , err)
				http.Error(w, "Error while saving file", http.StatusInternalServerError)
				return
			}
			fmt.Println("received and saved")
			return
		}

		// Serve the static index.html for all other GET requests
		http.ServeFile(w, r, "static/index.html")
	})

	url := fmt.Sprintf("http://%s:%d", getLANIPAddress(), port)
	fmt.Printf("Server available at: %s\n", url)

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	fmt.Println("Starting server...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("error " , err)
	}
}

// findAvailablePort looks for an available port between startPort and endPort
func findAvailablePort() (int, error) {
	for port := startPort; port <= endPort; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found")
}

// getLANIPAddress returns the local IP address of the machine
func getLANIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("error")
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "localhost"
}

