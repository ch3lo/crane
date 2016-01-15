package cli

import (
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
)

func findFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "search",
			Usage: "text which contains the service",
		},
	}
}

func findBefore(c *cli.Context) error {
	if c.String("search") == "" {
		return errors.New("El nombre de la imagen esta vacio")
	}
	return nil
}

func findCmd(c *cli.Context) {
	services := stackManager.FindServiceInformation(c.String("search"))
	for _, service := range services {
		fmt.Printf("Service %s running with %d instances\n", service.ID, len(service.Instances))
	}
}
