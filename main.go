package main

import (
	_ "github.com/latam-airlines/crane/Godeps/_workspace/src/github.com/latam-airlines/mesos-framework-factory/marathon"
	"github.com/latam-airlines/crane/cli"
)

func main() {
	cli.RunApp()
}
