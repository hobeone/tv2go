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

	"github.com/golang/glog"
)

func main() {
	flag.Set("alsologtostderr", "true")
	if len(os.Args) < 2 {
		glog.Fatalf("Too few arguments given")
	}

	glog.Infof("Got arguments: %v", os.Args)

	formVals := url.Values{
		"path":   {os.Args[1]},
		"source": {""},
	}

	req, err := http.NewRequest("POST", "http://localhost:9001/api/1/postprocess", strings.NewReader(formVals.Encode()))
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
