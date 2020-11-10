package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/endocrimes/keylight-go"
)

type lightListFlags []string

func (f *lightListFlags) String() string {
	return strings.Join(*f, ",")
}

func (f *lightListFlags) Set(value string) error {
	*f = append(*f, strings.TrimSpace(value))
	return nil
}

type lightDiscoverer struct {
	RequiredLights []string
	AllLights      bool
	Discovery      keylight.Discovery

	discoveredLights map[string]*keylight.Device
}

func (l *lightDiscoverer) runCollector(ctx context.Context) error {
	if l.discoveredLights == nil {
		l.discoveredLights = make(map[string]*keylight.Device)
	}

	resultsCh := l.Discovery.ResultsCh()
	for {
		select {
		case <-ctx.Done():
			return nil
		case light := <-resultsCh:
			if light == nil {
				return nil
			}

			if l.AllLights {
				l.discoveredLights[light.Name] = light
				continue
			}

			for _, req := range l.RequiredLights {
				// TODO: Should check if the requirement is a full name or short name
				//       and compare differently based on the two (full match vs suffix)
				if strings.HasSuffix(light.Name, req) {
					l.discoveredLights[light.Name] = light
				}
			}

			if len(l.discoveredLights) == len(l.RequiredLights) {
				return nil
			}
		}
	}
}

func validateAllRequiredLights(lights []*keylight.Device, requirements []string) error {
	if len(requirements) == 0 {
		return nil
	}

REQUIREMENTS:
	for _, req := range requirements {
		for _, light := range lights {
			if strings.HasSuffix(light.Name, req) {
				continue REQUIREMENTS
			}
		}
		return fmt.Errorf("no light found for requirement '%s'", req)
	}

	return nil
}

func (l *lightDiscoverer) DiscoveredLights() []*keylight.Device {
	var result []*keylight.Device
	for _, light := range l.discoveredLights {
		result = append(result, light)
	}
	return result
}

func (l *lightDiscoverer) Run(ctx context.Context) ([]*keylight.Device, error) {
	childCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	doneCh := make(chan error)

	go func() {
		err := l.Discovery.Run(childCtx)
		if err != nil {
			cancelFn()
		}
		doneCh <- err
	}()

	err := l.runCollector(childCtx)
	if err != nil {
		return nil, err
	}

	discoveryErr := <-doneCh
	if discoveryErr != nil {
		return nil, discoveryErr
	}

	lights := l.DiscoveredLights()
	err = validateAllRequiredLights(lights, l.RequiredLights)
	if err != nil {
		return nil, err
	}
	return lights, nil
}
