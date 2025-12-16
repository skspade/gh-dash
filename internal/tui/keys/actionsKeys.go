package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

type ActionsKeyMap struct {
	Rerun   key.Binding
	Cancel  key.Binding
	ViewPRs key.Binding
}

var ActionsKeys = ActionsKeyMap{
	Rerun: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "re-run workflow"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "cancel workflow"),
	),
	ViewPRs: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch to PRs"),
	),
}

func ActionsFullHelp() []key.Binding {
	return []key.Binding{
		ActionsKeys.Rerun,
		ActionsKeys.Cancel,
		ActionsKeys.ViewPRs,
	}
}

func rebindActionsKeys(keys []config.Keybinding) error {
	CustomActionsBindings = []key.Binding{}

	for _, actionsKey := range keys {
		if actionsKey.Builtin == "" {
			// Handle custom commands
			if actionsKey.Command != "" {
				name := actionsKey.Name
				if actionsKey.Name == "" {
					name = config.TruncateCommand(actionsKey.Command)
				}

				customBinding := key.NewBinding(
					key.WithKeys(actionsKey.Key),
					key.WithHelp(actionsKey.Key, name),
				)

				CustomActionsBindings = append(CustomActionsBindings, customBinding)
			}
			continue
		}

		log.Debug("Rebinding Actions key", "builtin", actionsKey.Builtin, "key", actionsKey.Key)

		var key *key.Binding

		switch actionsKey.Builtin {
		case "rerun":
			key = &ActionsKeys.Rerun
		case "cancel":
			key = &ActionsKeys.Cancel
		case "viewPRs":
			key = &ActionsKeys.ViewPRs
		default:
			return fmt.Errorf("unknown built-in actions key: '%s'", actionsKey.Builtin)
		}

		key.SetKeys(actionsKey.Key)

		helpDesc := key.Help().Desc
		if actionsKey.Name != "" {
			helpDesc = actionsKey.Name
		}
		key.SetHelp(actionsKey.Key, helpDesc)
	}

	return nil
}
