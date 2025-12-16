package tasks

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type UpdateWorkflowRunMsg struct {
	RunId       int64
	IsCancelled *bool
	IsRerun     *bool
}

func RerunWorkflow(ctx *context.ProgramContext, section SectionIdentifier, run *data.WorkflowRunData) tea.Cmd {
	return fireTask(ctx, GitHubTask{
		Id: fmt.Sprintf("workflow_rerun_%d", run.Id),
		Args: []string{
			"run",
			"rerun",
			fmt.Sprintf("%d", run.Id),
			"-R",
			run.Repository.FullName,
		},
		Section:      section,
		StartText:    fmt.Sprintf("Re-running workflow %s #%d", run.Name, run.RunNumber),
		FinishedText: fmt.Sprintf("Workflow %s #%d has been re-run", run.Name, run.RunNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			if err != nil {
				return UpdateWorkflowRunMsg{}
			}
			return UpdateWorkflowRunMsg{
				RunId:   run.Id,
				IsRerun: utils.BoolPtr(true),
			}
		},
	})
}

func CancelWorkflow(ctx *context.ProgramContext, section SectionIdentifier, run *data.WorkflowRunData) tea.Cmd {
	return fireTask(ctx, GitHubTask{
		Id: fmt.Sprintf("workflow_cancel_%d", run.Id),
		Args: []string{
			"run",
			"cancel",
			fmt.Sprintf("%d", run.Id),
			"-R",
			run.Repository.FullName,
		},
		Section:      section,
		StartText:    fmt.Sprintf("Cancelling workflow %s #%d", run.Name, run.RunNumber),
		FinishedText: fmt.Sprintf("Workflow %s #%d has been cancelled", run.Name, run.RunNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			if err != nil {
				return UpdateWorkflowRunMsg{}
			}
			return UpdateWorkflowRunMsg{
				RunId:       run.Id,
				IsCancelled: utils.BoolPtr(true),
			}
		},
	})
}
