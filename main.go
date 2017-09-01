package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
)

var (
	versionURL = flag.String("version-url", "", "URL to fetch the latest available version")
	delay      = flag.Duration("delay", 30*time.Minute, "The delay between version checks")

	currentVersion *semver.Version
)

func main() {
	flag.Parse()

	if *versionURL == "" {
		log.Fatal("I require a version-url parameter")
	}

	file, err := os.Open("/etc/os-release")
	if err != nil {
		log.Fatal(err)
		return
	}
	currentVersion, err = findVar("VERSION", file)
	file.Close()
	if err != nil {
		log.Fatal(err)
	}

	for {
		checkVersion()
		time.Sleep(*delay)
	}
}

func checkVersion() {
	log.Print("Checking ", *versionURL)

	resp, err := http.Get(*versionURL)
	if err != nil {
		log.Print("Failed to fetch: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Print("HTTP server replied ", resp.Status)
		return
	}

	version, err := findVar("COREOS_VERSION", resp.Body)
	if err != nil {
		log.Print(err)
		return
	}

	if !currentVersion.LessThan(*version) {
		log.Print("Not updated")
		return
	}

	log.Print("Updated: ", currentVersion, " => ", version)
	runSendNeedReboot([]string{})
}

func findVar(varName string, input io.ReadCloser) (*semver.Version, error) {
	data, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("Failed to read input: %v", err)
	}

	buf := bytes.NewBuffer(data)

	for {
		s, err := buf.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("Unable to find %s: %v", varName, err)
		}

		p := strings.SplitN(s, "=", 2)
		if len(p) != 2 {
			continue
		}
		if strings.TrimSpace(p[0]) != varName {
			continue
		}

		v, err := semver.NewVersion(strings.TrimSpace(p[1]))
		return v, err
	}
}
