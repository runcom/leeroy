package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
)

const (
	VERSION = "v0.1.0"
)

var (
	serverUri    string
	jenkinsUri   string
	jenkinsUser  string
	jenkinsPass  string
	buildCommits string
	ghtoken      string
	certFile     string
	keyFile      string
	configFile   string
	debug        bool
	version      bool

	builds []Build

	Context string = "janky"
)

type Build struct {
	Repo string `json:"github_repo"`
	Job  string `json:"jenkins_job_name"`
}

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&serverUri, "server-uri", "", "server uri for this instance of leeroy")
	flag.StringVar(&jenkinsUri, "jenkins-uri", "", "jenkins uri")
	flag.StringVar(&jenkinsUser, "jenkins-user", "", "jenkins user")
	flag.StringVar(&jenkinsPass, "jenkins-pass", "", "jenkins password")
	flag.StringVar(&buildCommits, "build-commits", "", "commits to build per PR [all, new, last]")
	flag.StringVar(&ghtoken, "gh-token", "", "github access token")
	flag.StringVar(&certFile, "cert", "", "path to ssl certificate")
	flag.StringVar(&keyFile, "key", "", "path to ssl key")
	flag.StringVar(&configFile, "config", "/etc/leeroy/config.json", "path to config file")
	flag.Parse()
}

func main() {
	// set log level
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if version {
		fmt.Println(VERSION)
		return
	}

	// read the config file
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Errorf("config file does not exist: %s", configFile)
		return
	}
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf("could not read config file: %v", err)
		return
	}
	if err := json.Unmarshal(config, &builds); err != nil {
		log.Errorf("error parsing config file as json: %v", err)
		return
	}

	// create mux server
	mux := http.NewServeMux()

	// jenkins notification endpoint
	mux.Handle("/notifications/jenkins", jenkinsHandler)

	// github webhooks endpoint
	mux.Handle("/notifications/github", githubHandler)

	// retry build endpoint
	mux.Handle("/build/retry", retryBuildHandler)

	// custom build endpoint
	mux.Handle("/build/custom", customBuildHandler)

	// set up the server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting server on port %q with baseuri %q", port, baseuri)
	if certFile != "" && keyFile != "" {
		log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
	} else {
		log.Fatal(server.ListenAndServe())
	}
}
