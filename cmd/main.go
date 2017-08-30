package main

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/dwin/gopherb2"
	"github.com/urfave/cli"
)

var (
	logDest string
	debug   bool
	logFile = "stderr"
)

func main() {
	switch logDest {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    100,
			MaxAge:     90,
			MaxBackups: 10,
		})
	}

	app := cli.NewApp()
	app.Name = "gopherb2"
	app.Version = "0.1.0"
	app.Description = "Application for managing and interacting with Backblaze B2"
	app.Usage = "[global options] command [command options] [arguments...]"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "log",
			Usage:       "gopherb2 -log `gopher.log`",
			Destination: &logDest,
		},
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "`-debug|-d` [command]",
			Destination: &debug,
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
						checkDebug()
						gopherb2.B2CreateBucket(c.Args().Get(0), false)
						return nil
					},
				},
				{
					Name:        "list",
					Usage:       "[global] bucket list",
					Description: "List all Buckets in Account",
					Action: func(c *cli.Context) error {
						buckets, err := gopherb2.GetBuckets()
						if err != nil {
							log.Fatal(err)
						}
						err = gopherb2.PrintBuckets(buckets)
						if err != nil {
							log.Fatal(err)
						}
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
				checkDebug()
				gopherb2.UploadFile(c.Args().Get(0), c.Args().Get(1))
				return nil
			},
		},
		{
			Name:        "file",
			Aliases:     []string{"files"},
			Usage:       "[global] file [command] [arguments..]",
			Description: "Manages B2 Files",
			Subcommands: []cli.Command{
				{
					Name:        "list",
					Usage:       "[global] file list [bucketId]",
					Description: "List all files in given Bucket",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: "all",
						},
					},
					Action: func(c *cli.Context) error {
						if !c.Bool("all") {
							fmt.Println("Non-Working: List All Files in all buckets")
						}

						gopherb2.B2ListFilenames(c.Args().Get(0), "")
						return nil
					},
				},
			},
		},
		{
			Name:        "version",
			Aliases:     []string{"v"},
			Usage:       "Display version",
			Description: "Display version",
			Action: func(c *cli.Context) error {
				fmt.Println("Version: " + app.Version)
				return nil
			},
		},
	}

	// B2ListBuckets()

	app.Run(os.Args)
}

func checkDebug() {
	if debug {
		fmt.Println("debug on")
		gopherb2.SetLogLevel("debug")
	}
	return
}
