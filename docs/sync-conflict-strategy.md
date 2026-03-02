# 同步文件冲突解决策略分析

## 当前实现

### Pull 阶段冲突处理

**位置**: `internal/sync/sync.go:96-106`

```go
// Pull remote changes first (if repo exists and has commits)
if m.git.IsRepo() {
    if _, err := m.git.GetCommitHash(); err == nil {
        if err := m.git.Pull(); err != nil {
            // Non-fatal: might be no remote commits yet
            fmt.Printf("[SyncManager] Pull warning (non-fatal): %v\n", err)
        }
    }
}
```

**Git Pull 实现** (`internal/sync/git.go:194-205`):
```go
func (g *GitClient) Pull() error {
    // Try to pull from main branch first
    _, err := g.runGit("pull", "origin", "main", "--rebase")
    if err == nil {
        return nil
    }

    // If main fails, try master
    _, err = g.runGit("pull", "origin", "master", "--rebase")
    return err
}
```

**策略**: `git pull --rebase`

---

### Push 阶段冲突处理

**位置**: `internal/sync/sync.go:148`

```go
// Push (force with lease for local-first strategy)
if err := m.git.Push(true); err != nil {
    result.Success = false
    result.Error = fmt.Sprintf("推送失败: %v", err)
    return result
}
```

**Git Push 实现** (`internal/sync/git.go:174-191`):
```go
func (g *GitClient) Push(force bool) error {
    // First, try normal push
    args := []string{"push", "-u", "origin", "main"}
    _, err := g.runGit(args...)
    if err == nil {
        return nil
    }

    // If normal push fails and force is requested
    if force {
        args = []string{"push", "--force-with-lease", "-u", "origin", "main"}
        _, err = g.runGit(args...)
        return err
    }

    return err
}
```

**策略**:
1. 先尝试正常 `push`
2. 失败后使用 `push --force-with-lease`

---

## 🔴 发现的严重问题

### 问题 1: Rebase 冲突未处理

**场景**:
```
1. 设备 A 修改 config.json，同步
2. 设备 B 修改 config.json（同一行），同步
3. 设备 B Pull → Rebase 冲突 ❌
```

**当前行为**:
- `git pull --rebase` 遇到冲突会**卡住**
- Git 进入 `rebase-apply` 状态
- 后续的 `git add`、`git commit` 都会失败
- 用户看到"推送失败"，但不知道原因

**影响**: 🔴 严重 - 同步完全失败，需要手动解决冲突

---

### 问题 2: Force-With-Lease 总是启用

**当前代码** (`internal/sync/sync.go:148`):
```go
if err := m.git.Push(true); err != nil {  // ← 总是 true
```

**问题**:
- 即使正常同步，也会在第二次尝试时强制推送
- 可能**覆盖其他设备的更改**
- `--force-with-lease` 虽然比 `--force` 安全，但仍会覆盖

**场景**:
```
1. 设备 A 推送 commit1
2. 设备 B 推送 commit2
3. 设备 A 再次同步，Pull 成功，Push 失败
4. 设备 A 自动使用 --force-with-lease
5. commit2 被 commit1' 覆盖 ❌
```

**影响**: 🟡 中等 - 数据丢失风险

---

### 问题 3: Pull 失败被忽略

**当前代码** (`internal/sync/sync.go:101-104`):
```go
if err := m.git.Pull(); err != nil {
    // Non-fatal: might be no remote commits yet
    fmt.Printf("[SyncManager] Pull warning (non-fatal): %v\n", err)
}
```

**问题**:
- Pull 失败（包括冲突）被视为非致命
- 继续执行后续操作
- 可能导致数据不一致

**影响**: 🟡 中等 - 静默失败

---

## ✅ 正确的冲突解决策略

### 策略 1: 本地优先（Local-First）

**适用场景**: 个人工具，单用户多设备

**流程**:
```
1. Pull 远程更改（rebase）
2. 如果有冲突：
   - 自动使用本地版本（ours）
   - 或自动使用远程版本（theirs）
3. Push 到远程（force-with-lease 作为备选）
```

**优点**:
- ✅ 简单，自动化
- ✅ 不会卡住

**缺点**:
- ❌ 可能丢失数据
- ❌ 不适合协作

---

### 策略 2: 三路合并（Merge）

**适用场景**: 团队协作

**流程**:
```
1. Pull 远程更改（merge）
2. 如果有冲突：
   - 创建合并提交
   - 或通知用户手动解决
3. Push 到远程（不用 force）
```

**优点**:
- ✅ 不丢失数据
- ✅ 保留历史

**缺点**:
- ❌ 可能需要手动解决冲突
- ❌ 历史复杂

---

### 策略 3: 版本控制（推荐）

**适用场景**: 配置文件同步

**流程**:
```
1. 为冲突文件创建备份
2. Pull 远程更改
3. 如果有冲突：
   - 保存本地版本为 file.local.json
   - 保存远程版本为 file.remote.json
   - 保存合并版本为 file.json
4. 通知用户
5. Push 到远程
```

**优点**:
- ✅ 不丢失数据
- ✅ 用户可以选择
- ✅ 自动化

**缺点**:
- ❌ 实现复杂

---

## 🛠️ 推荐修复方案

### 方案 A: 快速修复（本地优先）

**适合**: 当前单用户场景

**修改**:
```go
// Pull with conflict auto-resolution
func (g *GitClient) Pull() error {
    // Try rebase with auto-resolution
    _, err := g.runGit("pull", "origin", "main", "--rebase", "-X", "ours")
    if err != nil {
        // If rebase fails, abort and try merge
        g.runGit("rebase", "--abort")
        _, err = g.runGit("pull", "origin", "main", "--no-rebase")
    }
    return err
}

// Push without force by default
func (m *SyncManager) Sync() *SyncResult {
    // ... 前面的代码 ...

    // Push (only use force if really needed)
    if err := m.git.Push(false); err != nil {
        // Normal push failed, maybe remote has new commits
        // Try pull again
        if pullErr := m.git.Pull(); pullErr != nil {
            result.Error = fmt.Sprintf("同步冲突，请手动解决: %v", pullErr)
            return result
        }
        // Try push again
        if err := m.git.Push(false); err != nil {
            result.Error = fmt.Sprintf("推送失败: %v", err)
            return result
        }
    }
}
```

---

### 方案 B: 完整修复（版本控制）

**适合**: 生产环境

**修改**:
```go
func (m *SyncManager) handleConflicts() error {
    // 1. 检查是否有冲突文件
    conflicts, _ := m.git.GetConflicts()

    for _, file := range conflicts {
        // 2. 创建备份
        localPath := filepath.Join(m.dataDir, file)
        backupPath := localPath + ".conflict." + time.Now().Format("20060102-150405")

        // 3. 复制本地版本
        copyFile(localPath, backupPath)

        // 4. 使用远程版本
        m.git.CheckoutTheirs(file)

        // 5. 记录冲突
        m.config.AddConflict(file, backupPath)
    }

    // 6. 通知前端
    if len(conflicts) > 0 {
        m.emitEvent("conflicts-detected", conflicts)
    }

    return nil
}
```

---

## 📋 当前策略总结

| 策略 | 当前实现 | 问题 |
|------|---------|------|
| **Pull 冲突** | `--rebase` | ❌ 冲突会卡住 |
| **Push 冲突** | `--force-with-lease` | ⚠️ 可能覆盖远程 |
| **冲突检测** | 无 | ❌ 静默失败 |
| **用户通知** | 无 | ❌ 用户不知道 |

---

## 🎯 建议优先级

1. **立即修复**: 添加 rebase 冲突自动解决（`-X ours`）
2. **尽快修复**: 只在必要时使用 force
3. **后续优化**: 添加冲突检测和通知
4. **长期改进**: 实现版本控制策略

---

## 测试用例

### 测试 1: Rebase 冲突
```bash
# 设备 A
echo "A" > ~/.ltools/test.txt
# 同步

# 设备 B
echo "B" > ~/.ltools/test.txt
# 同步 → 应该自动解决冲突，不卡住
```

### 测试 2: 远程有新提交
```bash
# 设备 A 推送
# 设备 B 推送
# 设备 A 再次推送 → 应该先 pull 再 push，不用 force
```
