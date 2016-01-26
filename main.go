package main

import (
	"github.com/latam-airlines/crane/cli"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
)

func main() {
	cli.RunApp()
}
