package main

import (
	"bufio"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var dataDir = flag.String("dataDir", "/data", "path to backends")
var backends map[string]string

func initBackends() {
	var backendName string

	backends = make(map[string]string)

	fDataDir, err := os.Open(*dataDir)
	if err != nil {
		log.Fatalf("Failed opening directory: %s", err)
	}
	defer fDataDir.Close()

	for {
		list, err := fDataDir.Readdirnames(10)
		for _, name := range list {
			if !strings.HasPrefix(name, "cgp") {
				continue
			}

			backendName = getBackendNameByPath(&name)
			backends[backendName] = name
			fmt.Printf("%s -> %s\n", backendName, name)
		}
		if err == io.EOF {
			break
		}
	}

	if len(backends) == 0 {
		log.Fatalf("Couldn't found cgp directories in directory %s", *dataDir)
	}
}

var reDomainName = regexp.MustCompile("[ \t]+DomainName[ ]+=[ ]+([a-zA-Z\\.\\-0-9]+);")

func getBackendNameByPath(pathName *string) string {
	var filename = fmt.Sprintf("%s/%s/Settings/Main.settings", *dataDir, *pathName)
	fSettings, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed opening settings file: %s", err)
	}
	defer fSettings.Close()

	var foundStrings []string
	scanner := bufio.NewScanner(fSettings)
	for scanner.Scan() {
		foundStrings = reDomainName.FindStringSubmatch(scanner.Text())
		if foundStrings == nil {
			continue
		}
		return foundStrings[1]
	}
	if err := scanner.Err(); err != nil {
		log.Printf("WARN: reading file %s input: %s", filename, err)
	}

	return ""
}

var re = regexp.MustCompile("[ \t]+DomainName[ ]+=[ ]+([a-zA-Z\\.\\-0-9]+);")

func cGPHandleRequest(w http.ResponseWriter, req *http.Request) {
	// get backend, domain, email, filename
	// example: /getfile/test1.intranet.ru/test.ru/test@test.ru/test-directory/test-file.txt
	var params = strings.SplitN(req.URL.Path, "/", 3)
	switch {
	case len(params) < 2:
		// example: 123
		w.WriteHeader(400)
	case len(params) == 2:
		// example: /test
		handleCommand(w, req, params[1], "")
	case len(params) == 3:
		// example: /test/1234
		handleCommand(w, req, params[1], params[2])
	default:
		w.WriteHeader(400)
	}

	log.Printf("%q\n", html.EscapeString(req.URL.Path))
}

func handleCommand(w http.ResponseWriter, req *http.Request, command string, path string) {
	switch command {
	case "getfile":
		handleCommandGetFile(w, req, path)
	default:
		w.WriteHeader(400)
	}
}

func handleCommandGetFile(w http.ResponseWriter, req *http.Request, path string) {

	var err error

	// example: /getfile/test1.intranet.ru/test.ru/test@test.ru/test-directory/test-file.txt
	var params = strings.SplitN(path, "/", 4)

	if len(params) < 4 || params[3] == "" {
		returnError(w, 404, fmt.Sprintf("Bad path: %s\n", path))
		return
	}

	var backendPath, ok = backends[params[0]]
	if !ok {
		returnError(w, 404, fmt.Sprintf("Backend not found: '%s'\n", params[0]))
		return
	}

	var domain = params[1]
	if domain == "" {
		returnError(w, 404, fmt.Sprintf("Domain not found in request: '%s'\n", domain))
		return
	}

	var email = params[2]
	if email == "" {
		returnError(w, 404, fmt.Sprintf("Email not found in request: '%s'\n", email))
		return
	}

	var filePath = params[3]
	if filePath == "" {
		returnError(w, 404, fmt.Sprintf("File path is empty in request: '%s'\n", filePath))
		return
	}

	// get path to backend by backend name
	var emailPrefix string
	var emailLocalPart string

	emailPrefix = fmt.Sprintf("%s.sub/%s.sub", string(email[0]), string(email[1]))
	emailLocalPart = strings.SplitN(email, "@", 2)[0]

	var fullPath string

	fullPath = fmt.Sprintf("%s/%s/Domains/%s/%s/%s.macnt/%s", *dataDir, backendPath, domain, emailPrefix, emailLocalPart, filePath)

	// check file exists
	f, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Failed opening requested file: %s", err)
		w.WriteHeader(404)
		w.Header().Add("Content-type", "text/plain")
		fmt.Fprintf(w, "Failed open requested file\n")
		return
	}
	defer f.Close()

	// return content like text/plain in utf-8
	for err != io.EOF {
		_, err = io.CopyN(w, f, 1024)
	}
}

func returnError(w http.ResponseWriter, statusCode int, err string) {
	w.WriteHeader(statusCode)
	w.Header().Add("Content-type", "text/plain")
	fmt.Fprint(w, err)
}
