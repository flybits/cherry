package main

import (
	"log"
	"net/http"

	"github.com/flybits/cherry/examples/2-echo-server/handler"
	"github.com/flybits/cherry/examples/2-echo-server/version"
)

func main() {
	log.Printf("version: %s  revision: %s  branch: %s  goVersion: %s  buildTool: %s  buildTime: %s\n",
		version.Version, version.Revision, version.Branch, version.GoVersion, version.BuildTool, version.BuildTime)

	http.HandleFunc("/echo", handler.EchoHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
