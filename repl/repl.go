package repl

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/env"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

var (
	ErrUnknownCommand   = errors.New("unknown command")
	ErrArgumentRequired = errors.New("argument required")
	ErrUnknownTarget    = errors.New("unknown target")
)

func newUI() *UI {
	return &UI{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
}

type REPL struct {
	ui     *UI
	config *config.REPL
	env    *env.Env
	prompt *prompt.Prompt
	cmds   map[string]Commander
}

type UI struct {
	Reader            io.Reader
	Writer, ErrWriter io.Writer
	prompt            string
}

func NewBasicUI() *UI {
	return &UI{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
}

func NewREPL(config *config.REPL, env *env.Env, ui *UI) *REPL {
	cmds := map[string]Commander{
		"call":    &callCommand{env},
		"desc":    &descCommand{env},
		"package": &packageCommand{env},
		"service": &serviceCommand{env},
		"show":    &showCommand{env},
	}
	repl := &REPL{
		ui:     ui,
		config: config,
		env:    env,
		cmds:   cmds,
	}

	executor := &executor{repl: repl}
	completer := &completer{cmds: cmds, env: env}

	repl.prompt = prompt.New(
		executor.execute,
		completer.complete,
		prompt.OptionPrefix(repl.getPrompt()),

		prompt.OptionSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.White),
		prompt.OptionDescriptionTextColor(prompt.Black),

		prompt.OptionSelectedSuggestionBGColor(prompt.DarkBlue),
		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionSelectedDescriptionBGColor(prompt.Blue),
		prompt.OptionSelectedDescriptionTextColor(prompt.Black),
	)

	return repl
}

func (r *REPL) eval(l string) (string, error) {
	part := strings.Split(l, " ")

	if part[0] == "help" {
		r.showHelp(r.cmds)
		return "", nil
	}

	cmd, ok := r.cmds[part[0]]
	if !ok {
		return "", ErrUnknownCommand
	}

	var args []string
	if len(part) != 1 {
		if part[1] == "-h" || part[1] == "--help" {
			return cmd.Help(), nil
		}
		args = part[1:]
	}

	if err := cmd.Validate(args); err != nil {
		return "", err
	}
	return cmd.Run(args)
}

func (r *REPL) wrappedPrint(text string) {
	fmt.Fprintf(r.ui.Writer, "%s\n", text)
}

func (r *REPL) wrappedError(err error) {
	fmt.Fprintln(r.ui.ErrWriter, color.RedString("%s\n", err.Error()))
}

func (r *REPL) Start() error {
	r.printSplash(r.config.SplashTextPath)
	r.prompt.Run()
	return nil
}

func (r *REPL) Close() error {
	r.wrappedPrint("Good Bye :)")
	return nil
}

func (r *REPL) showHelp(cmds map[string]Commander) {
	var maxLen int
	// slice of [name, synopsis]
	text := make([][]string, len(cmds))
	for name, cmd := range cmds {
		text = append(text, []string{name, cmd.Synopsis()})
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	var cmdText string
	for name, cmd := range cmds {
		cmdText += fmt.Sprintf("  %-"+strconv.Itoa(maxLen)+"s    %s\n", name, cmd.Synopsis())
	}
	msg := fmt.Sprintf(`
Available commands:
%s

Show more details:
  <command> --help
`, cmdText)
	r.wrappedPrint(strings.TrimRight(msg, "\n"))
}

func (r *REPL) getPrompt() string {
	p := fmt.Sprintf("%s:%s> ", r.config.Server.Host, r.config.Server.Port)
	if dsn := r.env.GetDSN(); dsn != "" {
		p = fmt.Sprintf("%s@%s", dsn, p)
	}
	return p
}

func (r *REPL) printSplash(p string) {
	if p != "" {
		var abs string
		if strings.HasPrefix(p, "~/") {
			home, err := homedir.Dir()
			if err == nil {
				abs = filepath.Join(home, strings.TrimPrefix(p, "~/"))
			}
		} else {
			abs, _ = filepath.Abs(p)
		}
		if abs != "" {
			_, err := os.Stat(abs)
			if !os.IsNotExist(err) {
				b, err := ioutil.ReadFile(abs)
				if err == nil {
					fmt.Fprintln(r.ui.Writer, string(b))
				}
			}
		}
	}
}
