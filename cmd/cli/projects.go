package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects",
	}

	cmd.AddCommand(newProjectsListCmd())
	cmd.AddCommand(newProjectsCreateCmd())
	cmd.AddCommand(newProjectsGetCmd())
	cmd.AddCommand(newProjectsUpdateCmd())
	cmd.AddCommand(newProjectsDeleteCmd())
	return cmd
}

func newProjectsListCmd() *cobra.Command {
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			query := url.Values{}
			if limit > 0 {
				query.Set("limit", strconv.Itoa(limit))
			}
			if offset > 0 {
				query.Set("offset", strconv.Itoa(offset))
			}

			body, err := client.Get("/api/v1/projects", query)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var resp PaginatedResponse[ProjectResponse]
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			headers := []string{"ID", "NAME", "DESCRIPTION", "CREATED AT"}
			var rows [][]string
			for _, p := range resp.Items {
				rows = append(rows, []string{
					p.ID.String(),
					p.Name,
					truncate(p.Description, 40),
					p.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
			printTable(headers, rows)
			printMessage(fmt.Sprintf("\nShowing %d of %d projects", len(resp.Items), resp.Total))
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
	return cmd
}

func newProjectsCreateCmd() *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := CreateProjectRequest{
				Name:        name,
				Description: description,
			}

			body, err := client.Post("/api/v1/projects", req)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var p ProjectResponse
			if err := json.Unmarshal(body, &p); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("Project created: %s (%s)", p.Name, p.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Project name (required)")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&description, "description", "", "Project description")
	return cmd
}

func newProjectsGetCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a project by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			body, err := client.Get(fmt.Sprintf("/api/v1/projects/%s", id), nil)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var p ProjectResponse
			if err := json.Unmarshal(body, &p); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"ID", p.ID.String()},
				{"Name", p.Name},
				{"Description", p.Description},
				{"Owner ID", p.OwnerID.String()},
				{"Active", fmt.Sprintf("%v", p.IsActive)},
				{"Created At", p.CreatedAt.Format("2006-01-02 15:04:05")},
				{"Updated At", p.UpdatedAt.Format("2006-01-02 15:04:05")},
			}
			printTable(headers, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Project ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newProjectsUpdateCmd() *cobra.Command {
	var id, name, description string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := UpdateProjectRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				req.Description = &description
			}

			body, err := client.Put(fmt.Sprintf("/api/v1/projects/%s", id), req)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var p ProjectResponse
			if err := json.Unmarshal(body, &p); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("Project updated: %s (%s)", p.Name, p.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Project ID (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVar(&name, "name", "", "New project name")
	cmd.Flags().StringVar(&description, "description", "", "New project description")
	return cmd
}

func newProjectsDeleteCmd() *cobra.Command {
	var id string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirmAction(fmt.Sprintf("Delete project %s?", id), yes) {
				printMessage("Aborted.")
				return nil
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			_, err = client.Delete(fmt.Sprintf("/api/v1/projects/%s", id))
			if err != nil {
				return err
			}

			printMessage("Project deleted successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Project ID (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
