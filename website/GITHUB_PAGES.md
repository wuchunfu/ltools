# 🚀 GitHub Pages 配置

## 快速配置（2 分钟）

### 1️⃣ 启用 GitHub Pages

访问：https://github.com/lian-yang/ltools/settings/pages

在 **Source** 下拉菜单中选择：
- ✅ **GitHub Actions**（重要！）

### 2️⃣ 触发首次部署

访问：https://github.com/lian-yang/ltools/actions/workflows/deploy-website.yml

点击 **Run workflow** → **Run workflow**

### 3️⃣ 访问网站

等待 2-3 分钟后访问：
**https://lian-yang.github.io/ltools/**

## 📖 详细指南

- [完整配置指南](docs/GITHUB_PAGES_SETUP.md)
- [快速开始](docs/QUICK_START_PAGES.md)

## ✅ 验证部署状态

```bash
# 检查 GitHub Actions 状态
gh run list --workflow=deploy-website.yml --limit 1

# 或访问
# https://github.com/lian-yang/ltools/actions
```

## 🔧 本地预览

```bash
# 安装依赖
npm install

# 开发模式
npm run docs:dev

# 构建预览
npm run docs:build
npm run docs:preview
```

## 📝 更新网站

修改 `docs/` 下的文件后推送即可自动部署：

```bash
git add docs/
git commit -m "docs: 更新文档"
git push
```

---

🎉 配置完成后，每次推送都会自动部署！
