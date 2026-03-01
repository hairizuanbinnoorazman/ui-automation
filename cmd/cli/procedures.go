package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func newProceduresCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "procedures",
		Short: "Manage test procedures",
	}

	cmd.AddCommand(newProceduresListCmd())
	cmd.AddCommand(newProceduresCreateCmd())
	cmd.AddCommand(newProceduresGetCmd())
	cmd.AddCommand(newProceduresUpdateCmd())
	cmd.AddCommand(newProceduresDeleteCmd())
	cmd.AddCommand(newProceduresCreateVersionCmd())
	cmd.AddCommand(newProceduresVersionsCmd())
	return cmd
}

func newProceduresListCmd() *cobra.Command {
	var projectID string
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List test procedures for a project",
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

			body, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/procedures", projectID), query)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var resp PaginatedResponse[TestProcedureResponse]
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			headers := []string{"ID", "NAME", "VERSION", "IS LATEST", "CREATED AT"}
			var rows [][]string
			for _, p := range resp.Items {
				rows = append(rows, []string{
					p.ID.String(),
					p.Name,
					strconv.Itoa(int(p.Version)),
					fmt.Sprintf("%v", p.IsLatest),
					p.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
			printTable(headers, rows)
			printMessage(fmt.Sprintf("\nShowing %d of %d procedures", len(resp.Items), resp.Total))
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (required)")
	cmd.MarkFlagRequired("project-id")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
	return cmd
}

func newProceduresCreateCmd() *cobra.Command {
	var projectID, name, description, stepsFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new test procedure",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := CreateTestProcedureRequest{
				Name:        name,
				Description: description,
			}

			if stepsFile != "" {
				data, err := os.ReadFile(stepsFile)
				if err != nil {
					return fmt.Errorf("failed to read steps file: %w", err)
				}
				if err := json.Unmarshal(data, &req.Steps); err != nil {
					return fmt.Errorf("failed to parse steps file: %w", err)
				}
			}

			body, err := client.Post(fmt.Sprintf("/api/v1/projects/%s/procedures", projectID), req)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var p TestProcedureResponse
			if err := json.Unmarshal(body, &p); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("Test procedure created: %s (%s)", p.Name, p.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (required)")
	cmd.MarkFlagRequired("project-id")
	cmd.Flags().StringVar(&name, "name", "", "Procedure name (required)")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&description, "description", "", "Procedure description")
	cmd.Flags().StringVar(&stepsFile, "steps-file", "", "JSON file containing steps array")
	return cmd
}

func newProceduresGetCmd() *cobra.Command {
	var projectID, id string
	var draft bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a test procedure by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			query := url.Values{}
			if draft {
				query.Set("draft", "true")
			}

			body, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/procedures/%s", projectID, id), query)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var p TestProcedureResponse
			if err := json.Unmarshal(body, &p); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"ID", p.ID.String()},
				{"Name", p.Name},
				{"Description", p.Description},
				{"Project ID", p.ProjectID.String()},
				{"Version", strconv.Itoa(int(p.Version))},
				{"Is Latest", fmt.Sprintf("%v", p.IsLatest)},
				{"Steps", strconv.Itoa(len(p.Steps))},
				{"Created At", p.CreatedAt.Format("2006-01-02 15:04:05")},
				{"Updated At", p.UpdatedAt.Format("2006-01-02 15:04:05")},
			}
			printTable(headers, rows)

			if len(p.Steps) > 0 {
				printMessage("\nSteps:")
				for i, step := range p.Steps {
					printMessage(fmt.Sprintf("  %d. %s", i+1, step.Name))
					if step.Instructions != "" {
						printMessage(fmt.Sprintf("     %s", step.Instructions))
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (required)")
	cmd.MarkFlagRequired("project-id")
	cmd.Flags().StringVar(&id, "id", "", "Procedure ID (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().BoolVar(&draft, "draft", false, "Get draft version")
	return cmd
}

func newProceduresUpdateCmd() *cobra.Command {
	var projectID, id, name, description, stepsFile string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a test procedure draft",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := UpdateTestProcedureRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				req.Description = &description
			}
			if stepsFile != "" {
				data, err := os.ReadFile(stepsFile)
				if err != nil {
					return fmt.Errorf("failed to read steps file: %w", err)
				}
				var steps []struct {
					Name         string   `json:"name"`
					Instructions string   `json:"instructions"`
					ImagePaths   []string `json:"image_paths"`
				}
				if err := json.Unmarshal(data, &steps); err != nil {
					return fmt.Errorf("failed to parse steps file: %w", err)
				}
				req.Steps = &steps
			}

			body, err := client.Put(fmt.Sprintf("/api/v1/projects/%s/procedures/%s", projectID, id), req)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var p TestProcedureResponse
			if err := json.Unmarshal(body, &p); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("Test procedure updated: %s (%s)", p.Name, p.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (required)")
	cmd.MarkFlagRequired("project-id")
	cmd.Flags().StringVar(&id, "id", "", "Procedure ID (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVar(&name, "name", "", "New procedure name")
	cmd.Flags().StringVar(&description, "description", "", "New procedure description")
	cmd.Flags().StringVar(&stepsFile, "steps-file", "", "JSON file containing updated steps array")
	return cmd
}

func newProceduresDeleteCmd() *cobra.Command {
	var projectID, id string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a test procedure",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirmAction(fmt.Sprintf("Delete test procedure %s?", id), yes) {
				printMessage("Aborted.")
				return nil
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			_, err = client.Delete(fmt.Sprintf("/api/v1/projects/%s/procedures/%s", projectID, id))
			if err != nil {
				return err
			}

			printMessage("Test procedure deleted successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (required)")
	cmd.MarkFlagRequired("project-id")
	cmd.Flags().StringVar(&id, "id", "", "Procedure ID (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	return cmd
}

func newProceduresCreateVersionCmd() *cobra.Command {
	var projectID, id string

	cmd := &cobra.Command{
		Use:   "create-version",
		Short: "Commit the draft as a new version",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			body, err := client.Post(fmt.Sprintf("/api/v1/projects/%s/procedures/%s/versions", projectID, id), nil)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var p TestProcedureResponse
			if err := json.Unmarshal(body, &p); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("New version created: v%d (%s)", p.Version, p.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (required)")
	cmd.MarkFlagRequired("project-id")
	cmd.Flags().StringVar(&id, "id", "", "Procedure ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newProceduresVersionsCmd() *cobra.Command {
	var projectID, id string

	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List version history for a test procedure",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			body, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/procedures/%s/versions", projectID, id), nil)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var versions []TestProcedureResponse
			if err := json.Unmarshal(body, &versions); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			headers := []string{"ID", "VERSION", "NAME", "IS LATEST", "CREATED AT"}
			var rows [][]string
			for _, v := range versions {
				rows = append(rows, []string{
					v.ID.String(),
					strconv.Itoa(int(v.Version)),
					v.Name,
					fmt.Sprintf("%v", v.IsLatest),
					v.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
			printTable(headers, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (required)")
	cmd.MarkFlagRequired("project-id")
	cmd.Flags().StringVar(&id, "id", "", "Procedure ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}
