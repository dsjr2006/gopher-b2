package main

import (
	"os"

	"github.com/dsjr2006/gopherb2"
	"github.com/urfave/cli"
)

func main() {
	//var logTo string

	app := cli.NewApp()
	app.Name = "gopherb2"
	app.Version = "0.1.0"
	app.Description = "Application for managing Backblaze B2"
	app.Usage = "[global options] command [command options] [arguments...]"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log",
			Usage: "gopherb2 -log `gopher.log`",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "bucket",
			Aliases:     []string{"buckets"},
			Usage:       "[global] bucket [command] [arguments...]",
			Description: "Manages B2 Buckets",
			Subcommands: []cli.Command{
				{
					Name:        "create",
					Aliases:     []string{"new"},
					Usage:       "[global] bucket create [name of new bucket]",
					Description: "Creates New Backblaze B2 Bucket",
					Action: func(c *cli.Context) error {
						gopherb2.B2CreateBucket(c.Args().Get(0), false)
						return nil
					},
				},
			},
		},
		{
			Name:        "upload",
			Aliases:     []string{"put"},
			Usage:       "[global] upload [bucket id] [path or file]",
			Description: "Upload File to BackBlaze B2",
			Action: func(c *cli.Context) error {
				gopherb2.B2UploadFile(c.Args().Get(0), c.Args().Get(1))
				return nil
			},
		},
	}

	app.Run(os.Args)
}
