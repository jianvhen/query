package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jianvhen/query/g"
	"github.com/jianvhen/query/graph"
	"github.com/jianvhen/query/http"
	"github.com/jianvhen/query/proc"
)

func main() {
	cfg := flag.String("c", "cfg.json", "specify config file")
	version := flag.Bool("v", false, "show version")
	versionGit := flag.Bool("vg", false, "show version and git commit log")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}
	if *versionGit {
		fmt.Println(g.VERSION, g.COMMIT)
		os.Exit(0)
	}

	// config
	g.ParseConfig(*cfg)
	// proc
	proc.Start()

	// graph
	graph.Start()

	// http
	http.Start()

	select {}
}
