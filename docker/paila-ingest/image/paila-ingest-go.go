//
// This go project runs a http server that acts as the ingest for all log
// and spec files sent to it and queues the files for AI processing.
//
// Author: Chris Mayenschein
// GitHub: https://github.com/cmayen/paila
// Date: 2025-07-20
// Last Modified: 2025-07-20
//
// Usage: ./paila-ingest-go
//
// #############################################################################

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
  "regexp"
  "net"
)
// sudo docker exec -it new-image-name-localtest bash


// override listenPort with PAILA_INGEST_PORT env var in main func
var listenPort string = "8181"

const uploadsDir string = "/.paila-ingest/uploads"
const reportsDir string = "/.paila-ingest/reports"
const archiveDir string = "/.paila-ingest/archive"


// get the outbound up so we can output debug info on start
func GetLocalOutboundIP() (string, error) {
	// Dial a UDP connection to a well-known external address (e.g., Google DNS server).
	// The specific address and port do not need to be reachable or exist.
	// This action is solely to determine the local address used for outbound traffic.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("failed to dial UDP connection: %w", err)
	}
	defer conn.Close()

	// Retrieve the local address of the connection.
	localAddr := conn.LocalAddr()

	// Assert the type to net.UDPAddr to access the IP field.
	udpAddr, ok := localAddr.(*net.UDPAddr)
	if !ok {
		return "", fmt.Errorf("local address is not a UDP address")
	}

	// Convert the IP address to a string.
	return udpAddr.IP.String(), nil
}


// handler for the file upload
func uploadHandler(w http.ResponseWriter, r *http.Request) {


  // make sure it is a post method
	if r.Method != http.MethodPost {
    // Prepare JSON response
    response := map[string]string{
      "status": fmt.Sprintf("%d", http.StatusMethodNotAllowed),
      "message":  "Method not allowed",
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
		return
	}


	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit for form data
	if err != nil {
    // Prepare JSON response
    response := map[string]string{
      "status": fmt.Sprintf("%d", http.StatusBadRequest),
      "message":  fmt.Sprintf("Error parsing multipart form: %v", err),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
		return
	}


	// Get form fields
	hostR := r.FormValue("host")
	dateR := r.FormValue("date")


  // sanitizing
  // remove all non-alphanumeric characters except spaces
  reg := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
  host := reg.ReplaceAllString(hostR, "")
  date := reg.ReplaceAllString(dateR, "")


	// Get the "log" file
	file, handler, err := r.FormFile("log")
	if err != nil {
    // Prepare JSON response
    response := map[string]string{
      "status": fmt.Sprintf("%d", http.StatusInternalServerError),
      "message":  fmt.Sprintf("Error retrieving file 'log': %v", err),
      "filename": handler.Filename,
      "host":     host,
      "date":     date,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
		return
	}
	defer file.Close()


	// Create the directory if it doesn't exist
	//uploadDir := "./uploads" // Define your upload directory
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		err = os.Mkdir(uploadsDir, 0755)
		if err != nil {
      // Prepare JSON response
      response := map[string]string{
        "status": fmt.Sprintf("%d", http.StatusInternalServerError),
        "message":  fmt.Sprintf("Error creating upload directory: %v", err),
        "filename": handler.Filename,
        "host":     host,
        "date":     date,
      }
      w.Header().Set("Content-Type", "application/json")
      json.NewEncoder(w).Encode(response)
			return
		}
	}


	// Create a new file on the filesystem to save the uploaded content
	dstPath := filepath.Join(uploadsDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
    // Prepare JSON response
    response := map[string]string{
      "status": fmt.Sprintf("%d", http.StatusInternalServerError),
      "message":  fmt.Sprintf("Error creating destination file: %v", err),
      "filename": handler.Filename,
      "host":     host,
      "date":     date,
      "filepath": dstPath,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
		return
	}
	defer dst.Close()


	// Copy the uploaded file content to the destination file
  _, err = io.Copy(dst, file)
 	if err != nil {
    // Prepare JSON response
    response := map[string]string{
      "status": fmt.Sprintf("%d", http.StatusInternalServerError),
      "message":  fmt.Sprintf("Error copying file content: %v", err),
      "filename": handler.Filename,
      "host":     host,
      "date":     date,
      "filepath": dstPath,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
		return
	}


	// Prepare JSON response
	response := map[string]string{
    "status": "201",
		"message":  "File uploaded successfully",
		"filename": handler.Filename,
		"host":     host,
		"date":     date,
		"filepath": dstPath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

  // todo : use go func to call to script that will handle parsing
  //        sqlite database population, and the call to the 
  //        ai diagnostic tool api for the generated report
  // todo : this almost makes the setup for the dashboard to also
  //        be on this server. *shrugs* makes sense to me

}


func main() {
	http.HandleFunc("/uploadlog", uploadHandler)
  // get the ingest port from env variables if it exists
  listenPortEnv, exists := os.LookupEnv("PAILA_INGEST_PORT")
  if exists {
    listenPort = listenPortEnv
  }



	ip, err := GetLocalOutboundIP()
	if err != nil {
		//log.Fatalf("Error getting local outbound IP: %v", err)
	}

	fmt.Println("Server listening on :"+listenPort)
  fmt.Println("  ./paila-logpush.sh -u http://"+ip+":"+listenPort+"/uploadlog")
	http.ListenAndServe(":"+listenPort, nil)
}



