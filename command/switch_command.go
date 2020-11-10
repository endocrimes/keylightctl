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
    Control the provided light ID. Can either be a full key light name, e.g:
    Elgato\ Key\ Light\ 111A, or a short ID, e.g: 111A. -light can be provided
    multiple times and all provided lights will be modified.

	-light-addr <addr>:<port>
    Control the light at the provided address. Can be provided multiple times.
    Useful when wanting to avoid the slow discovery time over mDNS in automation.

  -brightness <brightness>
    When switching the light, also set the brightness to the given percentage.

  -temperature <temperature>
    When switching the light, also set the temperature to the given value.
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
	var namedLights lightListFlags
	var addressedLights lightListFlags
	var allLights bool
	var brightness, temperature int

	flags := c.Meta.FlagSet(c.Name(), FlagSetClient)
	flags.Usage = func() { c.UI.Output(c.Help()) }
	flags.DurationVar(&timeout, "timeout", 5*time.Second, "")
	flags.Var(&namedLights, "light", "")
	flags.Var(&addressedLights, "light-addr", "")
	flags.BoolVar(&allLights, "all", false, "")
	flags.IntVar(&brightness, "brightness", -1, "")
	flags.IntVar(&temperature, "temperature", -1, "")

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

	if allLights && len(namedLights) != 0 {
		c.UI.Error("Cannot specify --all and --light")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	if !allLights && len(namedLights) == 0 {
		c.UI.Error("Must specify one of --all and --light")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	discoveryCtx, discoveryCancelFn := context.WithTimeout(context.Background(), timeout)
	found, err := discoverLights(discoveryCtx, allLights, namedLights)
	discoveryCancelFn()
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if len(addressedLights) == 0 && len(found) == 0 {
		c.UI.Error("Found no matching lights during discovery")
		return 1
	}

	updateCtx, updateCancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer updateCancelFn()

	err = updateLights(updateCtx, found, desiredPowerState, temperature, brightness)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to update lights, err: %v", err))
	}

	return 0
}

func discoverLights(ctx context.Context, findAllLights bool, namedLights lightListFlags) ([]*keylight.Device, error) {
	discovery, err := keylight.NewDiscovery()
	if err != nil {
		return nil, fmt.Errorf("failed to setup mDNS discovery, err: %v", err)
	}

	discoverer := lightDiscoverer{
		Discovery:      discovery,
		AllLights:      findAllLights,
		RequiredLights: namedLights,
	}

	found, err := discoverer.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover lights, err: %v", err)
	}

	return found, nil
}

func updateLights(ctx context.Context, devices []*keylight.Device, desiredPowerState, temperature, brightness int) error {
	for _, dev := range devices {
		grp, err := dev.FetchLightGroup(ctx)
		if err != nil {
			return err
		}

		newGroup := grp.Copy()
		for _, light := range newGroup.Lights {
			light.On = desiredPowerState
			if temperature >= 0 {
				light.Temperature = temperature
			}
			if brightness >= 0 {
				light.Brightness = brightness
			}
		}

		_, err = dev.UpdateLightGroup(ctx, newGroup)
		if err != nil {
			return err
		}
	}

	return nil
}
