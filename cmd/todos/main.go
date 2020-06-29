package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/bbengfort/todos"
	"github.com/bbengfort/todos/client"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v3"

	// Load database dialects for use with gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
		{
			Name:     "createsuperuser",
			Usage:    "create an admin user to register other users",
			Action:   createSuperUser,
			Category: "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "u, username",
					Usage: "specify username instead of prompting",
				},
				cli.StringFlag{
					Name:  "e, email",
					Usage: "specify email instead of prompting",
				},
				cli.BoolFlag{
					Name:  "g, generate",
					Usage: "generate password instead of prompting",
				},
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "database connection uri",
					EnvVar: "DATABASE_URL",
				},
			},
		},
		{
			Name:     "configure",
			Usage:    "configure the local client to connect to the todos api",
			Action:   configure,
			Category: "client",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "e, endpoint",
					Usage: "specify the endpoint directly without prompting for it",
				},
				cli.StringFlag{
					Name:  "u, username",
					Usage: "specify the username directly without prompting for it",
				},
				cli.BoolFlag{
					Name:  "p, password",
					Usage: "prompt to enter a password into the credentials file",
				},
				cli.BoolFlag{
					Name:  "d, dir",
					Usage: "print the directory containing the configuration and exit",
				},
			},
		},
		{
			Name:     "status",
			Usage:    "make a status request to the server to check if its up",
			Action:   status,
			Category: "client",
			Flags:    []cli.Flag{},
		},
		{
			Name:     "login",
			Usage:    "authenticate with the api server and cache keys",
			Action:   login,
			Category: "client",
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

// characters for generating random passwords
const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func createSuperUser(c *cli.Context) (err error) {
	var db *gorm.DB
	if dburl := c.String("db"); dburl != "" {
		if db, err = gorm.Open("postgres", dburl); err != nil {
			return cli.NewExitError(err, 1)
		}
	} else {
		return cli.NewExitError("specify $DATABASE_URL to create the admin user", 1)
	}

	username := c.String("username")
	email := c.String("email")
	password := ""

	if c.Bool("generate") {
		seed := rand.New(rand.NewSource(time.Now().UnixNano()))

		b := make([]byte, 8)
		for i := range b {
			b[i] = charset[seed.Intn(len(charset))]
		}
		password = string(b)
	}

	if username == "" {
		username = client.Prompt("username", "")
	}

	if email == "" {
		email = client.Prompt("email", "")
	}

	if password == "" {
		password, err = client.PromptPassword("password", true, false)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	user := &todos.User{
		Username: username,
		Email:    email,
		IsAdmin:  true,
	}

	// TODO: this should be a standalone call rather than modifying the struct
	if user.Password, err = user.SetPassword(password); err != nil {
		return cli.NewExitError(err, 1)
	}

	// Save the user to the database
	if err = db.Create(user).Error; err != nil {
		return cli.NewExitError(err, 1)
	}

	// If the password was generated, inform the user
	if c.Bool("generate") {
		fmt.Printf("%s:%s\n", username, password)
	}
	return nil
}

//===========================================================================
// Client Commands
//===========================================================================

func configure(c *cli.Context) (err error) {
	if c.Bool("dir") {
		var cdir string
		if cdir, err = client.Configuration(); err != nil {
			return cli.NewExitError(err, 1)
		}
		fmt.Println(cdir)
		return
	}

	// Attempt to load the previous credentials to provide defaults if they exist.
	creds := client.Credentials{}
	creds.Load()

	if ep := c.String("endpoint"); ep != "" {
		creds.Endpoint = ep
	} else {
		creds.Endpoint = client.Prompt("endpoint", creds.Endpoint)
	}

	if un := c.String("username"); un != "" {
		creds.Username = un
	} else {
		creds.Username = client.Prompt("username", creds.Username)
	}

	if c.Bool("password") {
		if creds.Password, err = client.PromptPassword("password", true, true); err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	// Write the configuration to disk
	if err = creds.Dump(); err != nil {
		return cli.NewExitError(err, 1)
	}

	return nil
}

func status(c *cli.Context) (err error) {
	var api *client.Client
	if api, err = client.New(); err != nil {
		return cli.NewExitError(err, 1)
	}

	var data map[string]interface{}
	if data, err = api.Status(); err != nil {
		return cli.NewExitError(err, 1)
	}

	var out []byte
	if out, err = yaml.Marshal(data); err != nil {
		return cli.NewExitError(err, 1)
	}

	fmt.Print(string(out))
	return nil
}

func login(c *cli.Context) (err error) {
	var api *client.Client
	if api, err = client.New(); err != nil {
		return cli.NewExitError(err, 1)
	}

	if err = api.Login(); err != nil {
		return cli.NewExitError(err, 1)
	}

	return nil
}
