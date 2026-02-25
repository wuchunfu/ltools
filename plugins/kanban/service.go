package kanban

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// KanbanService exposes kanban management functionality to frontend
type KanbanService struct {
	plugin  *KanbanPlugin
	app     *application.App
	dataDir string
}

// NewKanbanService creates a new kanban service
func NewKanbanService(plugin *KanbanPlugin, app *application.App, dataDir string) *KanbanService {
	return &KanbanService{
		plugin:  plugin,
		app:     app,
		dataDir: dataDir,
	}
}

// SetDataDir sets the data directory
func (s *KanbanService) SetDataDir(dataDir string) {
	s.dataDir = dataDir
}

// emitEvent emits a kanban event
func (s *KanbanService) emitEvent(eventType string, data string) {
	if s.app != nil {
		s.app.Event.Emit("kanban:"+eventType, data)
	}
}

// ============================================================================
// Board Management API
// ============================================================================

// GetBoards returns all boards
func (s *KanbanService) GetBoards() []Board {
	if s.plugin.config == nil {
		return []Board{}
	}
	return s.plugin.config.Boards
}

// GetBoard returns a specific board by ID
func (s *KanbanService) GetBoard(id string) *Board {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == id {
			return &s.plugin.config.Boards[i]
		}
	}
	return nil
}

// CreateBoardResult represents the result of creating a board
type CreateBoardResult struct {
	Board *Board `json:"board"`
	Error string `json:"error,omitempty"`
}

// CreateBoard creates a new board
func (s *KanbanService) CreateBoard(name, description string) (*CreateBoardResult, error) {
	if name == "" {
		return &CreateBoardResult{Error: "看板名称不能为空"}, nil
	}

	now := time.Now()
	board := &Board{
		ID:          generateID(name),
		Name:        name,
		Description: description,
		Columns: []Column{
			{ID: generateID("col-todo"), Name: "待办", Position: 0, Cards: []Card{}},
			{ID: generateID("col-progress"), Name: "进行中", Position: 1, Cards: []Card{}},
			{ID: generateID("col-done"), Name: "已完成", Position: 2, Cards: []Card{}},
		},
		Labels: []Label{
			{ID: generateID("label-bug"), Name: "Bug", Color: "#EF4444"},
			{ID: generateID("label-feature"), Name: "功能", Color: "#3B82F6"},
			{ID: generateID("label-opt"), Name: "优化", Color: "#22C55E"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.plugin.config.Boards = append(s.plugin.config.Boards, *board)

	if err := s.plugin.SaveConfig(s.dataDir); err != nil {
		return &CreateBoardResult{Error: fmt.Sprintf("保存配置失败: %v", err)}, nil
	}

	s.emitEvent("board:created", board.ID)
	return &CreateBoardResult{Board: board}, nil
}

// UpdateBoard updates an existing board
func (s *KanbanService) UpdateBoard(id, name, description string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == id {
			s.plugin.config.Boards[i].Name = name
			s.plugin.config.Boards[i].Description = description
			s.plugin.config.Boards[i].UpdatedAt = time.Now()

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.emitEvent("board:updated", id)
			return nil
		}
	}
	return fmt.Errorf("看板未找到: %s", id)
}

// DeleteBoard deletes a board
func (s *KanbanService) DeleteBoard(id string) error {
	for i, board := range s.plugin.config.Boards {
		if board.ID == id {
			s.plugin.config.Boards = append(
				s.plugin.config.Boards[:i],
				s.plugin.config.Boards[i+1:]...,
			)

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.emitEvent("board:deleted", id)
			return nil
		}
	}
	return fmt.Errorf("看板未找到: %s", id)
}

// ============================================================================
// Column Management API
// ============================================================================

// CreateColumnResult represents the result of creating a column
type CreateColumnResult struct {
	Column *Column `json:"column"`
	Error  string  `json:"error,omitempty"`
}

// CreateColumn creates a new column in a board
func (s *KanbanService) CreateColumn(boardID, name string) (*CreateColumnResult, error) {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			position := len(s.plugin.config.Boards[i].Columns)
			column := &Column{
				ID:       generateID("col"),
				Name:     name,
				Position: position,
				Cards:    []Card{},
			}

			s.plugin.config.Boards[i].Columns = append(s.plugin.config.Boards[i].Columns, *column)
			s.plugin.config.Boards[i].UpdatedAt = time.Now()

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return &CreateColumnResult{Error: fmt.Sprintf("保存配置失败: %v", err)}, nil
			}

			s.emitEvent("column:created", boardID)
			return &CreateColumnResult{Column: column}, nil
		}
	}
	return &CreateColumnResult{Error: "看板未找到"}, nil
}

// UpdateColumn updates a column's name
func (s *KanbanService) UpdateColumn(boardID, columnID, name string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					s.plugin.config.Boards[i].Columns[j].Name = name
					s.plugin.config.Boards[i].UpdatedAt = time.Now()

					if err := s.plugin.SaveConfig(s.dataDir); err != nil {
						return fmt.Errorf("保存配置失败: %w", err)
					}

					s.emitEvent("column:updated", boardID)
					return nil
				}
			}
			return fmt.Errorf("列未找到: %s", columnID)
		}
	}
	return fmt.Errorf("看板未找到: %s", boardID)
}

// DeleteColumn deletes a column from a board
func (s *KanbanService) DeleteColumn(boardID, columnID string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j, col := range s.plugin.config.Boards[i].Columns {
				if col.ID == columnID {
					// Delete column
					s.plugin.config.Boards[i].Columns = append(
						s.plugin.config.Boards[i].Columns[:j],
						s.plugin.config.Boards[i].Columns[j+1:]...,
					)
					// Re-position remaining columns
					for k := range s.plugin.config.Boards[i].Columns {
						s.plugin.config.Boards[i].Columns[k].Position = k
					}
					s.plugin.config.Boards[i].UpdatedAt = time.Now()

					if err := s.plugin.SaveConfig(s.dataDir); err != nil {
						return fmt.Errorf("保存配置失败: %w", err)
					}

					s.emitEvent("column:deleted", boardID)
					return nil
				}
			}
			return fmt.Errorf("列未找到: %s", columnID)
		}
	}
	return fmt.Errorf("看板未找到: %s", boardID)
}

// MoveColumn moves a column to a new position
func (s *KanbanService) MoveColumn(boardID, columnID string, newPosition int) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			columns := &s.plugin.config.Boards[i].Columns
			var colIndex int = -1

			for j, col := range *columns {
				if col.ID == columnID {
					colIndex = j
					break
				}
			}

			if colIndex == -1 {
				return fmt.Errorf("列未找到: %s", columnID)
			}

			// Clamp newPosition
			if newPosition < 0 {
				newPosition = 0
			}
			if newPosition >= len(*columns) {
				newPosition = len(*columns) - 1
			}

			// Remove column from current position
			col := (*columns)[colIndex]
			*columns = append((*columns)[:colIndex], (*columns)[colIndex+1:]...)

			// Insert at new position
			*columns = append((*columns)[:newPosition], append([]Column{col}, (*columns)[newPosition:]...)...)

			// Update positions
			for k := range *columns {
				(*columns)[k].Position = k
			}

			s.plugin.config.Boards[i].UpdatedAt = time.Now()
			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.emitEvent("column:moved", boardID)
			return nil
		}
	}
	return fmt.Errorf("看板未找到: %s", boardID)
}

// ============================================================================
// Card Management API
// ============================================================================

// CreateCardResult represents the result of creating a card
type CreateCardResult struct {
	Card  *Card  `json:"card"`
	Error string `json:"error,omitempty"`
}

// CreateCard creates a new card in a column
func (s *KanbanService) CreateCard(boardID, columnID, title string, priority ...Priority) (*CreateCardResult, error) {
	if title == "" {
		return &CreateCardResult{Error: "卡片标题不能为空"}, nil
	}

	// Default priority
	cardPriority := PriorityMedium
	if len(priority) > 0 {
		cardPriority = priority[0]
	}

	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					now := time.Now()
					position := len(s.plugin.config.Boards[i].Columns[j].Cards)
					card := &Card{
						ID:          generateID("card"),
						Title:       title,
						Description: "",
						Priority:    cardPriority,
						Labels:      []string{},
						Checklists:  []ChecklistItem{},
						Position:    position,
						CreatedAt:   now,
					}

					s.plugin.config.Boards[i].Columns[j].Cards = append(
						s.plugin.config.Boards[i].Columns[j].Cards,
						*card,
					)
					s.plugin.config.Boards[i].UpdatedAt = time.Now()

					if err := s.plugin.SaveConfig(s.dataDir); err != nil {
						return &CreateCardResult{Error: fmt.Sprintf("保存配置失败: %v", err)}, nil
					}

					s.emitEvent("card:created", boardID)
					return &CreateCardResult{Card: card}, nil
				}
			}
			return &CreateCardResult{Error: "列未找到"}, nil
		}
	}
	return &CreateCardResult{Error: "看板未找到"}, nil
}

// UpdateCard updates a card's properties
func (s *KanbanService) UpdateCard(boardID, columnID, cardID string, updates CardUpdate) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					for k := range s.plugin.config.Boards[i].Columns[j].Cards {
						if s.plugin.config.Boards[i].Columns[j].Cards[k].ID == cardID {
							card := &s.plugin.config.Boards[i].Columns[j].Cards[k]

							if updates.Title != nil {
								card.Title = *updates.Title
							}
							if updates.Description != nil {
								card.Description = *updates.Description
							}
							if updates.Priority != nil {
								card.Priority = *updates.Priority
							}
							if updates.Labels != nil {
								card.Labels = *updates.Labels
							}
							if updates.DueDate != nil {
								card.DueDate = updates.DueDate
							}

							s.plugin.config.Boards[i].UpdatedAt = time.Now()

							if err := s.plugin.SaveConfig(s.dataDir); err != nil {
								return fmt.Errorf("保存配置失败: %w", err)
							}

							s.emitEvent("card:updated", boardID)
							return nil
						}
					}
					return fmt.Errorf("卡片未找到: %s", cardID)
				}
			}
			return fmt.Errorf("列未找到: %s", columnID)
		}
	}
	return fmt.Errorf("看板未找到: %s", boardID)
}

// DeleteCard deletes a card from a column
func (s *KanbanService) DeleteCard(boardID, columnID, cardID string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					for k, card := range s.plugin.config.Boards[i].Columns[j].Cards {
						if card.ID == cardID {
							// Delete card
							s.plugin.config.Boards[i].Columns[j].Cards = append(
								s.plugin.config.Boards[i].Columns[j].Cards[:k],
								s.plugin.config.Boards[i].Columns[j].Cards[k+1:]...,
							)
							// Re-position remaining cards
							for m := range s.plugin.config.Boards[i].Columns[j].Cards {
								s.plugin.config.Boards[i].Columns[j].Cards[m].Position = m
							}
							s.plugin.config.Boards[i].UpdatedAt = time.Now()

							if err := s.plugin.SaveConfig(s.dataDir); err != nil {
								return fmt.Errorf("保存配置失败: %w", err)
							}

							s.emitEvent("card:deleted", boardID)
							return nil
						}
					}
					return fmt.Errorf("卡片未找到: %s", cardID)
				}
			}
			return fmt.Errorf("列未找到: %s", columnID)
		}
	}
	return fmt.Errorf("看板未找到: %s", boardID)
}

// MoveCard moves a card to a different column or position
func (s *KanbanService) MoveCard(boardID, fromColumnID, toColumnID, cardID string, newPosition int) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			var card *Card
			var fromColIdx, cardIdx int = -1, -1

			// Find the card and its current column
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == fromColumnID {
					fromColIdx = j
					for k, c := range s.plugin.config.Boards[i].Columns[j].Cards {
						if c.ID == cardID {
							cardIdx = k
							cardCopy := c
							card = &cardCopy
							break
						}
					}
					break
				}
			}

			if card == nil {
				return fmt.Errorf("卡片未找到: %s", cardID)
			}

			// Find target column
			var toColIdx int = -1
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == toColumnID {
					toColIdx = j
					break
				}
			}

			if toColIdx == -1 {
				return fmt.Errorf("目标列未找到: %s", toColumnID)
			}

			// Remove card from source column
			s.plugin.config.Boards[i].Columns[fromColIdx].Cards = append(
				s.plugin.config.Boards[i].Columns[fromColIdx].Cards[:cardIdx],
				s.plugin.config.Boards[i].Columns[fromColIdx].Cards[cardIdx+1:]...,
			)

			// Re-position cards in source column
			for m := range s.plugin.config.Boards[i].Columns[fromColIdx].Cards {
				s.plugin.config.Boards[i].Columns[fromColIdx].Cards[m].Position = m
			}

			// Clamp newPosition
			targetCards := &s.plugin.config.Boards[i].Columns[toColIdx].Cards
			if newPosition < 0 {
				newPosition = 0
			}
			if newPosition > len(*targetCards) {
				newPosition = len(*targetCards)
			}

			// Insert card at new position
			card.Position = newPosition
			*targetCards = append((*targetCards)[:newPosition], append([]Card{*card}, (*targetCards)[newPosition:]...)...)

			// Re-position cards in target column
			for m := range *targetCards {
				(*targetCards)[m].Position = m
			}

			s.plugin.config.Boards[i].UpdatedAt = time.Now()
			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.emitEvent("card:moved", boardID)
			return nil
		}
	}
	return fmt.Errorf("看板未找到: %s", boardID)
}

// CompleteCard marks a card as completed
func (s *KanbanService) CompleteCard(boardID, columnID, cardID string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					for k := range s.plugin.config.Boards[i].Columns[j].Cards {
						if s.plugin.config.Boards[i].Columns[j].Cards[k].ID == cardID {
							now := time.Now()
							s.plugin.config.Boards[i].Columns[j].Cards[k].CompletedAt = &now
							s.plugin.config.Boards[i].UpdatedAt = time.Now()

							if err := s.plugin.SaveConfig(s.dataDir); err != nil {
								return fmt.Errorf("保存配置失败: %w", err)
							}

							s.emitEvent("card:completed", boardID)
							return nil
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("卡片未找到: %s", cardID)
}

// UncompleteCard removes completed status from a card
func (s *KanbanService) UncompleteCard(boardID, columnID, cardID string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					for k := range s.plugin.config.Boards[i].Columns[j].Cards {
						if s.plugin.config.Boards[i].Columns[j].Cards[k].ID == cardID {
							s.plugin.config.Boards[i].Columns[j].Cards[k].CompletedAt = nil
							s.plugin.config.Boards[i].UpdatedAt = time.Now()

							if err := s.plugin.SaveConfig(s.dataDir); err != nil {
								return fmt.Errorf("保存配置失败: %w", err)
							}

							s.emitEvent("card:uncompleted", boardID)
							return nil
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("卡片未找到: %s", cardID)
}

// ============================================================================
// Label Management API
// ============================================================================

// CreateLabelResult represents the result of creating a label
type CreateLabelResult struct {
	Label *Label `json:"label"`
	Error string `json:"error,omitempty"`
}

// CreateLabel creates a new label in a board
func (s *KanbanService) CreateLabel(boardID, name, color string) (*CreateLabelResult, error) {
	if name == "" {
		return &CreateLabelResult{Error: "标签名称不能为空"}, nil
	}

	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			label := &Label{
				ID:    generateID("label"),
				Name:  name,
				Color: color,
			}

			s.plugin.config.Boards[i].Labels = append(s.plugin.config.Boards[i].Labels, *label)
			s.plugin.config.Boards[i].UpdatedAt = time.Now()

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return &CreateLabelResult{Error: fmt.Sprintf("保存配置失败: %v", err)}, nil
			}

			s.emitEvent("label:created", boardID)
			return &CreateLabelResult{Label: label}, nil
		}
	}
	return &CreateLabelResult{Error: "看板未找到"}, nil
}

// UpdateLabel updates a label's name and color
func (s *KanbanService) UpdateLabel(boardID, labelID, name, color string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Labels {
				if s.plugin.config.Boards[i].Labels[j].ID == labelID {
					s.plugin.config.Boards[i].Labels[j].Name = name
					s.plugin.config.Boards[i].Labels[j].Color = color
					s.plugin.config.Boards[i].UpdatedAt = time.Now()

					if err := s.plugin.SaveConfig(s.dataDir); err != nil {
						return fmt.Errorf("保存配置失败: %w", err)
					}

					s.emitEvent("label:updated", boardID)
					return nil
				}
			}
			return fmt.Errorf("标签未找到: %s", labelID)
		}
	}
	return fmt.Errorf("看板未找到: %s", boardID)
}

// DeleteLabel deletes a label from a board
func (s *KanbanService) DeleteLabel(boardID, labelID string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j, label := range s.plugin.config.Boards[i].Labels {
				if label.ID == labelID {
					// Delete label
					s.plugin.config.Boards[i].Labels = append(
						s.plugin.config.Boards[i].Labels[:j],
						s.plugin.config.Boards[i].Labels[j+1:]...,
					)

					// Remove label from all cards
					for k := range s.plugin.config.Boards[i].Columns {
						for m := range s.plugin.config.Boards[i].Columns[k].Cards {
							newLabels := []string{}
							for _, l := range s.plugin.config.Boards[i].Columns[k].Cards[m].Labels {
								if l != labelID {
									newLabels = append(newLabels, l)
								}
							}
							s.plugin.config.Boards[i].Columns[k].Cards[m].Labels = newLabels
						}
					}

					s.plugin.config.Boards[i].UpdatedAt = time.Now()

					if err := s.plugin.SaveConfig(s.dataDir); err != nil {
						return fmt.Errorf("保存配置失败: %w", err)
					}

					s.emitEvent("label:deleted", boardID)
					return nil
				}
			}
			return fmt.Errorf("标签未找到: %s", labelID)
		}
	}
	return fmt.Errorf("看板未找到: %s", boardID)
}

// ============================================================================
// Checklist Management API
// ============================================================================

// AddChecklistItem adds a checklist item to a card
func (s *KanbanService) AddChecklistItem(boardID, columnID, cardID, text string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					for k := range s.plugin.config.Boards[i].Columns[j].Cards {
						if s.plugin.config.Boards[i].Columns[j].Cards[k].ID == cardID {
							item := ChecklistItem{
								ID:        generateID("item"),
								Text:      text,
								Completed: false,
							}

							s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists = append(
								s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists,
								item,
							)
							s.plugin.config.Boards[i].UpdatedAt = time.Now()

							if err := s.plugin.SaveConfig(s.dataDir); err != nil {
								return fmt.Errorf("保存配置失败: %w", err)
							}

							s.emitEvent("checklist:added", boardID)
							return nil
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("卡片未找到: %s", cardID)
}

// ToggleChecklistItem toggles a checklist item's completed state
func (s *KanbanService) ToggleChecklistItem(boardID, columnID, cardID, itemID string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					for k := range s.plugin.config.Boards[i].Columns[j].Cards {
						if s.plugin.config.Boards[i].Columns[j].Cards[k].ID == cardID {
							for m := range s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists {
								if s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists[m].ID == itemID {
									s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists[m].Completed =
										!s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists[m].Completed
									s.plugin.config.Boards[i].UpdatedAt = time.Now()

									if err := s.plugin.SaveConfig(s.dataDir); err != nil {
										return fmt.Errorf("保存配置失败: %w", err)
									}

									s.emitEvent("checklist:toggled", boardID)
									return nil
								}
							}
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("检查项未找到: %s", itemID)
}

// RemoveChecklistItem removes a checklist item from a card
func (s *KanbanService) RemoveChecklistItem(boardID, columnID, cardID, itemID string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					for k := range s.plugin.config.Boards[i].Columns[j].Cards {
						if s.plugin.config.Boards[i].Columns[j].Cards[k].ID == cardID {
							for m, item := range s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists {
								if item.ID == itemID {
									s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists = append(
										s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists[:m],
										s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists[m+1:]...,
									)
									s.plugin.config.Boards[i].UpdatedAt = time.Now()

									if err := s.plugin.SaveConfig(s.dataDir); err != nil {
										return fmt.Errorf("保存配置失败: %w", err)
									}

									s.emitEvent("checklist:removed", boardID)
									return nil
								}
							}
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("检查项未找到: %s", itemID)
}

// UpdateChecklistItem updates a checklist item's text
func (s *KanbanService) UpdateChecklistItem(boardID, columnID, cardID, itemID, text string) error {
	for i := range s.plugin.config.Boards {
		if s.plugin.config.Boards[i].ID == boardID {
			for j := range s.plugin.config.Boards[i].Columns {
				if s.plugin.config.Boards[i].Columns[j].ID == columnID {
					for k := range s.plugin.config.Boards[i].Columns[j].Cards {
						if s.plugin.config.Boards[i].Columns[j].Cards[k].ID == cardID {
							for m := range s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists {
								if s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists[m].ID == itemID {
									s.plugin.config.Boards[i].Columns[j].Cards[k].Checklists[m].Text = text
									s.plugin.config.Boards[i].UpdatedAt = time.Now()

									if err := s.plugin.SaveConfig(s.dataDir); err != nil {
										return fmt.Errorf("保存配置失败: %w", err)
									}

									s.emitEvent("checklist:updated", boardID)
									return nil
								}
							}
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("检查项未找到: %s", itemID)
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateID generates a unique ID
func generateID(prefix string) string {
	normalized := strings.ToLower(strings.TrimSpace(prefix))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")

	result := strings.Builder{}
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	id := result.String()
	if id == "" {
		id = "id"
	}

	hash := md5.Sum([]byte(prefix + time.Now().Format(time.RFC3339Nano)))
	hashSuffix := hex.EncodeToString(hash[:])[:8]

	return fmt.Sprintf("%s-%s", id, hashSuffix)
}
