package command

import (
	"context"
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

	discoveredLights []*keylight.KeyLight
}

func (l *lightDiscoverer) runCollector(ctx context.Context) error {
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
				l.discoveredLights = append(l.discoveredLights, light)
				continue
			}

			for _, req := range l.RequiredLights {
				// TODO: Should check if the requirement is a full name or short name
				//       and compare differently based on the two (full match vs suffix)
				if strings.HasSuffix(light.Name, req) {
					l.discoveredLights = append(l.discoveredLights, light)
				}
			}

			// TODO: Potential bug here if a light is discovered multiple times during
			//       this phase (e.g flaky network or power cycling). Should probably
			//       store discovered lights as map[name]light to avoid this.
			if len(l.discoveredLights) == len(l.RequiredLights) {
				return nil
			}
		}
	}
}

func (l *lightDiscoverer) Run(ctx context.Context) ([]*keylight.KeyLight, error) {
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

	return l.discoveredLights, nil
}
