package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Manage API tokens",
	}

	cmd.AddCommand(newTokensListCmd())
	cmd.AddCommand(newTokensCreateCmd())
	cmd.AddCommand(newTokensRevokeCmd())
	return cmd
}

func newTokensListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List API tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/tokens", nil)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var resp TokenListResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			headers := []string{"ID", "NAME", "SCOPE", "ACTIVE", "EXPIRES AT", "CREATED AT"}
			var rows [][]string
			for _, t := range resp.Tokens {
				rows = append(rows, []string{
					t.ID,
					t.Name,
					t.Scope,
					fmt.Sprintf("%v", t.IsActive),
					t.ExpiresAt,
					t.CreatedAt,
				})
			}
			printTable(headers, rows)
			printMessage(fmt.Sprintf("\nTotal: %d tokens", resp.Total))
			return nil
		},
	}
}

func newTokensCreateCmd() *cobra.Command {
	var name, scope string
	var expiresInHours int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API token",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := CreateTokenRequest{
				Name:           name,
				Scope:          scope,
				ExpiresInHours: expiresInHours,
			}

			body, err := client.Post("/api/v1/tokens", req)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var resp CreateTokenResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage("Token created successfully!")
			printMessage(fmt.Sprintf("  ID:         %s", resp.ID))
			printMessage(fmt.Sprintf("  Name:       %s", resp.Name))
			printMessage(fmt.Sprintf("  Scope:      %s", resp.Scope))
			printMessage(fmt.Sprintf("  Expires At: %s", resp.ExpiresAt))
			printMessage("")
			printMessage(fmt.Sprintf("  Token: %s", resp.Token))
			printMessage("")
			printMessage("WARNING: This token will not be shown again. Save it now!")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Token name (required)")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&scope, "scope", "read_only", "Token scope: read_only or read_write")
	cmd.Flags().IntVar(&expiresInHours, "expires-in-hours", 0, "Token expiry in hours (0 for default)")
	return cmd
}

func newTokensRevokeCmd() *cobra.Command {
	var id string
	var yes bool

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an API token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirmAction(fmt.Sprintf("Revoke token %s?", id), yes) {
				printMessage("Aborted.")
				return nil
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			_, err = client.Delete(fmt.Sprintf("/api/v1/tokens/%s", id))
			if err != nil {
				return err
			}

			printMessage("Token revoked successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Token ID (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	return cmd
}
