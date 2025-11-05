package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jirihnidek/rhsm2"
	"github.com/urfave/cli/v2"
)

var (
	appName    = "sub-man"
	appVersion = "0.0.1"
)

var rhsmClient *rhsm2.RHSMClient

func pingAction(ctx *cli.Context) error {
	_, err := rhsmClient.GetServerEndpoints(nil)
	if err != nil {
		return err
	}
	fmt.Printf("Candlepin server https://%s:%s is running\n",
		rhsmClient.RHSMConf.Server.Hostname,
		rhsmClient.RHSMConf.Server.Port,
	)
	return nil
}

// versionAction tries to print version and version of server
func versionAction(ctx *cli.Context) error {
	clientVer, err := clientVersion()
	if err != nil {
		return err
	}
	fmt.Printf("sub-man: %s\n", clientVer)
	serverVer, serverRulesVer, err := serverVersion()
	if err != nil {
		return err
	}
	fmt.Printf("subscription management server: %s\n", *serverVer)
	fmt.Printf("subscription management rules: %s\n", *serverRulesVer)
	return nil
}

// releaseAction tries to print available releases
func releaseAction(ctx *cli.Context) error {
	if ctx.Bool("show") {
		currentRelease, err := rhsmClient.GetReleaseFromServer(nil)
		if err != nil {
			return err
		}
		if currentRelease == "" {
			fmt.Printf("current release: not set\n")
		} else {
			fmt.Printf("current release: %s\n", currentRelease)
		}
		return nil
	}

	if release := ctx.String("set"); release != "" {
		err := rhsmClient.SetReleaseOnServer(nil, release)
		if err != nil {
			return err
		}
		fmt.Printf("release set to %s\n", release)
		return nil
	}

	if ctx.Bool("unset") {
		err := rhsmClient.SetReleaseOnServer(nil, "")
		if err != nil {
			return err
		}
		fmt.Println("release was unset")
		return nil
	}

	if ctx.Bool("list") {
		releases, err := rhsmClient.GetCdnReleases(nil)
		if err != nil {
			return err
		}

		for release := range releases {
			fmt.Printf("%s\n", release)
		}

		return nil
	}

	return nil
}

// identityAction tries to prettyPrint system identity
func identityAction(ctx *cli.Context) error {
	uuid, err := rhsmClient.GetConsumerUUID()
	if err != nil {
		return err
	}

	owner, err := rhsmClient.GetOwner()
	if err != nil {
		return nil
	}

	fmt.Printf("system identity: %v\n", *uuid)
	fmt.Printf("org ID: %v\n", owner)

	return nil
}

// configAction tries to prettyPrint configuration
func configAction(ctx *cli.Context) error {
	err := prettyPrint(rhsmClient.RHSMConf)

	if err != nil {
		return nil
	}

	return nil
}

// beforeRegisterAction checks possible combinations of CLI options
func beforeRegisterAction(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")
	org := ctx.String("organization")
	environments := ctx.StringSlice("environments")
	activationKeys := ctx.StringSlice("activation-key")

	// Username has to be provided with password
	if len(username) > 0 {
		if len(password) == 0 {
			return fmt.Errorf("--username USERNAME has to be provided with --password PASSWORD")
		}
	}
	// Check if username/password and any activation key was provided, because
	// it is not possible to use both at the same time
	if len(username) > 0 && len(activationKeys) > 0 {
		return fmt.Errorf("cannot use both username/password and activation key(s) at the same time")
	}

	if len(activationKeys) > 0 {
		if len(org) == 0 {
			return fmt.Errorf("organization ID must be provided when using activation key(s)")
		}
	}

	// It is also not possible to use environment with activation key(s)
	if len(activationKeys) > 0 && len(environments) > 0 {
		return fmt.Errorf("cannot use both activation key(s) with environment(s) at the same time")
	}

	return nil
}

// registerAction tries to register system
func registerAction(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")
	org := ctx.String("organization")
	environments := ctx.StringSlice("environments")
	activationKeys := ctx.StringSlice("activation-key")

	var options = make(map[string]string)
	if len(activationKeys) > 0 {
		_, err := rhsmClient.RegisterOrgActivationKeys(&org, activationKeys, nil, nil)
		return err
	} else {
		if org != "" {
			options["org"] = org
		}
		if len(environments) > 0 {
			options["environments"] = strings.Join(environments, ",")
		}
		_, err := rhsmClient.RegisterUsernamePassword(&username, &password, &options, nil)
		return err
	}
}

// organizationAction tries to get list of organizations for given user
func organizationsAction(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")

	orgs, err := rhsmClient.GetOrgs(username, password, nil)
	if err != nil {
		return err
	}

	for _, org := range orgs {
		fmt.Printf("name: %s\n", org.DisplayName)
		fmt.Printf("key: %s\n\n", org.Key)
	}

	return nil
}

// environmentsAction tries to get list of environment for given user and organization
func environmentsAction(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")
	org := ctx.String("organization")

	environments, err := rhsmClient.GetEnvironments(username, password, org, nil)
	if err != nil {
		return err
	}

	for _, environment := range environments {
		fmt.Printf("name: %s\n", environment.Name)
		fmt.Printf("description: %s\n\n", environment.Description)
	}

	return nil
}

// unregisterAction tries to unregister the system from candlepin server
func unregisterAction(ctx *cli.Context) error {
	return rhsmClient.Unregister(nil)
}

// beforeAction is triggered before other actions are triggered
func beforeAction(ctx *cli.Context) error {
	var err error
	confFilePath := ctx.String("config")

	rhsm2.SetUserAgentCmd(appName)

	rhsmClient, err = rhsm2.GetRHSMClient(&confFilePath)

	if err != nil {
		return fmt.Errorf("failed to create RHSM client: %v", err)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:    appName,
		Version: appVersion,
		Usage:   "Minimalistic CLI client for RHSM",
	}

	// List of CLI options of application
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:      "config",
			Hidden:    true,
			Value:     rhsm2.DefaultRHSMConfFilePath,
			TakesFile: true,
			Usage:     "Read config values from `FILE`",
		},
	}

	// List of sub-command of application
	app.Commands = []*cli.Command{
		{
			Name: "register",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "username",
					Usage:   "register with `USERNAME`",
					Aliases: []string{"u"},
				},
				&cli.StringFlag{
					Name:    "password",
					Usage:   "register with `PASSWORD`",
					Aliases: []string{"p"},
				},
				&cli.StringFlag{
					Name:    "organization",
					Usage:   "register with `ID`",
					Aliases: []string{"o", "org"},
				},
				&cli.StringSliceFlag{
					Name:    "environments",
					Usage:   "register with `ENVIRONMENT(s)`",
					Aliases: []string{"e"},
				},
				&cli.StringSliceFlag{
					Name:    "activation-key",
					Usage:   "register with `KEY(s)`",
					Aliases: []string{"k"},
				},
			},
			Usage:       "Register system",
			UsageText:   fmt.Sprintf("%v register [command options]", app.Name),
			Description: "The register command registers the system to Red Hat Subscription Management",
			Before:      beforeRegisterAction,
			Action:      registerAction,
		},
		{
			Name: "organizations",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "username",
					Usage:   "get organization for given `USERNAME`",
					Aliases: []string{"u"},
				},
				&cli.StringFlag{
					Name:    "password",
					Usage:   "`PASSWORD` of given `USERNAME`",
					Aliases: []string{"p"},
				},
			},
			Usage:       "Print organizations",
			UsageText:   fmt.Sprintf("%v orgs", app.Name),
			Description: "Get list of organizations for given user",
			Aliases:     []string{"orgs"},
			Action:      organizationsAction,
		},
		{
			Name: "environments",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "username",
					Usage:   "register with `USERNAME`",
					Aliases: []string{"u"},
				},
				&cli.StringFlag{
					Name:    "password",
					Usage:   "register with `PASSWORD`",
					Aliases: []string{"p"},
				},
				&cli.StringFlag{
					Name:    "organization",
					Usage:   "register with `ID`",
					Aliases: []string{"o", "org"},
				},
			},
			Usage:       "Print environments",
			UsageText:   fmt.Sprintf("%v environments", app.Name),
			Description: "Get list of environments for given user and organization",
			Action:      environmentsAction,
		},
		{
			Name:        "unregister",
			Usage:       "Unregister system",
			UsageText:   fmt.Sprintf("%v unregister", app.Name),
			Description: "Unregister the system",
			Action:      unregisterAction,
		},
		{
			Name:        "release",
			Usage:       "Manage release",
			UsageText:   fmt.Sprintf("%v release", app.Name),
			Description: "Manage release of system",
			Action:      releaseAction,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "list",
					Usage:   "list available releases",
					Aliases: []string{"l"},
					Value:   true,
				},
				&cli.BoolFlag{
					Name:    "show",
					Usage:   "show current release",
					Aliases: []string{"s"},
					Value:   false,
				},
				&cli.StringFlag{
					Name:    "set",
					Usage:   "set release",
					Aliases: []string{"S"},
				},
				&cli.BoolFlag{
					Name:    "unset",
					Usage:   "unset release",
					Aliases: []string{"U"},
				},
			},
		},
		{
			Name:        "identity",
			Usage:       "Print identity",
			UsageText:   fmt.Sprintf("%v identity", app.Name),
			Description: "Print identity of system",
			Action:      identityAction,
		},
		{
			Name:        "config",
			Usage:       "Print configuration",
			UsageText:   fmt.Sprintf("%v config", app.Name),
			Description: fmt.Sprintf("Print configuration of %v", app.Name),
			Action:      configAction,
		},
		{
			Name:        "version",
			Usage:       "Print version",
			UsageText:   fmt.Sprintf("%s version", app.Name),
			Description: fmt.Sprintf("Print version of %s and server", app.Name),
			Action:      versionAction,
		},
		{
			Name:        "ping",
			Usage:       "Ping server",
			UsageText:   fmt.Sprintf("%s ping", app.Name),
			Description: "Try to ping candlepin server",
			Action:      pingAction,
		},
	}

	app.Before = beforeAction

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
