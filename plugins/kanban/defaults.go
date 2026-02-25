package kanban

import "time"

// CreateDefaultBoard creates a default board with preset columns and labels
func CreateDefaultBoard() *Board {
	now := time.Now()
	return &Board{
		ID:          "default-board",
		Name:        "我的第一个看板",
		Description: "欢迎使用看板插件，开始管理你的任务吧！",
		Columns: []Column{
			{
				ID:       "col-todo",
				Name:     "待办",
				Position: 0,
				Cards: []Card{
					{
						ID:          "card-welcome",
						Title:       "欢迎使用看板",
						Description: "这是一个示例卡片。你可以：\n• 点击卡片查看详情\n• 拖拽卡片到不同列\n• 添加标签和子任务",
						Priority:    PriorityMedium,
						Labels:      []string{"label-feature"},
						Checklists: []ChecklistItem{
							{ID: "item-1", Text: "创建你的第一张卡片", Completed: false},
							{ID: "item-2", Text: "尝试拖拽卡片", Completed: false},
							{ID: "item-3", Text: "添加标签", Completed: false},
						},
						Position:  0,
						CreatedAt: now,
					},
				},
			},
			{
				ID:       "col-progress",
				Name:     "进行中",
				Position: 1,
				Cards:    []Card{},
			},
			{
				ID:       "col-done",
				Name:     "已完成",
				Position: 2,
				Cards:    []Card{},
			},
		},
		Labels: []Label{
			{ID: "label-bug", Name: "Bug", Color: "#EF4444"},
			{ID: "label-feature", Name: "功能", Color: "#3B82F6"},
			{ID: "label-opt", Name: "优化", Color: "#22C55E"},
			{ID: "label-docs", Name: "文档", Color: "#F59E0B"},
			{ID: "label-urgent", Name: "紧急", Color: "#DC2626"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
