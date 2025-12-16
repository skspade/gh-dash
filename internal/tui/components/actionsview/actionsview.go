package actionsview

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/actionsrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Model struct {
	ctx       *context.ProgramContext
	sectionId int
	run       *actionsrow.Data
	width     int
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		run: nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	if m.run == nil || m.run.Run == nil {
		return "No workflow run selected"
	}

	s := strings.Builder{}

	// Header: repo and run number
	s.WriteString(m.renderHeader())
	s.WriteString("\n")

	// Title/display title
	s.WriteString(m.renderTitle())
	s.WriteString("\n\n")

	// Status with icon
	s.WriteString(m.renderStatus())
	s.WriteString("\n\n")

	// Workflow name
	s.WriteString(m.renderWorkflowName())
	s.WriteString("\n\n")

	// Branch and commit
	s.WriteString(m.renderBranchAndCommit())
	s.WriteString("\n\n")

	// Event and actor
	s.WriteString(m.renderEventAndActor())
	s.WriteString("\n\n")

	// Timing info
	s.WriteString(m.renderTiming())
	s.WriteString("\n\n")

	// Separator
	s.WriteString(lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintBorder).Render(
		strings.Repeat(lipgloss.NormalBorder().Bottom, m.width-5)),
	)
	s.WriteString("\n\n")

	// Actions hint
	s.WriteString(m.renderActionsHint())

	return s.String()
}

func (m *Model) renderHeader() string {
	run := m.run.Run
	return lipgloss.NewStyle().
		PaddingLeft(1).
		Width(m.width).
		Background(m.ctx.Theme.SelectedBackground).
		Foreground(m.ctx.Theme.SecondaryText).
		Render(fmt.Sprintf("%s · Run #%d", run.Repository.FullName, run.RunNumber))
}

func (m *Model) renderTitle() string {
	run := m.run.Run
	title := run.DisplayTitle
	if title == "" {
		title = run.Name
	}
	return lipgloss.NewStyle().
		Height(3).
		Width(m.width).
		Background(m.ctx.Theme.SelectedBackground).
		PaddingLeft(1).
		Render(lipgloss.PlaceVertical(3, lipgloss.Center,
			m.ctx.Styles.Common.MainTextStyle.
				Background(m.ctx.Theme.SelectedBackground).
				Render(title)))
}

func (m *Model) renderStatus() string {
	run := m.run.Run
	icon, color := m.getStatusIconAndColor()

	statusText := getStatusText(run)
	pill := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Render(fmt.Sprintf("%s %s", icon, statusText))

	return lipgloss.JoinHorizontal(lipgloss.Left,
		" ",
		pill,
	)
}

func (m *Model) getStatusIconAndColor() (string, string) {
	run := m.run.Run

	if run.Status == "in_progress" {
		return constants.WorkflowRunningIcon, m.ctx.Styles.Colors.OpenPR.Dark
	}
	if run.Status == "queued" {
		return constants.WorkflowQueuedIcon, "#58a6ff" // Blue
	}

	switch run.Conclusion {
	case "success":
		return constants.SuccessIcon, m.ctx.Styles.Colors.OpenPR.Dark
	case "failure":
		return constants.FailureIcon, m.ctx.Styles.Colors.ClosedPR.Dark
	case "cancelled":
		return constants.WorkflowCancelledIcon, m.ctx.Theme.FaintText.Dark
	case "skipped":
		return constants.WorkflowSkippedIcon, m.ctx.Theme.FaintText.Dark
	default:
		return constants.WaitingIcon, m.ctx.Theme.FaintText.Dark
	}
}

func getStatusText(run *data.WorkflowRunData) string {
	if run.Status == "in_progress" {
		return "In Progress"
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
	default:
		if len(run.Status) == 0 {
			return "Unknown"
		}
		return strings.ToUpper(run.Status[:1]) + run.Status[1:]
	}
}

func (m *Model) renderWorkflowName() string {
	run := m.run.Run
	return lipgloss.JoinVertical(lipgloss.Left,
		m.ctx.Styles.Common.MainTextStyle.Bold(true).Underline(true).Render("Workflow"),
		"",
		lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText).Render(run.Name),
	)
}

func (m *Model) renderBranchAndCommit() string {
	run := m.run.Run

	branchStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.SecondaryText).
		Background(lipgloss.Color("#2d333b")).
		Padding(0, 1)

	shaStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.FaintText)

	return lipgloss.JoinVertical(lipgloss.Left,
		m.ctx.Styles.Common.MainTextStyle.Bold(true).Underline(true).Render("Branch & Commit"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top,
			" ",
			branchStyle.Render(run.HeadBranch),
			" ",
			shaStyle.Render(run.ShortSha()),
		),
	)
}

func (m *Model) renderEventAndActor() string {
	run := m.run.Run

	eventDisplay := run.Event
	switch run.Event {
	case "pull_request":
		eventDisplay = "Pull Request"
	case "push":
		eventDisplay = "Push"
	case "schedule":
		eventDisplay = "Schedule"
	case "workflow_dispatch":
		eventDisplay = "Manual"
	case "release":
		eventDisplay = "Release"
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.ctx.Styles.Common.MainTextStyle.Bold(true).Underline(true).Render("Triggered"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText).Render(eventDisplay),
			lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(" by "),
			lipgloss.NewStyle().Bold(true).Foreground(m.ctx.Theme.PrimaryText).Render("@"+run.Actor.Login),
		),
	)
}

func (m *Model) renderTiming() string {
	run := m.run.Run

	// Time elapsed since start
	timeAgo := utils.TimeElapsed(run.CreatedAt)

	// Duration
	duration := run.Duration()
	durationStr := formatDuration(duration)

	return lipgloss.JoinVertical(lipgloss.Left,
		m.ctx.Styles.Common.MainTextStyle.Bold(true).Underline(true).Render("Timing"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("Started: "),
			lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText).Render(timeAgo+" ago"),
		),
		lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("Duration: "),
			lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText).Render(durationStr),
		),
	)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func (m *Model) renderActionsHint() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.ctx.Styles.Common.MainTextStyle.Bold(true).Underline(true).Render("Actions"),
		"",
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("o  Open in GitHub"),
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("r  Re-run workflow"),
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("c  Cancel workflow"),
	)
}

func (m *Model) SetSectionId(id int) {
	m.sectionId = id
}

func (m *Model) SetRow(d *actionsrow.Data) {
	m.run = d
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}
