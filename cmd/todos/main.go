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
			Flags:    []cli.Flag{},
		},
	}

	// Run the CLI program
	app.Run(os.Args)
}

//===========================================================================
// Server Commands
//===========================================================================

func serve(c *cli.Context) (err error) {
	if err = todos.Serve(); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}
