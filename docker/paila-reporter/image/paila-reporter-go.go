//
// This go project runs a http server for the reporter.
//
// Author: Chris Mayenschein
// GitHub: https://github.com/cmayen/paila
// Date: 2025-07-21
// Last Modified: 2025-07-23
//
// Usage: ./paila-reporter-go
//
// #############################################################################

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var publicDir string = "/.paila-reporter/public"
var directoryToScan string = "/.paila-ingest" // Change this to the paila-ingest base folder
var walkFolders [3]string = [3]string{
	"uploads",
	"reports",

	"archive",
}

var ollamaApiGenerateUrl string = "http://localhost:11434/api/generate"

// The content below consists of two sections: The first section titled
// "Logged Issues Report" is the log information to be reviewed. The second section titled
// "System Information Report" contains system information only to be used for helping diagnose the logs.

var ollamaInstructions string = `You are a devops system administrator in charge of monitoring logs for issues and suggesting resolutions. Go through all of the following log information, generate a detailed report about the issues found, and include suggestions for resolutions of the issues.
Do not explain what each log file is for. Provide a summary of issues and stay focused on explaining those issues with examples of resolutions.
---\n`

func main() {
	ollamaApiGenerateUrlEnv, exists := os.LookupEnv("PAILA_OLLAMA_URL")
	if exists {
		ollamaApiGenerateUrl = ollamaApiGenerateUrlEnv
	}

	mux := http.NewServeMux()

	finalHandler := http.HandlerFunc(final)
	mux.Handle("/", Middleware(finalHandler))

	mux.Handle("/report-data", Middleware(http.HandlerFunc(reportDataHandler)))
	mux.Handle("/report-generate", Middleware(http.HandlerFunc(reportGenerateHandler)))

	server := &http.Server{
		Addr:           ":80",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   960 * time.Second, // increased to allow for long processing times
		MaxHeaderBytes: 1 << 20,
		// BaseContext:    func(_ net.Listener) context.Context { return context.TODO() },
	}
	err := server.ListenAndServe()
	//err := server.ListenAndServeTLS("server.crt", "server.key")
	log.Fatal(err)
	/*
		filePath := "basic-layout-and-structure-for-tech-blog-pt-1.html" // Replace with your file path
		htmlString, err := parseTemplateForFile(filePath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(htmlString)
	*/

}
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Print("        Middleware: " + r.RequestURI + " (before)")
		next.ServeHTTP(w, r) //.WithContext(ctx))
		//log.Print("        Middleware: " + r.RequestURI + " (after)")
	})
}

func final(w http.ResponseWriter, r *http.Request) {
	//user, ok := auth.CurrentUser(r)
	//user, ok := r.Context().Value(userKey{}).(*ContextUser) // Cast to your User type
	//if ok {
	//	log.Println(user.LoginName + "nice")
	//}
	// we didnt intecept anything yet, so check if a file exists, and send that
	err := fileServerWithMime(w, r)
	if err != nil {
		// intercept the 404 error and check for a file in place. This is more for the blog thing but this
		// website is a shti version of ssi anyways
		if err.Error() == "404:NotFound" {
			//log.Println("\n\n\n 404:NotFound:" + r.RequestURI + " \n\n\n ")
			// check the path and setup for the index document if the extended
			// path data is empty ( not just "public" or what publicDir was set to )
			path := filepath.Join(publicDir, filepath.Clean(r.URL.Path))
			log.Println(path + "\n ")
			if path == publicDir {
				r.URL.Path = "index"

				s, err := pailaIndexContent(w, r)
				if err != nil {
					// error handling index document content
				} else {
					s, err := parseTemplateForString(w, r, s)
					if err != nil {
						w.WriteHeader(http.StatusNotFound)
						//log.Println("443 404 " + r.RequestURI)
						http.ServeFile(w, r, filepath.Join(publicDir, "404.html"))
					}
					w.Header().Set("Content-Type", "text/html")
					w.Write([]byte(s))
				}

			} else {

				fP, err := lookForFileToParse(w, r)
				if err != nil && err.Error() == "404:NotFound" {
					w.WriteHeader(http.StatusNotFound)
					//log.Println("443 404 " + r.RequestURI)
					http.ServeFile(w, r, filepath.Join(publicDir, "404.html"))
				} else {
					// parse the template-able file
					s, err := parseTemplateForFile(w, r, fP)
					if err != nil {
						w.WriteHeader(http.StatusNotFound)
						//log.Println("443 404 " + r.RequestURI)
						http.ServeFile(w, r, filepath.Join(publicDir, "404.html"))
					} else {
						//

						//
						//log.Println("443 200 " + r.RequestURI + " (" + fP + ")")
						w.Header().Set("Content-Type", "text/html")
						w.Write([]byte(s))
						//http.ServeFile(w, r, path)
						//

						//
					}

				}
			}
		}
	}
	//w.Write([]byte("OK"))
}

func lookForFileToParse(w http.ResponseWriter, r *http.Request) (string, error) {
	path := filepath.Join(publicDir, filepath.Clean(r.URL.Path))
	path = strings.TrimSuffix(path, "/")
	path = path + ".html"
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		// file found
		// make sure the url does NOT end with a slash, redirect to one without
		if strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "/"), http.StatusMovedPermanently)
			return "", errors.New("301:StatusMovedPermanently")
		}
		//log.Println("Found:" + path)
		return path, nil
	} else {
		//log.Println("NotFOund:" + path)
		return "", errors.New("404:NotFound")
	}
}

func parseTemplateForFile(w http.ResponseWriter, r *http.Request, filePath string) (string, error) {
	htmlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	return parseTemplateForString(w, r, string(htmlBytes))
}

func parseTemplateForString(w http.ResponseWriter, r *http.Request, htmlString string) (string, error) {
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlString))
	if err != nil {
		log.Fatal("Error loading HTML: ", err)
	}
	// Find the meta description tag
	description := doc.Find("meta[name='description']").AttrOr("content", "Meta description not found.")
	// now that we have the description, remove that node from the files head
	doc.Find("head>meta[name='description']").Remove()
	// Find the meta keywords tag
	keywords := doc.Find("meta[name='keywords']").AttrOr("content", "Meta keywords not found.")
	// now that we have the keywords, remove that node from the files head
	doc.Find("head>meta[name='keywords']").Remove()
	// the keywords from the meta are what will be used to generate the tags for the blog post
	// this then can be added to the database to lookup by tag when the page is loaded if
	// there is no keyword/tags data in there yet

	// Find the meta keywords tag
	author := doc.Find("meta[name='author']").AttrOr("content", "")
	// now that we have the keywords, remove that node from the files head
	doc.Find("head>meta[name='author']").Remove()

	// find any div.uri elements and put the request uri in there, (sanitized of course)
	cDate, _ := doc.Find("meta[name='date']").Attr("content")
	sDate, err := time.Parse("2006-01-02 15:04:05 Z0700", cDate) // https://dev.to/luthfisauqi17/golangs-unique-way-to-parse-string-to-time-2jmk // hr=03=12hr hr=15=24hr
	if err != nil {
		log.Println(err.Error())
	}
	//log.Println("sDate.String():" + sDate.String())
	// find the date element and put it into place formatted, along with the author name?
	divDateCs := doc.Find("article.blog-article div.date")
	divDateCs.SetHtml(html.EscapeString(sDate.Format("January 2, 2006")))

	// build up content for the tags container
	var hT strings.Builder
	keywordsA := strings.Split(keywords, " ")
	for index, value := range keywordsA {
		//fmt.Println("Index:", index, "Value:", value)
		if index > 0 {
		}
		hT.WriteString(fmt.Sprintf(`<a href="/blog/tags/%s">%s</a>`, value, value))
	}
	// find the tags containers in the blog post and populate them with the keywords found
	tagsContainers := doc.Find("article.blog-article div span.tags")
	tagsContainers.SetHtml(hT.String())

	// find any div.uri elements and put the request uri in there, (sanitized of course)
	divUriCs := doc.Find("div.uri")
	divUriCs.SetHtml(html.EscapeString(r.RequestURI))
	//

	// locate the previous and next links that target sibling html files and
	// remove the .html
	/*
		htmlSiblingLinks := doc.Find("article.blog-article>div.series-links>a")
		htmlSiblingLinks.Each(func(i int, s *goquery.Selection) {
			// Get the href attribute
			href, exists := s.Attr("href")
			log.Println("href=" + href)
			if exists {
				// Check if the href ends with ".html"
				if strings.HasSuffix(href, ".html") {
					// Remove ".html" from the end
					newHref := strings.TrimSuffix(href, ".html")
					// Set the modified href attribute
					s.SetAttr("href", newHref)
				}
			}
		})
	*/
	//

	//
	// <time itemprop="startDate" datetime="2009-10-15T19:00-08:00">15 October 2009, 7PM</time>

	// get the title
	title := doc.Find("head>title").Text()
	// remove the title from the head
	doc.Find("head>title").Remove()
	// get what is left of the head into a string
	headTxt, err := doc.Find("head").Html()
	if err != nil {
		log.Fatal(err)
	}
	// =================================================================================
	// =================================================================================

	htmlBytesT, err := os.ReadFile(publicDir + "/template.html")
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	// Load the HTML document
	docT, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlBytesT)))
	if err != nil {
		log.Fatal(err)
	}
	// get the title
	titleT := docT.Find("head>title").Text()
	// prepend the post title to the site title
	docT.Find("head>title").SetHtml(title + " - " + titleT)

	// if there is a description, update it, else add one
	docT.Find("meta[name='description']").SetAttr("content", description)

	// if there is a description, update it, else add one
	if author != "" {
		docT.Find("meta[name='author']").SetAttr("content", author)
	}

	// add what is left of the head content from the page, and insert it into the tempaltge
	//docT.Find("head").Append(headTxt)
	docT.Find("head").AppendHtml(headTxt)

	//
	//
	//
	//
	bh, err := doc.Find("body").Html()
	if err != nil {
		log.Fatal(err)
	}
	docT.Find("body>main>div").SetHtml(bh)
	//
	//
	//
	//

	h, err := docT.Html()
	if err != nil {
		log.Fatal(err)
	}
	return h, nil
}

// Serve static files with OS-based MIME types
// func fileServerWithMime(publicDir string) http.HandlerFunc {
// return func(w http.ResponseWriter, r *http.Request) {
func fileServerWithMime(w http.ResponseWriter, r *http.Request) error {

	path := filepath.Join(publicDir, filepath.Clean(r.URL.Path))

	if path == publicDir {
		path = path + "/index"
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		ext := filepath.Ext(path)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		//log.Println("443 200 " + r.RequestURI + " (" + path + ")")
		w.Header().Set("Content-Type", mimeType)
		http.ServeFile(w, r, path)
		return nil
	} else {
		/*
			w.WriteHeader(http.StatusNotFound)
			log.Println("443 404 " + r.RequestURI + " (" + path + ")")
			http.ServeFile(w, r, filepath.Join(publicDir, "404.html"))
		*/
		return errors.New("404:NotFound")
	}
	// Custom 404
	//w.WriteHeader(http.StatusNotFound)
	//http.ServeFile(w, r, filepath.Join(publicDir, "404.html"))
}

//

// paila index content
func pailaIndexContent(w http.ResponseWriter, r *http.Request) (string, error) {
	// collect host names and dates from the files in the ingress filesystem
	hostMap := make(map[string][]string)
	//
	/*
		walkFolders := [3]string{
			"uploads",
			"reports",
			"archive",
		}
	*/
	// walk the folders looking for .logs.txt files to generate host and date lists.
	for i := 0; i < len(walkFolders); i++ {
		// dir
		dirScan := directoryToScan + "/" + walkFolders[i]
		// walk the uploads
		errwd := filepath.WalkDir(dirScan, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err // deal with errors during traversal
			}
			// Check if it's a regular file and ends with .logs.txt
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".logs.txt") {
				// split up the name to get the host and date
				nameParts := strings.Split(d.Name(), "--")
				nameParts[1] = strings.Replace(nameParts[1], ".logs.txt", "", -1)
				if _, ok := hostMap[nameParts[0]]; !ok {
					hostMap[nameParts[0]] = []string{}
				}
				hostMap[nameParts[0]] = appendIfNotFound(hostMap[nameParts[0]], nameParts[1])
			}
			return nil
		})
		if errwd != nil {
			// log.Fatalf("Error walking the directory: %v", errwd)
			return "", errors.New("500:Error walking the directory")
		}
	}

	h := "<!doctype html><html><head><script src=\"/hostmap_ui.js\"></script>"
	jsonString, errJ := json.Marshal(hostMap)
	if errJ != nil {
		return "", errors.New("500:Error Marshal the hostMap")
	}
	h += "</head><body>"
	h += "<div>"
	h += "<label for=\"select-host\">Host:</label> <select id=\"select-host\" onchange=\"hostmap_ui_update()\"></select>&nbsp;"
	h += "<label for=\"select-date\">Date:</label> <select id=\"select-date\" onchange=\"hostmap_ui_update()\"></select>"
	h += "</div><div id=\"paila_log_content\">"
	//

	//
	h += "</div>"
	//
	h += "<script>var hostMap=" + string(jsonString) + ";hostmap_ui(hostMap);</script>"
	return h + "</body></html>", nil
}

// appendIfNotFound appends a string to a slice only if it's not already present.
func appendIfNotFound(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice // String found, return original slice
		}
	}
	return append(slice, s) // String not found, append it
}

func reportDataHandler(w http.ResponseWriter, r *http.Request) {

	contentLogs := ""
	contentSpecs := ""
	contentReport := ""

	// get the host and date values from the url
	params := r.URL.Query()

	// sanitizing: remove all non-alphanumeric characters except hyphens and periods
	reg := regexp.MustCompile(`[^a-zA-Z0-9.-]`)
	pHost := reg.ReplaceAllString(params.Get("host"), "")
	pDate := reg.ReplaceAllString(params.Get("date"), "")

	fmt.Printf("reportDataHandler: pHost=%s pDate=%s \n", pHost, pDate)

	//

	//

	// look for the files
	for i := 0; i < len(walkFolders); i++ {
		filePath := directoryToScan + "/" + walkFolders[i] + "/" + pHost + "--" + pDate + ".logs.txt"

		if fileExists(filePath) {
			contentBytes, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file '%s': %v\n", filePath, err)
				//return
			} else {
				fileContent := string(contentBytes)
				splitContent := strings.Split(fileContent, "============================================\n= Begin System Information Report")
				// if
				//splitContent[1] = "============================================\n= Begin System Information Report" + splitContent[1]
				contentLogs = splitContent[0]
				contentSpecs = "" // splitContent[1]
				//return
				fmt.Printf("read success file '%s'\n", filePath)
			}
		}

	}
	//

	//

	//
	// now look for an ai generated report
	filePath := directoryToScan + "/reports/" + pHost + "--" + pDate + ".report.txt"

	if fileExists(filePath) {
		contentBytes, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file '%s': %v\n", filePath, err)
			//return
		} else {
			contentReport = string(contentBytes)
			//return
			fmt.Printf("read success file '%s'\n", filePath)
		}
	}

	retJson := map[string]string{
		"host":   pHost,
		"date":   pDate,
		"logs":   contentLogs,
		"specs":  contentSpecs,
		"report": contentReport,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(retJson)

}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false // File does not exist
		}
		// Handle other potential errors (e.g., permissions)
		fmt.Printf("Error checking file %s: %v\n", filename, err)
		return false // Or handle as an error case
	}
	return true // File exists
}

//

//

//

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"` // Set to false for a single response
}

//

//

type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// mux.Handle("/report-generate", Middleware(http.HandlerFunc(reportGenerateHandler)))
func reportGenerateHandler(w http.ResponseWriter, r *http.Request) {

	// get the host and date values from the url
	params := r.URL.Query()

	// sanitizing: remove all non-alphanumeric characters except hyphens and periods
	reg := regexp.MustCompile(`[^a-zA-Z0-9.-]`)
	pHost := reg.ReplaceAllString(params.Get("host"), "")
	pDate := reg.ReplaceAllString(params.Get("date"), "")

	// make sure the raw source for the report exists.

	fileContent := ""
	// look for the files
	for i := 0; i < len(walkFolders); i++ {
		filePath := directoryToScan + "/" + walkFolders[i] + "/" + pHost + "--" + pDate + ".logs.txt"
		if fileExists(filePath) {
			contentBytes, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file '%s': %v\n", filePath, err)
				//return
			} else {
				fileContent = string(contentBytes)
				//fmt.Printf("read success file '%s'\n", filePath)
			}
		}
	}
	//

	if fileContent != "" {

		//

		//

		//
		requestBody := OllamaRequest{
			Model:  "gemma3", // Replace with your desired model
			Prompt: ollamaInstructions + fileContent,
			Stream: false, // Set to false to get a single, complete response
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			log.Fatalf("Error marshalling request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Error marshalling request body: " + err.Error()))
		}

		//client := &http.Client{Timeout: 5 * time.Second}
		resp, err := http.Post(ollamaApiGenerateUrl, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			log.Fatalf("Error making HTTP request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Error making HTTP request: " + err.Error()))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Fatalf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fmt.Sprintf("API request failed with status: %d%s", resp.StatusCode, string(bodyBytes))))
		}

		// Read the entire response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fmt.Sprintf("Error reading response body: %v", err)))
		}

		var responseData OllamaResponse
		err = json.Unmarshal(bodyBytes, &responseData)
		if err != nil {
			log.Fatalf("Error unmarshalling response body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fmt.Sprintf("Error unmarshalling response body:: %v", err)))
		}

		fmt.Println("Generated Response:", responseData.Response)

		reportPath := directoryToScan + "/reports/" + pHost + "--" + pDate + ".report.txt"

		err = os.WriteFile(reportPath, []byte(responseData.Response), 0644) // 0644 sets file permissions
		if err != nil {
			log.Fatal(err)
			retJson := map[string]string{"success": "0"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(retJson)
			return
		}
		retJson := map[string]string{"success": "1", "report": responseData.Response}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(retJson)
		return
	} else {
		retJson := map[string]string{"success": "0"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(retJson)
	}
}

//
