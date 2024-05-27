package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"main/data"
	"main/tools"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

// Timeouts and cancellation mechanism
func waitServer(url string, duration time.Duration) bool {
	deadline := time.Now().Add(duration)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func main() {
	if !waitServer("https://localhost:2222", 5 * time.Second) {
		fmt.Println("Server inactive")
		return
	}
	var choice int

	for {
		fmt.Println("Main Menu")
		fmt.Println("1. Get message")
		fmt.Println("2. Send file")
		fmt.Println("3. Quit")
		fmt.Print(">> ")
		fmt.Scanf("%d\n", &choice)
		if choice == 1 {
			getMessage()
		} else if choice == 2 {
			sendFile()
		} else if choice == 3 {
			break
		} else {
			fmt.Println("Invalid choice")
		}
	}
}

func getMessage() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get("https://localhost:2222")
	tools.ErrorHandler(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	tools.ErrorHandler(err)
	fmt.Println("Server: ", string(data))

	tlsDetails(resp)
}

func sendFile() {
	var name string
	var age int

	scanner := bufio.NewReader(os.Stdin)

	fmt.Print("Input name: ")
	name, _ = scanner.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Input age: ")
	fmt.Scanf("%d\n", &age)

	// Posting JSON
	person := data.Person{Name: name, Age: age}
	jsonData, err := json.Marshal(person)
	tools.ErrorHandler(err)

	// Posting multipart form
	temp := new(bytes.Buffer)
	w := multipart.NewWriter(temp)

	personField, err := w.CreateFormField("Person")
	tools.ErrorHandler(err)

	_, err = personField.Write(jsonData)
	tools.ErrorHandler(err)

	file, err := os.Open("./file.txt")
	tools.ErrorHandler(err)
	defer file.Close()

	fileField, err := w.CreateFormFile("File", file.Name())
	tools.ErrorHandler(err)

	_, err = io.Copy(fileField, file)
	tools.ErrorHandler(err)

	err = w.Close()
	tools.ErrorHandler(err)

	req, err := http.NewRequest("POST", "https://localhost:2222/sendFile", temp)
	tools.ErrorHandler(err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	tools.ErrorHandler(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	tools.ErrorHandler(err)

	fmt.Println("Server: ", string(data))

	tlsDetails(resp)
}

// TCP Dial
func tlsDetails(resp *http.Response) {
	state := resp.TLS
	if state == nil {
		fmt.Println("TLS not found")
		return
	}

	fmt.Printf("TLS Version: %s\n", tlsVersion(state.Version))
	fmt.Printf("Cipher Suite: %s\n", tls.CipherSuiteName(state.CipherSuite))

	if len(state.PeerCertificates) > 0 {
		issuer := state.PeerCertificates[0].Issuer.Organization
		fmt.Printf("Issuer Organization: %s\n", strings.Join(issuer, ", "))
	} else {
		fmt.Println("Peer certificates not found")
	}
}

func tlsVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}
