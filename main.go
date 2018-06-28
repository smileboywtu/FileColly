package main

import (
	"os"
	"os/signal"
	"fmt"
	"syscall"

	"github.com/urfave/cli"
	"github.com/yudai/gotty/pkg/homedir"
	"github.com/smileboywtu/FileColly/common"
	collector "github.com/smileboywtu/FileColly/colly"
	"time"
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
			EnvVar: "COLLY_CONFIG",
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

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGKILL, syscall.SIGTERM)

		go func() {
			<-sigs
			colly.ShutDown()
		}()

		colly.FileWalkerInst.OnFilter(collector.FileWalkerGenericFilter)
		colly.OnFilter(collector.CollectorGenericFilter)

		for {
			colly.Start()
			time.Sleep(time.Duration(1 * time.Second))
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
