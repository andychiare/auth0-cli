package cli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/spf13/cobra"

	"github.com/auth0/auth0-cli/internal/ansi"
)

func userBlocksCmd(cli *cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocks",
		Short: "Manage brute-force protection user blocks",
		Long:  "Manage brute-force protection user blocks.",
	}

	cmd.SetUsageTemplate(resourceUsageTemplate())
	cmd.AddCommand(listUserBlocksCmd(cli))
	cmd.AddCommand(deleteUserBlocksCmd(cli))

	return cmd
}

func listUserBlocksCmd(cli *cli) *cobra.Command {
	var inputs struct {
		userIdentifier string
	}

	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.MaximumNArgs(1),
		Short: "List brute-force protection blocks for a given user",
		Long:  "List brute-force protection blocks for a given user by user ID, username, phone number or email.",
		Example: `  auth0 users blocks list <user-id|username|email|phone-number>
  auth0 users blocks list <user-id|username|email|phone-number> --json
  auth0 users blocks list "auth0|61b5b6e90783fa19f7c57dad"
  auth0 users blocks list "frederik@travel0.com"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				if err := userIdentifier.Ask(cmd, &inputs.userIdentifier); err != nil {
					return err
				}
			} else {
				inputs.userIdentifier = args[0]
			}

			var userBlocks []*management.UserBlock
			err := ansi.Waiting(func() (err error) {
				userBlocks, err = cli.api.User.Blocks(cmd.Context(), inputs.userIdentifier)
				if mErr, ok := err.(management.Error); ok && mErr.Status() != http.StatusBadRequest {
					return err
				}

				userBlocks, err = cli.api.User.BlocksByIdentifier(cmd.Context(), inputs.userIdentifier)
				return err
			})
			if err != nil {
				return fmt.Errorf("failed to list user blocks for user with ID %s: %w", inputs.userIdentifier, err)
			}

			cli.renderer.UserBlocksList(userBlocks)
			return nil
		},
	}

	cmd.Flags().BoolVar(&cli.json, "json", false, "Output in json format.")

	return cmd
}

func deleteUserBlocksCmd(cli *cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unblock",
		Short: "Remove brute-force protection blocks for users",
		Long:  "Remove brute-force protection blocks for users by user ID, username, phone number or email.",
		Example: `  auth0 users blocks unblock <user-id1|username1|email1|phone-number1> <user-id2|username2|email2|phone-number2>
  auth0 users blocks unblock "auth0|61b5b6e90783fa19f7c57dad"
  auth0 users blocks unblock "frederik@travel0.com" "poovam@travel0.com"
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			identifiers := make([]string, len(args))
			if len(args) == 0 {
				var id string
				if err := userIdentifier.Ask(cmd, &id); err != nil {
					return err
				}
				identifiers = append(identifiers, id)
			} else {
				identifiers = append(identifiers, args...)
			}

			return ansi.Spinner("Unblocking user(s)...", func() error {
				var errs []error
				for _, identifier := range identifiers {
					if identifier != "" {
						err := cli.api.User.Unblock(cmd.Context(), identifier)
						if mErr, ok := err.(management.Error); ok && mErr.Status() != http.StatusBadRequest {
							errs = append(errs, fmt.Errorf("failed to unblock user with identifier %s: %w", identifier, err))
							continue
						}

						err = cli.api.User.UnblockByIdentifier(cmd.Context(), identifier)
						if err != nil {
							errs = append(errs, fmt.Errorf("failed to unblock user with identifier %s: %w", identifier, err))
						}
					}
				}
				return errors.Join(errs...)
			})
		},
	}

	return cmd
}
