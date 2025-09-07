package api

import (
	"fmt"
	"hama-shell/internal/session/infra"
	"hama-shell/internal/session/model"
	"os"
	"text/tabwriter"
)

// SessionAPI provides high-level session operations
type SessionAPI struct {
	sessionMgr *infra.SessionManager
}

// NewSessionAPI creates a new SessionAPI instance
func NewSessionAPI() *SessionAPI {
	return &SessionAPI{
		sessionMgr: infra.NewSessionManager(),
	}
}

// ListSessions displays all sessions based on filter
func (api *SessionAPI) ListSessions(showAll bool, statusFilter string) error {
	filter := model.SessionFilter{
		ShowAll: showAll,
		Status:  statusFilter,
	}

	sessions, err := api.sessionMgr.ListSessions(filter)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		if statusFilter != "" {
			fmt.Printf("(Filtered by status: %s)\n", statusFilter)
		}
		if !showAll {
			fmt.Println("Use --all flag to show stopped sessions")
		}
		return nil
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SESSION ID\tSTATUS\tSTART TIME\tCOMMAND")
	fmt.Fprintln(w, "----------\t------\t----------\t-------")

	for _, session := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			session.ID,
			session.Status,
			session.StartTime.Format("2006-01-02 15:04:05"),
			session.Command,
		)
	}

	w.Flush()

	// Show session count
	fmt.Printf("\nTotal sessions: %d\n", len(sessions))

	if statusFilter != "" {
		fmt.Printf("(Filtered by status: %s)\n", statusFilter)
	}

	return nil
}
