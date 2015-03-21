package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/jacobstr/confer"
)

func main() {
	config := confer.NewConfig()
	config.ReadPaths("config.yaml")
	config.SetDefault("app.host", "localhost")
	spew.Dump(config.GetString("app.host"))
	spew.Dump(config.GetString("app.port"))

	glog.Infof("Got arguments: %v", os.Args)

	serverURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", config.GetString("app.host"), config.GetString("app.port")),
		Path:   "/api/1/postprocess",
	}

	flag.Set("logtostderr", "true")
	if len(os.Args) < 2 {
		fmt.Println("Too few arguments given")
		os.Exit(1)
	}
	formVals := url.Values{
		"path":   {os.Args[1]},
		"source": {""},
	}

	req, err := http.NewRequest("POST", serverURL.String(),
		strings.NewReader(formVals.Encode()))
	if err != nil {
		glog.Fatalf("Error creating request: %s", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/event-stream")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		glog.Fatalf("Error making request: %s", err)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			glog.Fatalf("Error reading from server: %s", err)
		}
		fmt.Print(string(line))
	}
}
