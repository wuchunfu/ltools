# 发布流程

本文档说明如何发布 LTools 新版本。

## 发布前准备

### 1. 版本号规划

使用语义化版本（SemVer）：

```
MAJOR.MINOR.PATCH

- MAJOR: 不兼容的 API 修改
- MINOR: 向后兼容的功能新增
- PATCH: 向后兼容的问题修复
```

**示例**：
- `0.1.0` → `0.1.1`：Bug 修复
- `0.1.1` → `0.2.0`：新功能
- `0.2.0` → `1.0.0`：重大更新

### 2. 更新日志

准备更新日志（RELEASE.md）：

```markdown
# v1.2.0

## 新功能
- 添加二维码识别功能
- 支持自定义主题

## 改进
- 优化搜索性能
- 改进内存使用

## 修复
- 修复剪贴板历史显示问题
- 修复快捷键冲突

## 其他
- 更新依赖版本
- 改进文档
```

### 3. 版本更新

更新版本号：

**build/config.yml**：
```yaml
info:
  version: "1.2.0"  # 更新此处
```

**package.json**（如有）：
```json
{
  "version": "1.2.0"
}
```

### 4. 测试验证

发布前必须测试：

**功能测试**：
- [ ] 所有插件正常工作
- [ ] 快捷键正常
- [ ] 搜索功能正常
- [ ] 自动更新正常

**平台测试**：
- [ ] macOS (ARM64)
- [ ] macOS (Intel)
- [ ] Windows
- [ ] Linux

**构建测试**：
```bash
# 测试构建
task build

# 运行应用测试
./bin/ltools
```

## 发布步骤

### 1. 创建发布分支

```bash
# 从 main 创建发布分支
git checkout main
git pull origin main
git checkout -b release/v1.2.0
```

### 2. 更新版本和文档

```bash
# 更新版本号
vim build/config.yml

# 更新更新日志
vim RELEASE.md

# 提交更改
git add .
git commit -m "chore: 准备发布 v1.2.0"
```

### 3. 合并到 main

```bash
# 推送发布分支
git push origin release/v1.2.0

# 创建 Pull Request
# 等待审查和合并
```

### 4. 创建 Git Tag

合并后创建标签：

```bash
# 切换到 main
git checkout main
git pull origin main

# 创建标签
git tag -a v1.2.0 -m "Release v1.2.0"

# 推送标签
git push origin v1.2.0
```

### 5. GitHub Actions 自动构建

推送标签后，GitHub Actions 自动：

1. 构建所有平台
2. 创建 GitHub Release
3. 上传安装包
4. 生成更新清单

**监控构建**：
- 访问 Actions 页面
- 查看构建进度
- 检查是否成功

### 6. 验证发布

**检查 Release**：
1. 访问 [Releases](https://github.com/lian-yang/ltools/releases)
2. 确认 v1.2.0 已创建
3. 检查所有文件已上传

**测试下载**：
```bash
# 下载测试
curl -LO https://github.com/lian-yang/ltools/releases/download/v1.2.0/ltools-v1.2.0-darwin-arm64.tar.gz

# 验证文件
shasum -a 256 ltools-v1.2.0-darwin-arm64.tar.gz
```

**测试自动更新**：
1. 安装新版本
2. 检查更新功能
3. 验证更新清单

## 发布后工作

### 1. 更新官网

```bash
# 更新下载页面
vim website/docs/download.md

# 更新更新日志
vim website/docs/changelog.md

# 提交并推送
git add .
git commit -m "docs: 更新 v1.2.0 文档"
git push origin main
```

### 2. 公告发布

**GitHub Release**：
- 编辑 Release Notes
- 添加截图和 GIF
- 说明重要变更

**社交媒体**（如有）：
- Twitter
- 微博
- 等等

### 3. 关闭里程碑

在 GitHub：
1. 关闭相关 Issues
2. 更新 Milestone
3. 归档项目看板

### 4. 收集反馈

发布后：
- 监控 Issue 反馈
- 收集用户意见
- 规划下个版本

## 热修复流程

紧急问题的快速修复：

### 1. 创建热修复分支

```bash
# 基于 tag 创建
git checkout -b hotfix/v1.2.1 v1.2.0
```

### 2. 修复问题

```bash
# 修复代码
git add .
git commit -m "fix: 修复紧急问题"
```

### 3. 快速发布

```bash
# 更新版本号
vim build/config.yml  # 1.2.1

# 创建标签
git tag -a v1.2.1 -m "Hotfix v1.2.1"
git push origin v1.2.1
```

## 回滚流程

如果发布出现严重问题：

### 1. 标记问题版本

```bash
# 删除远程标签
git push --delete origin v1.2.0

# 删除本地标签
git tag -d v1.2.0
```

### 2. 发布修复版本

```bash
# 修复后重新发布
git tag -a v1.2.1 -m "Hotfix v1.2.1"
git push origin v1.2.1
```

## 发布检查清单

### 发布前

- [ ] 版本号已更新
- [ ] 更新日志已准备
- [ ] 所有测试通过
- [ ] 多平台测试完成
- [ ] 文档已更新

### 发布中

- [ ] Git 标签已创建
- [ ] GitHub Actions 成功
- [ ] Release 已创建
- [ ] 文件已上传

### 发布后

- [ ] 下载测试通过
- [ ] 自动更新测试
- [ ] 官网已更新
- [ ] 公告已发布
- [ ] 反馈已收集

## 常见问题

### Q: 构建失败怎么办？

**A**:
1. 查看 Actions 日志
2. 修复构建错误
3. 删除失败标签
4. 重新创建标签

### Q: 如何撤销发布？

**A**:
1. 删除 GitHub Release
2. 删除 Git 标签
3. 修复问题后重新发布

### Q: 自动更新不工作？

**A**:
1. 检查 update.json
2. 验证下载链接
3. 测试更新流程

## 参考资源

- [语义化版本](https://semver.org/)
- [GitHub Actions](https://docs.github.com/en/actions)
- [发布最佳实践](https://docs.github.com/en/repositories/releasing-projects-on-github)
