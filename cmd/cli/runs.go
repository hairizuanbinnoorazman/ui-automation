package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hairizuanbinnoorazman/ui-automation/testrun"
	"github.com/spf13/cobra"
)

func newRunsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Manage test runs",
	}

	cmd.AddCommand(newRunsListCmd())
	cmd.AddCommand(newRunsCreateCmd())
	cmd.AddCommand(newRunsGetCmd())
	cmd.AddCommand(newRunsUpdateCmd())
	cmd.AddCommand(newRunsStartCmd())
	cmd.AddCommand(newRunsCompleteCmd())
	return cmd
}

func newRunsListCmd() *cobra.Command {
	var procedureID string
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List test runs for a procedure",
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

			body, err := client.Get(fmt.Sprintf("/api/v1/procedures/%s/runs", procedureID), query)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var resp PaginatedResponse[TestRunResponse]
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			headers := []string{"ID", "PROCEDURE ID", "STATUS", "VERSION", "STARTED AT", "COMPLETED AT"}
			var rows [][]string
			for _, r := range resp.Items {
				startedAt := "-"
				if r.StartedAt != nil {
					startedAt = r.StartedAt.Format("2006-01-02 15:04:05")
				}
				completedAt := "-"
				if r.CompletedAt != nil {
					completedAt = r.CompletedAt.Format("2006-01-02 15:04:05")
				}
				rows = append(rows, []string{
					r.ID.String(),
					r.TestProcedureID.String(),
					string(r.Status),
					fmt.Sprintf("v%d", r.ProcedureVersion),
					startedAt,
					completedAt,
				})
			}
			printTable(headers, rows)
			printMessage(fmt.Sprintf("\nShowing %d of %d runs", len(resp.Items), resp.Total))
			return nil
		},
	}

	cmd.Flags().StringVar(&procedureID, "procedure-id", "", "Test procedure ID (required)")
	cmd.MarkFlagRequired("procedure-id")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
	return cmd
}

func newRunsCreateCmd() *cobra.Command {
	var procedureID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new test run",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			body, err := client.Post(fmt.Sprintf("/api/v1/procedures/%s/runs", procedureID), nil)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var r TestRunResponse
			if err := json.Unmarshal(body, &r); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("Test run created: %s (status: %s)", r.ID, r.Status))
			return nil
		},
	}

	cmd.Flags().StringVar(&procedureID, "procedure-id", "", "Test procedure ID (required)")
	cmd.MarkFlagRequired("procedure-id")
	return cmd
}

func newRunsGetCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a test run by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			body, err := client.Get(fmt.Sprintf("/api/v1/runs/%s", id), nil)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var r TestRunResponse
			if err := json.Unmarshal(body, &r); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			startedAt := "-"
			if r.StartedAt != nil {
				startedAt = r.StartedAt.Format("2006-01-02 15:04:05")
			}
			completedAt := "-"
			if r.CompletedAt != nil {
				completedAt = r.CompletedAt.Format("2006-01-02 15:04:05")
			}
			assignedTo := "-"
			if r.AssignedTo != nil {
				assignedTo = r.AssignedTo.String()
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"ID", r.ID.String()},
				{"Procedure ID", r.TestProcedureID.String()},
				{"Status", string(r.Status)},
				{"Executed By", r.ExecutedBy.String()},
				{"Assigned To", assignedTo},
				{"Notes", r.Notes},
				{"Started At", startedAt},
				{"Completed At", completedAt},
				{"Created At", r.CreatedAt.Format("2006-01-02 15:04:05")},
			}
			printTable(headers, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Test run ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newRunsUpdateCmd() *cobra.Command {
	var id, notes, assignedTo string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a test run",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := UpdateTestRunRequest{}
			if cmd.Flags().Changed("notes") {
				req.Notes = &notes
			}
			if cmd.Flags().Changed("assigned-to") {
				req.AssignedTo = &assignedTo
			}

			body, err := client.Put(fmt.Sprintf("/api/v1/runs/%s", id), req)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var r TestRunResponse
			if err := json.Unmarshal(body, &r); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("Test run updated: %s (status: %s)", r.ID, r.Status))
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Test run ID (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVar(&notes, "notes", "", "Test run notes")
	cmd.Flags().StringVar(&assignedTo, "assigned-to", "", "User ID to assign to (empty string to unassign)")
	return cmd
}

func newRunsStartCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a test run",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			body, err := client.Post(fmt.Sprintf("/api/v1/runs/%s/start", id), nil)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var r TestRunResponse
			if err := json.Unmarshal(body, &r); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("Test run started: %s", r.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Test run ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newRunsCompleteCmd() *cobra.Command {
	var id, status, notes string

	cmd := &cobra.Command{
		Use:   "complete",
		Short: "Complete a test run",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := testrun.Status(status)
			if !s.IsFinal() {
				return fmt.Errorf("invalid status: must be passed, failed, or skipped")
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			req := CompleteTestRunRequest{
				Status: s,
				Notes:  notes,
			}

			body, err := client.Post(fmt.Sprintf("/api/v1/runs/%s/complete", id), req)
			if err != nil {
				return err
			}

			if flagJSON {
				var raw json.RawMessage
				json.Unmarshal(body, &raw)
				printJSON(raw)
				return nil
			}

			var r TestRunResponse
			if err := json.Unmarshal(body, &r); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			printMessage(fmt.Sprintf("Test run completed: %s (status: %s)", r.ID, r.Status))
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Test run ID (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVar(&status, "status", "", "Final status: passed, failed, or skipped (required)")
	cmd.MarkFlagRequired("status")
	cmd.Flags().StringVar(&notes, "notes", "", "Completion notes")
	return cmd
}
