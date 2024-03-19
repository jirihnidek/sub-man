package main

import (
	"fmt"
	"github.com/jirihnidek/rhsm2"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var (
	appName    = "sub-man"
	appVersion = "0.0.1"
)

var rhsmClient *rhsm2.RHSMClient

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

// registerAction tries to register system
func registerAction(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")
	org := ctx.String("organization")
	activationKeys := ctx.StringSlice("activation-key")

	// Check if username/password and any activation key was provided, because
	// it is not possible to use both at the same time
	if len(username) > 0 && len(activationKeys) > 0 {
		return fmt.Errorf("cannot use both username/password and activation key(s) at the same time")
	}

	if len(activationKeys) > 0 {
		if len(org) == 0 {
			return fmt.Errorf("organization ID must be provided when using activation key(s)")
		}
		_, err := rhsmClient.RegisterOrgActivationKeys(&org, activationKeys, nil)
		return err
	} else {
		_, err := rhsmClient.RegisterUsernamePasswordOrg(&username, &password, &org, nil)
		return err
	}
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

	rhsmClient, err = rhsm2.CreateRHSMClient(&confFilePath)

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
					Aliases: []string{"o"},
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
			Action:      registerAction,
		},
		{
			Name:        "unregister",
			Usage:       "Unregister system",
			UsageText:   fmt.Sprintf("%v unregister", app.Name),
			Description: "Unregister the system",
			Action:      unregisterAction,
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
	}

	app.Before = beforeAction

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
