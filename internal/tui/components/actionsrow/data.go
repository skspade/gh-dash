package actionsrow

import (
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

type Data struct {
	Run *data.WorkflowRunData
}

func (d Data) GetTitle() string {
	return d.Run.DisplayTitle
}

func (d Data) GetRepoNameWithOwner() string {
	return d.Run.Repository.FullName
}

func (d Data) GetNumber() int {
	return d.Run.RunNumber
}

func (d Data) GetUrl() string {
	return d.Run.HtmlUrl
}

func (d Data) GetUpdatedAt() time.Time {
	return d.Run.UpdatedAt
}

func (d Data) GetCreatedAt() time.Time {
	return d.Run.CreatedAt
}
