package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/endocrimes/keylight-go"
	"github.com/mitchellh/cli"
)

type SwitchCommand struct {
	Meta
}

func (c *SwitchCommand) Help() string {
	helpText := `
Usage: keylightctl switch <on|off> [options]

 Switch keylights on and off.

General Options:

  ` + generalOptionsUsage() + `

Switch Specific Options:

  -timeout <duration>
    Sets the maximum time to listen for accessories (default: 5s)

  -all
    Modify all keylights that are discovered within the timeout window

  -light <light-id>
    Modify the provided light ID. Can either be a full key light name, e.g:
    Elgato\ Key\ Light\ 111A, or a short ID, e.g: 111A. -light can be provided
    multiple times and all provided lights will be modified.
`
	return strings.TrimSpace(helpText)
}

func (f *SwitchCommand) Synopsis() string {
	return "Switch keylights on and off"
}

func (f *SwitchCommand) Name() string { return "switch" }

func (c *SwitchCommand) Run(args []string) int {
	c.UI = &cli.PrefixedUi{
		OutputPrefix: "  ",
		InfoPrefix:   "  ",
		ErrorPrefix:  "==> ",
		Ui:           c.UI,
	}

	var timeout time.Duration
	var lights lightListFlags
	var allLights bool

	flags := c.Meta.FlagSet(c.Name(), FlagSetClient)
	flags.Usage = func() { c.UI.Output(c.Help()) }
	flags.DurationVar(&timeout, "timeout", 5*time.Second, "")
	flags.Var(&lights, "light", "")
	flags.BoolVar(&allLights, "all", false, "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if l := len(args); l != 1 {
		c.UI.Error("This command requires (1) argument")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	if args[0] != "on" && args[0] != "off" {
		c.UI.Error("Argument must be 'on' or off'")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	desiredPowerState := 0
	if args[0] == "on" {
		desiredPowerState = 1
	}

	if allLights && len(lights) != 0 {
		c.UI.Error("Cannot specify --all and --light")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	discovery, err := keylight.NewDiscovery()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to setup mDNS discovery, err: %v", err))
		return 1
	}

	discoverer := lightDiscoverer{
		Discovery:      discovery,
		AllLights:      allLights,
		RequiredLights: lights,
	}

	discoveryCtx, cancelFn := context.WithTimeout(context.Background(), timeout)
	defer cancelFn()

	found, err := discoverer.Run(discoveryCtx)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to discover lights, err: %v", err))
		return 1
	}

	if len(found) == 0 {
		c.UI.Error("Found no matching lights during discovery")
		return 1
	}

	updateCtx, updateCancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer updateCancelFn()

	for _, light := range found {
		opts, err := light.FetchLightOptions(updateCtx)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch light options (%s), err: %v", light.Name, err))
			return 1
		}

		newOpts := opts.Copy()
		for _, l := range newOpts.Lights {
			l.On = desiredPowerState
		}

		_, err = light.UpdateLightOptions(updateCtx, newOpts)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to update light (%s), err: %v", light.Name, err))
			return 1
		}
	}

	return 0
}
