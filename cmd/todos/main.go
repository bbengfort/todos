package main

import (
	"os"

	"github.com/bbengfort/todos"
	"github.com/joho/godotenv"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	// Load the .env file if it exists
	godotenv.Load()

	// Instantiate the CLI application
	app := cli.NewApp()
	app.Name = "todos"
	app.Version = todos.Version
	app.Usage = "a simple todos server and CLI for personal task tracking"

	app.Commands = []cli.Command{
		{
			Name:     "serve",
			Usage:    "run a todos server",
			Action:   serve,
			Category: "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, addr",
					Usage: "specify an ip address to bind on",
				},
				cli.IntFlag{
					Name:  "p, port",
					Usage: "specify the port to bind on",
				},
				cli.BoolFlag{
					Name:  "s, tls",
					Usage: "use tls with the server (requires domain)",
				},
				cli.StringFlag{
					Name:  "d, domain",
					Usage: "specify the domain of the server for tls",
				},
			},
		},
	}

	// Run the CLI program
	app.Run(os.Args)
}

//===========================================================================
// Server Commands
//===========================================================================

func serve(c *cli.Context) (err error) {
	// The configuration is primarily loaded from the environment and defaults
	var conf todos.Settings
	if conf, err = todos.Config(); err != nil {
		return cli.NewExitError(err, 1)
	}

	// Update the configuration from the CLI flags
	if addr := c.String("addr"); addr != "" {
		conf.Bind = addr
	}

	if port := c.Int("port"); port > 0 {
		conf.Port = port
	}

	if useTLS := c.Bool("tls"); useTLS {
		conf.UseTLS = useTLS
	}

	if domain := c.String("domain"); domain != "" {
		conf.Domain = domain
	}

	// Create the API server
	var api *todos.API
	if api, err = todos.New(conf); err != nil {
		return cli.NewExitError(err, 1)
	}

	// Run the API server
	if err = api.Serve(); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}
