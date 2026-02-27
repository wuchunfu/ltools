# Markdown 插件设计方案

**日期**: 2026-02-26
**状态**: 已确认，开始实现

## 概述

新增 Markdown 编辑器插件，提供编辑 + 实时预览 + 导出的完整方案。

## 技术选型

| 组件 | 技术方案 | 理由 |
|------|----------|------|
| 编辑器 | Monaco Editor | 复用现有依赖，语法高亮好 |
| 预览渲染 | react-markdown + 插件 | 轻量、React 原生、插件生态丰富 |
| 导出格式 | HTML + PDF | 满足大多数场景 |
| 界面布局 | 分屏模式 | 左侧编辑、右侧实时预览 |

## 前端依赖

```json
{
  "react-markdown": "^9.0.x",
  "remark-gfm": "^4.0.x",
  "remark-math": "^6.0.x",
  "rehype-katex": "^7.0.x",
  "rehype-highlight": "^7.0.x",
  "rehype-raw": "^7.0.x",
  "katex": "^0.16.x"
}
```

## 文件结构

```
plugins/markdown/
├── plugin.go              # 插件定义和服务注册
└── service.go             # 后端服务（文件导入/导出、PDF 生成）

frontend/src/components/
└── MarkdownWidget.tsx     # 主组件（分屏布局）
```

## 后端服务 API

```go
type MarkdownService struct {
    plugin *MarkdownPlugin
    app    *application.App
}

func (s *MarkdownService) ImportFile() (*FileResult, error)
func (s *MarkdownService) ExportHTML(content string, filename string) (string, error)
func (s *MarkdownService) ExportPDF(htmlContent string) error
```

## 插件元数据

```go
&plugins.PluginMetadata{
    ID:          "markdown.builtin",
    Name:        "Markdown 编辑器",
    Version:     "1.0.0",
    Description: "支持实时预览的 Markdown 编辑与导出工具",
    Keywords:    []string{"markdown", "md", "编辑器", "文档"},
    Permissions: []string{plugins.PermissionFileSystem},
    HasPage:     true,
    ShowInMenu:  true,
}
```

## 前端功能

### 工具栏
- 格式：加粗、斜体、删除线、行内代码
- 插入：链接、图片、代码块、表格
- 标题：H1-H6
- 列表：有序、无序、任务列表
- 操作：撤销、重做、清空

### 状态栏
- 字符数 / 字数 / 行数
- 光标位置
- 保存状态

### 预览特性
- 同步滚动
- 目录大纲（TOC）
- 代码块复制按钮
- 响应式布局

## 实现步骤

1. 创建后端插件骨架（plugin.go, service.go）
2. 安装前端依赖
3. 开发 MarkdownWidget.tsx 组件
4. 生成 Wails 绑定
5. 配置路由和导航
