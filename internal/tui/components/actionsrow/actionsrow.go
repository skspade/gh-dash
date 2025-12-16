package actionsrow

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type WorkflowRun struct {
	Ctx     *context.ProgramContext
	Data    *Data
	Columns []table.Column
}

func (w *WorkflowRun) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(w.Ctx)
}

func (w *WorkflowRun) renderStatus() string {
	run := w.Data.Run
	style := lipgloss.NewStyle()

	// Handle in-progress and queued states
	if run.Status == "in_progress" {
		return style.Foreground(w.Ctx.Theme.WarningText).Render(constants.WorkflowRunningIcon)
	}
	if run.Status == "queued" {
		return style.Foreground(w.Ctx.Theme.FaintText).Render(constants.WorkflowQueuedIcon)
	}

	// For completed runs, check the conclusion
	switch run.Conclusion {
	case "success":
		return style.Foreground(w.Ctx.Theme.SuccessText).Render(constants.SuccessIcon)
	case "failure":
		return style.Foreground(w.Ctx.Theme.ErrorText).Render(constants.FailureIcon)
	case "cancelled":
		return style.Foreground(w.Ctx.Theme.FaintText).Render(constants.WorkflowCancelledIcon)
	case "skipped":
		return style.Foreground(w.Ctx.Theme.FaintText).Render(constants.WorkflowSkippedIcon)
	case "neutral":
		return style.Foreground(w.Ctx.Theme.FaintText).Render(constants.EmptyIcon)
	case "timed_out":
		return style.Foreground(w.Ctx.Theme.ErrorText).Render(constants.FailureIcon)
	case "action_required":
		return style.Foreground(w.Ctx.Theme.WarningText).Render(constants.WaitingIcon)
	default:
		return style.Foreground(w.Ctx.Theme.FaintText).Render(constants.EmptyIcon)
	}
}

func (w *WorkflowRun) renderRepoName() string {
	return w.getTextStyle().Foreground(w.Ctx.Theme.FaintText).Render(w.Data.Run.Repository.FullName)
}

func (w *WorkflowRun) renderWorkflowName() string {
	return w.getTextStyle().Render(w.Data.Run.Name)
}

func (w *WorkflowRun) renderBranch() string {
	return w.getTextStyle().Foreground(w.Ctx.Theme.FaintText).Render(w.Data.Run.HeadBranch)
}

func (w *WorkflowRun) renderCommitSha() string {
	return w.getTextStyle().Foreground(w.Ctx.Theme.FaintText).Render(w.Data.Run.ShortSha())
}

func (w *WorkflowRun) renderActor() string {
	return w.getTextStyle().Foreground(w.Ctx.Theme.FaintText).Render(w.Data.Run.Actor.Login)
}

func (w *WorkflowRun) renderDuration() string {
	duration := w.Data.Run.Duration()
	return w.getTextStyle().Foreground(w.Ctx.Theme.FaintText).Render(formatDuration(duration))
}

func (w *WorkflowRun) renderUpdatedAt() string {
	timeFormat := w.Ctx.Config.Defaults.DateFormat

	if timeFormat == "" || timeFormat == "relative" {
		return w.getTextStyle().Foreground(w.Ctx.Theme.FaintText).Render(utils.TimeElapsed(w.Data.Run.UpdatedAt))
	}

	return w.getTextStyle().Foreground(w.Ctx.Theme.FaintText).Render(w.Data.Run.UpdatedAt.Format(timeFormat))
}

func (w *WorkflowRun) renderEvent() string {
	return w.getTextStyle().Foreground(w.Ctx.Theme.FaintText).Render(w.Data.Run.Event)
}

func (w *WorkflowRun) renderTitle() string {
	title := w.Data.Run.DisplayTitle
	if title == "" {
		title = w.Data.Run.Name
	}
	return w.getTextStyle().Render(title)
}

func (w *WorkflowRun) ToTableRow(isSelected bool) table.Row {
	return table.Row{
		w.renderStatus(),
		w.renderRepoName(),
		w.renderWorkflowName(),
		w.renderBranch(),
		w.renderCommitSha(),
		w.renderActor(),
		w.renderDuration(),
		w.renderUpdatedAt(),
		w.renderEvent(),
	}
}

// GetStatusText returns a human-readable status text
func GetStatusText(run *data.WorkflowRunData) string {
	if run.Status == "in_progress" {
		return "Running"
	}
	if run.Status == "queued" {
		return "Queued"
	}

	switch run.Conclusion {
	case "success":
		return "Success"
	case "failure":
		return "Failed"
	case "cancelled":
		return "Cancelled"
	case "skipped":
		return "Skipped"
	case "neutral":
		return "Neutral"
	case "timed_out":
		return "Timed Out"
	case "action_required":
		return "Action Required"
	default:
		return run.Conclusion
	}
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		if secs > 0 {
			return fmt.Sprintf("%dm%ds", mins, secs)
		}
		return fmt.Sprintf("%dm", mins)
	}

	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if mins > 0 {
		return fmt.Sprintf("%dh%dm", hours, mins)
	}
	return fmt.Sprintf("%dh", hours)
}
