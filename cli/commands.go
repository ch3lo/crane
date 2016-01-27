package cli

import "github.com/codegangsta/cli"

var commands = []cli.Command{
	{
		Name:    "deploy",
		Aliases: []string{"d"},
		Usage:   "deploy or scale a service",
		Flags:   deployFlags(),
		Before:  deployBefore,
		Action:  deployCmd,
	},
	{
		Name:    "find",
		Aliases: []string{"f"},
		Usage:   "find services",
		Flags:   findFlags(),
		Before:  findBefore,
		Action:  findCmd,
	},
	{
		Name:    "delete",
		Aliases: []string{"x"},
		Usage:   "delete service",
		Flags:   deleteFlags(),
		Before:  deleteBefore,
		Action:  deleteCmd,
	},
}
