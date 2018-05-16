package main

import (
	"fmt"
	"os"
	"time"
	"github.com/urfave/cli"
	"github.com/yudai/gotty/pkg/homedir"
	"github.com/smileboywtu/FileCollector/common"
	collector "github.com/smileboywtu/FileCollector/colly"
)

var email string
var author string
var version string

func main() {

	app := cli.NewApp()
	app.Name = "File Collector"
	app.Version = version
	app.Author = author
	app.Email = email
	app.Usage = "collect file from directory and send to redis"
	app.HideHelp = true

	cli.AppHelpTemplate = helpTemplate

	appOptions := &collector.AppConfigOption{}
	if err := common.ApplyDefaultValues(appOptions); err != nil {
		exit(err, 1)
	}

	cliFlags, flagMappings, err := common.GenerateFlags(appOptions)
	if err != nil {
		exit(err, 3)
	}

	app.Flags = append(
		cliFlags,
		cli.StringFlag{
			Name:   "config",
			Value:  "config.yaml",
			Usage:  "Config file path",
			EnvVar: "CROC_CONFIG",
		},
	)

	app.Action = func(c *cli.Context) {

		configFile := c.String("config")
		_, err := os.Stat(homedir.Expand(configFile))
		if configFile != "config.yaml" || !os.IsNotExist(err) {
			if err := common.ApplyConfigFileYaml(configFile, appOptions); err != nil {
				exit(err, 2)
			}
		}

		common.ApplyFlags(cliFlags, flagMappings, c, appOptions)

		colly, errs := collector.NewCollector(appOptions)
		if errs != nil {
			fmt.Fprintf(os.Stderr, "start error: %s", errs.Error())
			os.Exit(-1)
		}

		colly.FileWalkerInst.OnFilter(collector.FileWalkerGenericFilter)
		colly.OnFilter(collector.CollectorGenericFilter)

		for {
			colly.SendFlow()
			time.Sleep(time.Duration(3 * time.Second))
		}
	}

	app.Run(os.Args)
}

func exit(err error, code int) {
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}
