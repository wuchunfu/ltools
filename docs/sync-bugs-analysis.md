# 同步逻辑 Bug 分析和修复计划

## 发现的 Bug

### 🐛 Bug 1: 文件删除不同步（严重）
**位置**: `internal/sync/sync.go:210-255` (`copyFilesToSyncDir`)

**问题描述**:
- 当前只复制文件到同步目录
- **不删除**同步目录中已不存在于数据目录的文件
- 导致：用户删除数据目录的文件后，远程仓库仍保留这些文件

**影响**: 🔴 严重 - 数据不一致

**示例**:
```
1. 用户在数据目录创建 file1.txt
2. 同步到远程仓库
3. 用户删除 file1.txt
4. 再次同步
5. 结果：远程仓库仍然有 file1.txt ❌
```

**修复方案**:
```go
// 在 copyFilesToSyncDir 开始时：
// 1. 先删除同步目录中所有文件（除了 .git 目录）
// 2. 然后复制数据目录的文件
// 或者：
// 1. 复制文件
// 2. 删除同步目录中不存在于数据目录的文件
```

---

### 🐛 Bug 2: Pull 假设远程分支是 main（中等）
**位置**: `internal/sync/git.go:186` (`Pull`)

**问题描述**:
```go
_, err := g.runGit("pull", "origin", "main", "--rebase")
```
- 硬编码远程分支名为 `main`
- 如果远程分支是 `master` 或其他名称，会失败

**影响**: 🟡 中等 - 与某些仓库不兼容

**修复方案**:
```go
// 1. 先获取远程分支名称
// 2. 动态使用正确的分支名
```

---

### 🐛 Bug 3: Push 第一次失败（中等）
**位置**: `internal/sync/git.go:174-182` (`Push`)

**问题描述**:
```go
func (g *GitClient) Push(force bool) error {
    args := []string{"push", "-u", "origin", "main"}
    if force {
        args = []string{"push", "--force-with-lease", "-u", "origin", "main"}
    }
```

- 第一次推送时，远程分支不存在
- `--force-with-lease` 需要远程分支已存在才能工作
- 导致第一次推送失败

**影响**: 🟡 中等 - 首次同步失败

**修复方案**:
```go
// 1. 先尝试正常 push
// 2. 如果失败且有远程分支，再用 force-with-lease
// 或者：
// 1. 检查远程分支是否存在
// 2. 根据情况选择 push 策略
```

---

### 🐛 Bug 4: 自动同步 goroutine 没有错误处理（轻微）
**位置**: `internal/sync/sync.go:307-316` (`StartAutoSync`)

**问题描述**:
```go
go func() {
    for {
        select {
        case <-m.ticker.C:
            m.Sync()  // ❌ 忽略了返回结果
        case <-m.stopChan:
            return
        }
    }
}()
```

- `m.Sync()` 的错误被忽略
- 用户不知道自动同步失败

**影响**: 🟢 轻微 - 静默失败

**修复方案**:
```go
go func() {
    for {
        select {
        case <-m.ticker.C:
            result := m.Sync()
            if !result.Success {
                fmt.Printf("[SyncManager] Auto-sync failed: %s\n", result.Error)
                // 可以触发事件通知前端
            }
        case <-m.stopChan:
            return
        }
    }
}()
```

---

### 🐛 Bug 5: 文件复制没有同步模式（轻微）
**位置**: `internal/sync/sync.go:258-287` (`copyFile`)

**问题描述**:
- 复制文件时保留了源文件权限
- 但可能不是最优的（如可执行权限）

**影响**: 🟢 轻微 - 权限问题

---

## 修复优先级

| Bug | 优先级 | 影响 | 复杂度 |
|-----|--------|------|--------|
| Bug 1: 文件删除不同步 | 🔴 P0 | 数据不一致 | 中 |
| Bug 2: Pull 分支名硬编码 | 🟡 P1 | 兼容性 | 低 |
| Bug 3: Push 首次失败 | 🟡 P1 | 首次同步 | 中 |
| Bug 4: 自动同步无错误处理 | 🟢 P2 | 可观察性 | 低 |
| Bug 5: 文件权限 | 🟢 P3 | 边缘情况 | 低 |

## 建议修复顺序

1. **立即修复**: Bug 1（文件删除不同步）
2. **尽快修复**: Bug 2 + Bug 3（Pull/Push 问题）
3. **后续优化**: Bug 4 + Bug 5（错误处理和权限）

## 测试用例

### Bug 1 测试
```bash
1. 创建文件: echo "test" > ~/.ltools/test.txt
2. 执行同步: 应该同步 test.txt
3. 删除文件: rm ~/.ltools/test.txt
4. 执行同步: test.txt 应该从远程仓库删除
5. 检查远程: git ls-remote 应该不包含 test.txt
```

### Bug 2 测试
```bash
1. 使用 master 分支的仓库
2. 执行同步: 应该能正常 pull
```

### Bug 3 测试
```bash
1. 使用空仓库
2. 第一次同步: 应该成功
```
