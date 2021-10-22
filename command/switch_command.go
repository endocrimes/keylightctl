package command

import (
	"context"
	"fmt"
	"strconv"
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
Usage: keylightctl switch [options] <on|off|toggle> 

 Switch keylights on and off.

General Options:

  ` + generalOptionsUsage() + `

Switch Specific Options:

  -timeout <duration>
    Sets the maximum time to listen for accessories (default: 5s)

  -all
    Modify all keylights that are discovered within the timeout window

  -light <light-id-or-addr>
    Modify the provided light. Can either be a full key light name, e.g:
    Elgato\ Key\ Light\ 111A, or a short ID, e.g: 111A. -light can be provided
    multiple times and all provided lights will be modified. If only addresses
    are provided then we will skip going through discovery.

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
	var requestedLights lightListFlags
	var allLights bool
	var brightness, temperature int

	flags := c.Meta.FlagSet(c.Name(), FlagSetClient)
	flags.Usage = func() { c.UI.Output(c.Help()) }
	flags.DurationVar(&timeout, "timeout", 5*time.Second, "")
	flags.Var(&requestedLights, "light", "")
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

	if args[0] != "on" && args[0] != "off" && args[0] != "toggle" {
		c.UI.Error("Argument must be 'on', off', or 'toggle'")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	toggle := false
	desiredPowerState := 0
	if args[0] == "on" {
		desiredPowerState = 1
	} else if args[0] == "toggle" {
		toggle = true
	}

	if allLights && len(requestedLights) != 0 {
		c.UI.Error("Cannot specify --all and --light together")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	if !allLights && len(requestedLights) == 0 {
		c.UI.Error("One of --all and --light must be provided")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	discoveryCtx, cancelFn := context.WithTimeout(context.Background(), timeout)
	defer cancelFn()
	found, err := c.discoverLights(discoveryCtx, requestedLights, allLights)
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
			if toggle {
				if l.On == 0 {
					l.On = 1
				} else {
					l.On = 0
				}
			} else {
				l.On = desiredPowerState
			}
			if temperature >= 0 {
				l.Temperature = temperature
			}
			if brightness >= 0 {
				l.Brightness = brightness
			}
		}

		_, err = light.UpdateLightOptions(updateCtx, newOpts)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to update light (%s), err: %v", light.Name, err))
			return 1
		}
	}

	return 0
}

func (c *SwitchCommand) discoverLights(ctx context.Context, lightInfo lightListFlags, discoverAll bool) ([]*keylight.KeyLight, error) {
	var result []*keylight.KeyLight

	specifiedLights := selectLights(lightInfo, isDirectLightAddress)
	if len(specifiedLights) != 0 {
		for _, lightAddr := range specifiedLights {
			parts := strings.Split(lightAddr, ":")
			port, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse port from light (%s), err: %w", lightAddr, err)
			}
			light := &keylight.KeyLight{
				DNSAddr: parts[0],
				Port:    port,
			}

			result = append(result, light)
		}
	}

	lightsToDiscover := selectLights(lightInfo, invert(isDirectLightAddress))
	if len(lightsToDiscover) == 0 && !discoverAll {
		return result, nil
	}

	discovery, err := keylight.NewDiscovery()
	if err != nil {
		return nil, fmt.Errorf("failed to setup discoverer, err: %w", err)
	}

	discoverer := lightDiscoverer{
		Discovery:      discovery,
		AllLights:      discoverAll,
		RequiredLights: lightsToDiscover,
	}

	discoveredLights, err := discoverer.Run(ctx)
	if err != nil {
		return nil, err
	}

	result = append(result, discoveredLights...)
	return result, nil
}

func selectLights(lights lightListFlags, selectFunc func(string) bool) lightListFlags {
	var result lightListFlags
	for _, l := range lights {
		if selectFunc(l) {
			result = append(result, l)
		}
	}

	return result
}

// isDirectLightAddress is a hacky implementation to check if the provided string
// is a light identifier or an address we can use to reach a light - currently it
// only checks whether the string contains a `:` to seperate the IP or DNS Addr
// and the Port (as we currently require both, rather than providing a default
// port).
func isDirectLightAddress(light string) bool {
	return strings.Contains(light, ":")
}

// invert inverts the result of a string -> bool func for use with the
// `selectLights` function.
func invert(innerFunc func(string) bool) func(string) bool {
	return func(str string) bool {
		return !innerFunc(str)
	}
}
