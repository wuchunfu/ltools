package kanban

import "time"

// Priority represents card priority level
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// Board represents a kanban board
type Board struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Columns     []Column  `json:"columns"`
	Labels      []Label   `json:"labels"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Column represents a column in a board
type Column struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	Cards    []Card `json:"cards"`
}

// Card represents a card in a column
type Card struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Priority    Priority        `json:"priority"`
	Labels      []string        `json:"labels"` // Label IDs
	DueDate     *time.Time      `json:"dueDate"`
	Checklists  []ChecklistItem `json:"checklists"`
	Position    int             `json:"position"`
	CreatedAt   time.Time       `json:"createdAt"`
	CompletedAt *time.Time      `json:"completedAt"`
}

// Label represents a label/tag for cards
type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"` // hex color
}

// ChecklistItem represents a checklist item in a card
type ChecklistItem struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

// KanbanConfig represents the persistent configuration
type KanbanConfig struct {
	Version int     `json:"version"`
	Boards  []Board `json:"boards"`
}

// CardUpdate represents fields that can be updated on a card
type CardUpdate struct {
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Priority    *Priority       `json:"priority,omitempty"`
	Labels      *[]string       `json:"labels,omitempty"`
	DueDate     *time.Time      `json:"dueDate,omitempty"`
}
