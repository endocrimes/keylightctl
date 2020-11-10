package keylight

import (
	"context"

	"github.com/oleksandr/bonjour"
)

var _ Discovery = &bonjourDiscovery{}

type bonjourDiscovery struct {
	resolver  *bonjour.Resolver
	resultsCh chan *Device
}

func newBonjourDiscovery() (*bonjourDiscovery, error) {
	resolver, err := bonjour.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	return &bonjourDiscovery{
		resolver:  resolver,
		resultsCh: make(chan *Device, 5), // Buffer a few results to simplify client impls
	}, nil
}

func (d *bonjourDiscovery) Run(ctx context.Context) error {
	results := make(chan *bonjour.ServiceEntry)
	err := d.resolver.Browse("_elg._tcp", "", results)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			close(d.resultsCh)
			d.resolver.Exit <- true
			return nil
		case e := <-results:
			d.resultsCh <- &Device{
				Name:    e.Instance,
				DNSAddr: e.HostName,
				Port:    e.Port,
			}
		}
	}
}

func (d *bonjourDiscovery) ResultsCh() <-chan *Device {
	return d.resultsCh
}
