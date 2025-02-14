package cmd

import (
	"fmt"
	"github.com/4thel00z/pcopy/client"
	"github.com/4thel00z/pcopy/config"
	"github.com/4thel00z/pcopy/server"
	"github.com/urfave/cli/v2"
)

var cmdLink = &cli.Command{
	Name:      "link",
	Aliases:   []string{"n"},
	Usage:     "Generate direct download link to clipboard content",
	UsageText: "pcopy link [OPTIONS..] [[CLIPBOARD]:[ID]]",
	Action:    execLink,
	Category:  categoryClient,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "load config file from `FILE`"},
		&cli.StringFlag{Name: "username", Aliases: []string{"u"}, DefaultText: "", Usage: "set basic auth user name, in case remote clipboard is behind a proxy"},
		&cli.StringFlag{Name: "password", Aliases: []string{"pw"}, DefaultText: "", Usage: "set basic auth password, in case remote clipboard is behind a proxy"},
	},
	Description: `Retrieves the link for the given clipboard file that can be used to share
with others.

Examples:
  pcopy link                  # Generates link for the default clipboard
  pcopy link work:            # Generates link for default file in clipboard 'work'`,
}

func execLink(c *cli.Context) error {
	conf, id, err := parseLinkArgs(c)
	if err != nil {
		return err
	}
	pclient, err := client.NewClient(conf, c.String("username"), c.String("password"))
	if err != nil {
		return err
	}
	info, err := pclient.FileInfo(id)
	if err != nil {
		return err
	}
	fmt.Fprint(c.App.ErrWriter, server.FileInfoInstructions(info))
	return nil
}

func parseLinkArgs(c *cli.Context) (*config.Config, string, error) {
	configFileOverride := c.String("config")

	// Parse clipboard and file
	clipboard, id := config.DefaultClipboard, config.DefaultID
	if c.NArg() > 0 {
		var err error
		clipboard, id, err = parseClipboardAndID(c.Args().First(), configFileOverride)
		if err != nil {
			return nil, "", err
		}
	}

	// Load config
	configFile, conf, err := parseAndLoadConfig(configFileOverride, clipboard)
	if err != nil {
		return nil, "", cli.Exit("clipboard does not exist", 1)
	}

	// Load defaults
	if id == "" {
		id = conf.DefaultID
	}
	if conf.CertFile == "" {
		conf.CertFile = config.DefaultCertFile(configFile, true)
	}

	return conf, id, nil
}
