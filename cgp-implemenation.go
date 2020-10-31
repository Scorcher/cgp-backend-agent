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

func cGPHandleRequest(w http.ResponseWriter, req *http.Request) {
	// get backend, domain, email, filename

	// get path to backend by backend name

	// check file exists

	// return content like text/plain in utf-8

	// log request

	fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(req.URL.Path))
	log.Printf("%q\n", html.EscapeString(req.URL.Path))
}
