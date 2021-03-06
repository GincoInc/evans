package repl

import (
	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

type serviceCommand struct {
	env *env.Env
}

func (c *serviceCommand) Synopsis() string {
	return "set the service as the current selected service"
}

func (c *serviceCommand) Help() string {
	return "usage: service <service name>"
}

func (c *serviceCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "service name")
	}
	return nil
}

func (c *serviceCommand) Run(args []string) (string, error) {
	if err := c.env.UseService(args[0]); err != nil {
		return "", err
	}
	return "", nil
}
