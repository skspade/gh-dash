package actionssection

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/actionsrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

const SectionType = "action"

type Model struct {
	section.BaseModel
	Runs  []actionsrow.Data
	Repos []string
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.ActionsSectionConfig,
	lastUpdated time.Time,
	createdAt time.Time,
) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg.ToSectionConfig(),
			Type:        SectionType,
			Columns:     GetSectionColumns(cfg, ctx),
			Singular:    m.GetItemSingularForm(),
			Plural:      m.GetItemPluralForm(),
			LastUpdated: lastUpdated,
			CreatedAt:   createdAt,
		},
	)
	m.Runs = []actionsrow.Data{}
	m.Repos = cfg.Repos

	return m
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		if m.IsSearchFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return m, blinkCmd

			case tea.KeyEnter:
				m.SearchValue = m.SearchBar.Value()
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

			break
		}

		if m.IsPromptConfirmationFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.PromptConfirmationBox.Reset()
				cmd = m.SetIsPromptConfirmationShown(false)
				return m, cmd

			case tea.KeyEnter:
				input := m.PromptConfirmationBox.Value()
				action := m.GetPromptConfirmationAction()
				row := m.GetCurrRow()
				if row != nil && (input == "Y" || input == "y") {
					runData := row.(*actionsrow.Data)
					switch action {
					case "rerun":
						cmd = m.rerunWorkflow(runData.Run)
					case "cancel":
						cmd = m.cancelWorkflow(runData.Run)
					}
				}

				m.PromptConfirmationBox.Reset()
				blinkCmd := m.SetIsPromptConfirmationShown(false)

				return m, tea.Batch(cmd, blinkCmd)
			}

			break
		}

		switch {
		case key.Matches(msg, keys.ActionsKeys.Rerun):
			if m.GetCurrRow() != nil {
				m.SetPromptConfirmationAction("rerun")
				cmd = m.SetIsPromptConfirmationShown(true)
			}

		case key.Matches(msg, keys.ActionsKeys.Cancel):
			if m.GetCurrRow() != nil {
				m.SetPromptConfirmationAction("cancel")
				cmd = m.SetIsPromptConfirmationShown(true)
			}
		}

	case tasks.UpdateWorkflowRunMsg:
		for i, currRun := range m.Runs {
			if currRun.Run.Id != msg.RunId {
				continue
			}

			if msg.IsCancelled != nil && *msg.IsCancelled {
				currRun.Run.Status = "completed"
				currRun.Run.Conclusion = "cancelled"
			}
			if msg.IsRerun != nil && *msg.IsRerun {
				currRun.Run.Status = "queued"
				currRun.Run.Conclusion = ""
			}
			m.Runs[i] = currRun
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			break
		}

	case SectionWorkflowRunsFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			m.Runs = msg.Runs
			m.TotalCount = msg.TotalCount
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			m.Table.UpdateLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.Table.SetRows(m.BuildRows())
	m.SearchBar = search

	prompt, promptCmd := m.PromptConfirmationBox.Update(msg)
	m.PromptConfirmationBox = prompt

	table, tableCmd := m.Table.Update(msg)
	m.Table = table

	return m, tea.Batch(cmd, searchCmd, promptCmd, tableCmd)
}

func GetSectionColumns(
	cfg config.ActionsSectionConfig,
	ctx *context.ProgramContext,
) []table.Column {
	sLayout := cfg.Layout

	statusLayout := sLayout.Status
	repoLayout := sLayout.Repo
	workflowLayout := sLayout.Workflow
	branchLayout := sLayout.Branch
	commitShaLayout := sLayout.CommitSha
	actorLayout := sLayout.Actor
	durationLayout := sLayout.Duration
	updatedAtLayout := sLayout.UpdatedAt
	eventLayout := sLayout.Event

	return []table.Column{
		{
			Title:  "",
			Width:  utils.IntPtr(3),
			Hidden: statusLayout.Hidden,
		},
		{
			Title:  "Repo",
			Width:  utils.IntPtr(utils.Max(utils.IntPtrOr(repoLayout.Width, 20), 10)),
			Hidden: repoLayout.Hidden,
		},
		{
			Title:  "Workflow",
			Width:  utils.IntPtr(utils.Max(utils.IntPtrOr(workflowLayout.Width, 20), 10)),
			Hidden: workflowLayout.Hidden,
		},
		{
			Title:  "Branch",
			Width:  utils.IntPtr(utils.Max(utils.IntPtrOr(branchLayout.Width, 15), 8)),
			Hidden: branchLayout.Hidden,
		},
		{
			Title:  "Commit",
			Width:  utils.IntPtr(utils.Max(utils.IntPtrOr(commitShaLayout.Width, 8), 7)),
			Hidden: commitShaLayout.Hidden,
		},
		{
			Title:  "Actor",
			Width:  utils.IntPtr(utils.Max(utils.IntPtrOr(actorLayout.Width, 12), 8)),
			Hidden: actorLayout.Hidden,
		},
		{
			Title:  "Duration",
			Width:  utils.IntPtr(utils.Max(utils.IntPtrOr(durationLayout.Width, 8), 5)),
			Hidden: durationLayout.Hidden,
		},
		{
			Title:  "Updated",
			Width:  utils.IntPtr(utils.Max(utils.IntPtrOr(updatedAtLayout.Width, 6), 4)),
			Hidden: updatedAtLayout.Hidden,
		},
		{
			Title:  "Event",
			Width:  utils.IntPtr(utils.Max(utils.IntPtrOr(eventLayout.Width, 12), 6)),
			Hidden: eventLayout.Hidden,
		},
	}
}

func (m Model) BuildRows() []table.Row {
	var rows []table.Row
	currItem := m.Table.GetCurrItem()
	for i, currRun := range m.Runs {
		runModel := actionsrow.WorkflowRun{
			Ctx:     m.Ctx,
			Data:    &currRun,
			Columns: m.Table.Columns,
		}
		rows = append(
			rows,
			runModel.ToTableRow(currItem == i),
		)
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func (m *Model) NumRows() int {
	return len(m.Runs)
}

type SectionWorkflowRunsFetchedMsg struct {
	Runs       []actionsrow.Data
	TotalCount int
	TaskId     string
}


func (m *Model) GetCurrRow() data.RowData {
	if len(m.Runs) == 0 {
		return nil
	}
	run := m.Runs[m.Table.GetCurrItem()]
	return &run
}

func (m *Model) FetchNextPageSectionRows() []tea.Cmd {
	if m == nil {
		return nil
	}

	// No pagination for Actions - we always fetch fresh data
	var cmds []tea.Cmd

	taskId := fmt.Sprintf("fetching_actions_%d_%s", m.Id, time.Now().String())
	isFirstFetch := m.LastFetchTaskId == ""
	m.LastFetchTaskId = taskId
	task := context.Task{
		Id:        taskId,
		StartText: fmt.Sprintf(`Fetching Actions for "%s"`, m.Config.Title),
		FinishedText: fmt.Sprintf(
			`Actions for "%s" have been fetched`,
			m.Config.Title,
		),
		State: context.TaskStart,
		Error: nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	fetchCmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.ActionsLimit
		}

		res, err := data.FetchLatestWorkflowRuns(m.Repos, *limit)
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Err:         err,
			}
		}

		runs := make([]actionsrow.Data, 0)
		for i := range res.Runs {
			runs = append(runs, actionsrow.Data{Run: &res.Runs[i]})
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionWorkflowRunsFetchedMsg{
				Runs:       runs,
				TotalCount: res.TotalCount,
				TaskId:     taskId,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	m.IsLoading = true
	if isFirstFetch {
		m.SetIsLoading(true)
		cmds = append(cmds, m.Table.StartLoadingSpinner())
	}

	return cmds
}

func (m *Model) ResetRows() {
	m.Runs = nil
	m.BaseModel.ResetRows()
}

func FetchAllSections(
	ctx *context.ProgramContext,
	actions []section.Section,
) (sections []section.Section, fetchAllCmd tea.Cmd) {
	fetchActionsCmds := make([]tea.Cmd, 0, len(ctx.Config.ActionsSections))
	sections = make([]section.Section, 0, len(ctx.Config.ActionsSections))
	for i, sectionConfig := range ctx.Config.ActionsSections {
		sectionModel := NewModel(
			i+1,
			ctx,
			sectionConfig,
			time.Now(),
			time.Now(),
		)
		if len(actions) > 0 && len(actions) >= i+1 && actions[i+1] != nil {
			oldSection := actions[i+1].(*Model)
			sectionModel.Runs = oldSection.Runs
			sectionModel.LastFetchTaskId = oldSection.LastFetchTaskId
		}
		sections = append(sections, &sectionModel)
		fetchActionsCmds = append(
			fetchActionsCmds,
			sectionModel.FetchNextPageSectionRows()...)
	}
	return sections, tea.Batch(fetchActionsCmds...)
}

func (m Model) GetItemSingularForm() string {
	return "Workflow"
}

func (m Model) GetItemPluralForm() string {
	return "Workflows"
}

func (m Model) GetTotalCount() int {
	return m.TotalCount
}

func (m *Model) SetIsLoading(val bool) {
	m.IsLoading = val
	m.Table.SetIsLoading(val)
}

func (m Model) GetPagerContent() string {
	pagerContent := ""
	timeElapsed := utils.TimeElapsed(m.LastUpdated())
	if timeElapsed == "now" {
		timeElapsed = "just now"
	} else {
		timeElapsed = fmt.Sprintf("~%v ago", timeElapsed)
	}
	if m.TotalCount > 0 {
		pagerContent = fmt.Sprintf(
			"%v Updated %v • %v %v/%v",
			constants.WaitingIcon,
			timeElapsed,
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
		)
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}

// rerunWorkflow re-runs a workflow using gh CLI
func (m *Model) rerunWorkflow(run *data.WorkflowRunData) tea.Cmd {
	return tasks.RerunWorkflow(m.Ctx, tasks.SectionIdentifier{Id: m.Id, Type: SectionType}, run)
}

// cancelWorkflow cancels a running workflow using gh CLI
func (m *Model) cancelWorkflow(run *data.WorkflowRunData) tea.Cmd {
	return tasks.CancelWorkflow(m.Ctx, tasks.SectionIdentifier{Id: m.Id, Type: SectionType}, run)
}
