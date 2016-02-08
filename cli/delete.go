package cli

import (
	"errors"

	"github.com/codegangsta/cli"
	"github.com/latam-airlines/crane/logger"
)

func deleteFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "service-id",
			Usage: "service id of the service",
		},
	}
}

func deleteBefore(c *cli.Context) error {
	if c.String("service-id") == "" {
		return errors.New("Flag \"service-id\" is empty")
	}
	return nil
}

func deleteCmd(c *cli.Context) {
	err := stackManager.DeleteService(c.String("service-id"))
	if err != nil {
		logger.Instance().Fatalln("Error deleting service", err)
	} else {
		logger.Instance().Infoln("Service deleted: ", c.String("service-id"))
	}
}
