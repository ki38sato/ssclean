package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

var version string

func main() {

	var profile, region, kind string
	var keep int
	var filters []string
	var dryrun, delss bool

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "profile",
			Usage:       "aws iam profile name",
			Destination: &profile,
		},
		cli.StringFlag{
			Name:        "region",
			Usage:       "aws region name",
			Destination: &region,
		},
		cli.StringFlag{
			Name:        "kind",
			Usage:       "choice ami or snapshot",
			Destination: &kind,
		},
		cli.BoolFlag{
			Name:        "dryrun",
			Usage:       "dryrun",
			Destination: &dryrun,
		},
		cli.IntFlag{
			Name:        "keep",
			Usage:       "specify keep age",
			Destination: &keep,
		},
		cli.StringSliceFlag{
			Name:  "filters",
			Usage: "filters",
		},
		cli.BoolFlag{
			Name:        "del-snapshot",
			Usage:       "delete snapshot with ami",
			Destination: &delss,
		},
	}

	app.Name = "ssclean"
	app.Usage = ""
	app.Version = version
	app.Action = func(c *cli.Context) error {

		filters = c.StringSlice("filters")

		svc, err := createEC2Session(profile, region)
		if err != nil {
			return err
		}

		switch kind {
		case "ami":
			err := rmImages(svc, keep, filters, dryrun, delss)
			if err != nil {
				return err
			}
		case "snapshot":
			err := rmSnapshots(svc, keep, filters, dryrun)
			if err != nil {
				return err
			}
		default:
			fmt.Printf("invalid kind:%s. kind is 'ami' or 'snapshot'\n", kind)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Println(err.Error())
	}
}
