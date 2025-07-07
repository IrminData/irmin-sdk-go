package irminmodels

type RepositoryEvent string

const (
	PreCommit        RepositoryEvent = "pre-commit"
	PostCommit       RepositoryEvent = "post-commit"
	PreMerge         RepositoryEvent = "pre-merge"
	PostMerge        RepositoryEvent = "post-merge"
	PreCreateBranch  RepositoryEvent = "pre-create-branch"
	PostCreateBranch RepositoryEvent = "post-create-branch"
	PreDeleteBranch  RepositoryEvent = "pre-delete-branch"
	PostDeleteBranch RepositoryEvent = "post-delete-branch"
	PreCreateTag     RepositoryEvent = "pre-create-tag"
	PostCreateTag    RepositoryEvent = "post-create-tag"
	PreDeleteTag     RepositoryEvent = "pre-delete-tag"
	PostDeleteTag    RepositoryEvent = "post-delete-tag"
)

type WorkflowRunEvent string

const (
	PreWorkflowRun  WorkflowRunEvent = "pre-workflow-run"
	PostWorkflowRun WorkflowRunEvent = "post-workflow-run"
)

type WorkflowTriggerType string

const (
	TimeTriggerType        WorkflowTriggerType = "time"
	RepositoryTriggerType  WorkflowTriggerType = "repository-event"
	WorkflowRunTriggerType WorkflowTriggerType = "workflow-run-event"
)

type ScheduleTrigger struct {
	Type WorkflowTriggerType `json:"type" validate:"required,oneof=time repository-event workflow-run-event"`

	// Time trigger
	RRule *string `json:"rrule,omitempty" validate:"validrrule,required_if=Type time,required_without=Cron"`
	Cron  *string `json:"cron,omitempty" validate:"validcron,required_if=Type time,required_without=RRule"`

	// Repository event trigger
	RepositoryEvent *RepositoryEvent `json:"repository_event,omitempty" validate:"required_if=Type repository-event,oneof=pre-commit post-commit pre-merge post-merge pre-create-branch post-create-branch pre-delete-branch post-delete-branch pre-create-tag post-create-tag pre-delete-tag post-delete-tag"`
	Repository      *string          `json:"repository,omitempty" validate:"required_with=RepositoryEvent,validslug"` // Slug of the repository
	RepositoryRef   *string          `json:"repository_ref,omitempty" validate:"required_with=RepositoryEvent"`

	// Workflow run event trigger
	WorkflowRunEvent *WorkflowRunEvent `json:"workflow_run_event,omitempty" validate:"required_if=Type workflow-run-event,oneof=pre-workflow-run post-workflow-run"`
	WorkflowID       *string           `json:"workflow_id,omitempty" validate:"required_with=WorkflowRunEvent,validsqid=workflows"` // Sqid of the workflow
}

type Schedule struct {
	Triggers    []ScheduleTrigger `json:"triggers" validate:"dive"`
	MaxRetries  int               `json:"max_retries,omitempty" validate:"min=0,max=10"`
	MaxRuntime  int               `json:"max_runtime,omitempty" validate:"min=1"`
	MinInterval int               `json:"min_interval,omitempty" validate:"min=1"`
}
