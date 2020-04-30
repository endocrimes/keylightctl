package command

import "github.com/mitchellh/cli"

func Commands(metaPtr *Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"discover": func() (cli.Command, error) {
			return &DiscoverCommand{
				Meta: *metaPtr,
			}, nil
		},
		"switch": func() (cli.Command, error) {
			return &SwitchCommand{
				Meta: *metaPtr,
			}, nil
		},
	}
}
