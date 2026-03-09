# 如何贡献

感谢你对 LTools 的关注！我们欢迎所有形式的贡献。

## 贡献方式

### 🐛 报告 Bug

发现问题？帮助我们改进！

1. **搜索现有 Issue**
   - 访问 [GitHub Issues](https://github.com/lian-yang/ltools/issues)
   - 搜索是否已有相同问题

2. **创建新 Issue**
   - 点击「New Issue」
   - 选择 Bug Report 模板
   - 详细描述问题

3. **提供信息**
   - 操作系统和版本
   - LTools 版本
   - 复现步骤
   - 期望行为
   - 实际行为
   - 截图或日志

### 💡 提出新功能

有好的想法？告诉我们！

1. **搜索现有建议**
   - 检查是否已有类似建议

2. **创建 Feature Request**
   - 选择 Feature Request 模板
   - 描述功能需求
   - 说明使用场景
   - 提供示例（如有）

### 📝 改进文档

文档需要改进？直接修改！

1. **Fork 仓库**
   ```bash
   git clone https://github.com/YOUR_USERNAME/ltools.git
   ```

2. **修改文档**
   - 文档位置：`website/docs/`
   - 使用 Markdown 格式

3. **本地预览**
   ```bash
   cd website
   npm install
   npm run docs:dev
   ```

4. **提交 PR**
   - 创建新分支
   - 提交修改
   - 创建 Pull Request

### 🔧 贡献代码

想要贡献代码？太棒了！

1. **Fork 并 Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/ltools.git
   cd ltools
   ```

2. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **开发环境**
   ```bash
   # 安装依赖
   go mod download
   cd frontend && npm install && cd ..

   # 启动开发模式
   task dev
   ```

4. **编写代码**
   - 遵循代码规范
   - 添加测试
   - 更新文档

5. **提交代码**
   ```bash
   git add .
   git commit -m "feat: 添加某某功能"
   git push origin feature/your-feature-name
   ```

6. **创建 Pull Request**
   - 访问你 fork 的仓库
   - 点击「New Pull Request」
   - 填写 PR 模板

## 开发流程

### 1. 选择任务

- 查看 [Good First Issue](https://github.com/lian-yang/ltools/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
- 查看 [Help Wanted](https://github.com/lian-yang/ltools/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
- 或提出自己的改进建议

### 2. 讨论方案

在开始编码前：
- 在 Issue 中讨论实现方案
- 与维护者确认设计方向
- 避免重复工作

### 3. 开发实现

- 编写清晰、可维护的代码
- 遵循项目代码规范
- 添加必要的注释
- 编写单元测试

### 4. 测试验证

- 确保所有测试通过
- 手动测试新功能
- 检查边界情况
- 测试多平台兼容性

### 5. 代码审查

- 响应审查意见
- 及时修改问题
- 保持讨论专业和友好

## Pull Request 指南

### PR 标题格式

使用约定式提交格式：

- `feat:` 新功能
- `fix:` Bug 修复
- `docs:` 文档更新
- `style:` 代码格式
- `refactor:` 重构
- `test:` 测试
- `chore:` 构建/工具

示例：
```
feat: 添加二维码识别功能
fix: 修复剪贴板历史显示问题
docs: 更新插件开发指南
```

### PR 描述模板

```markdown
## 变更类型
- [ ] 新功能
- [ ] Bug 修复
- [ ] 文档更新
- [ ] 代码重构
- [ ] 其他

## 变更说明
简要描述这个 PR 做了什么

## 相关 Issue
关闭 #issue_number

## 测试说明
如何测试这些变更

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 已添加测试
- [ ] 文档已更新
- [ ] 所有测试通过
```

### 代码审查标准

我们会检查：
- ✅ 代码质量和可读性
- ✅ 是否遵循代码规范
- ✅ 是否有足够的测试
- ✅ 文档是否完整
- ✅ 性能影响
- ✅ 向后兼容性

## 社区准则

### 行为准则

- 尊重所有贡献者
- 保持专业和友好
- 接受建设性批评
- 关注对社区最好的方案

### 沟通方式

- GitHub Issues：问题跟踪
- Pull Requests：代码审查
- Discussions：一般讨论
- 保持讨论公开透明

## 获得帮助

### 文档资源

- [开发文档](/dev/)
- [插件开发](/dev/plugin-development)
- [架构设计](/dev/architecture)

### 社区支持

- [GitHub Discussions](https://github.com/lian-yang/ltools/discussions)
- [Issue Tracker](https://github.com/lian-yang/ltools/issues)

## 认可贡献

我们会在以下地方感谢贡献者：
- Release Notes
- README 贡献者列表
- 项目主页

## 许可证

通过贡献代码，你同意你的代码将采用 [MIT 许可证](https://github.com/lian-yang/ltools/blob/main/LICENSE) 发布。

---

再次感谢你的贡献！🎉
