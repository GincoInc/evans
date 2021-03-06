package repl

import (
	"strings"

	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

type showCommand struct {
	env *env.Env
}

func (c *showCommand) Synopsis() string {
	return "show package, service or RPC names"
}

func (c *showCommand) Help() string {
	return "usage: show <package | service | message | rpc>"
}

func (c *showCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "target type (package, service, message)")
	}
	return nil
}

func (c *showCommand) Run(args []string) (string, error) {
	target := args[0]

	switch strings.ToLower(target) {
	case "p", "package", "packages":
		return c.env.GetPackages().String(), nil

	case "s", "svc", "service", "services":
		svc, err := c.env.GetServices()
		if err != nil {
			return "", err
		}
		return svc.String(), nil

	case "m", "msg", "message", "messages":
		msg, err := c.env.GetMessages()
		if err != nil {
			return "", err
		}
		return msg.String(), nil

	case "a", "r", "rpc", "api":
		rpcs, err := c.env.GetRPCs()
		if err != nil {
			return "", err
		}
		return rpcs.String(), nil

	default:
		return "", errors.Wrap(ErrUnknownTarget, target)
	}
}
