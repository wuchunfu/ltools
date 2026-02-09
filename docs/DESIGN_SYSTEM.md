# LTools UI 重新设计总结

## 设计系统概览

### 核心设计理念
- **风格**: Glassmorphism（玻璃拟态）+ 深色开发者主题
- **主色调**: 紫罗兰 (#7C3AED)
- **设计目标**: 现代、专业、适合开发者工具的界面

---

## 颜色系统

### 主色调
```
Primary:      #7C3AED (紫罗兰)
Hover:        #6D28D9
Light:        #A78BFA
Lighter:      #C4B5FD
```

### 辅助色
```
Success:      #22C55E (绿色)
Warning:      #F59E0B (琥珀色)
Error:        #EF4444 (红色)
Info:         #3B82F6 (蓝色)
```

### 背景色（深色主题）
```
Primary:      #0D0F1A (主背景)
Secondary:    #151925 (次级背景)
Tertiary:     #1A1F2E (三级背景)
Elevated:     #212838 (浮起元素)
```

### 文本色
```
Primary:      #FAF5FF (主要文本)
Secondary:    rgba(250, 245, 255, 0.7)
Tertiary:     rgba(250, 245, 255, 0.5)
Muted:        rgba(250, 245, 255, 0.3)
```

### 边框色
```
Subtle:       rgba(255, 255, 255, 0.06)
Default:      rgba(255, 255, 255, 0.1)
Strong:       rgba(255, 255, 255, 0.15)
```

---

## 字体系统

### 标题字体
```
字体: Space Grotesk
用途: 标题、导航、重要文本
字重: 400, 500, 600, 700
```

### 正文字体
```
字体: DM Sans
用途: 正文、UI 文本、描述
字重: 400, 500, 700
```

### 字体大小
```
H1:  2.5rem (40px)
H2:  2rem (32px)
H3:  1.5rem (24px)
H4:  1.25rem (20px)
H5:  1.125rem (18px)
H6:  1rem (16px)
Body: 1rem (16px)
Small: 0.875rem (14px)
```

---

## Glassmorphism 效果

### 基础玻璃效果
```css
.glass {
  background: rgba(33, 40, 56, 0.6);
  backdrop-filter: blur(12px);
  border: 1px solid var(--color-border-default);
}
```

### 轻量玻璃（卡片）
```css
.glass-light {
  background: rgba(33, 40, 56, 0.4);
  backdrop-filter: blur(8px);
  border: 1px solid var(--color-border-subtle);
}
```

### 重度玻璃（模态框）
```css
.glass-heavy {
  background: rgba(13, 15, 26, 0.85);
  backdrop-filter: blur(20px);
  border: 1px solid var(--color-border-strong);
}
```

---

## 交互设计

### 过渡时长
```
快速交互:    150ms
标准交互:    200ms
慢速交互:    300ms
```

### 缓动函数
```
进入:        ease-out
退出:        ease-in
标准:        ease-out
```

### 悬停效果
- 卡片悬停: `hover-lift` (上移 2px + 阴影)
- 按钮: 背景色变化 + 可选的上移效果
- 链接: 颜色变化 + 可选的下划线

### 状态指示
```
已启用:  bg-[#22C55E]/10 text-[#22C55E]
已禁用:  bg-[#F59E0B]/10 text-[#F59E0B]
错误:    bg-[#EF4444]/10 text-[#EF4444]
```

---

## 组件规范

### 按钮

#### 主要按钮
```tsx
<button className="px-6 py-3 bg-[#7C3AED] hover:bg-[#6D28D9]
  text-white rounded-lg transition-all duration-200
  font-medium clickable hover-lift">
  按钮
</button>
```

#### 次要按钮
```tsx
<button className="px-6 py-3 bg-white/5 hover:bg-white/10
  text-white/80 rounded-lg transition-all duration-200
  font-medium clickable border border-white/10">
  按钮
</button>
```

#### 危险按钮
```tsx
<button className="px-6 py-3 bg-[#EF4444]/10 hover:bg-[#EF4444]/20
  text-[#EF4444] rounded-lg transition-all duration-200
  font-medium clickable border border-[#EF4444]/20">
  删除
</button>
```

### 输入框
```tsx
<input className="w-full px-4 py-3
  bg-[#0D0F1A]/50 border border-white/10 rounded-lg
  text-white placeholder-white/30
  focus:outline-none focus:ring-2 focus:ring-[#7C3AED]/50
  focus:border-[#7C3AED]/50
  transition-all duration-200"
  placeholder="输入内容..."
/>
```

### 卡片
```tsx
<div className="glass-light rounded-xl p-5
  hover-lift transition-all duration-200">
  {/* 卡片内容 */}
</div>
```

### 徽章
```tsx
<span className="px-3 py-1 rounded-full
  text-xs font-medium border
  bg-[#7C3AED]/10 text-[#A78BFA]
  border-[#7C3AED]/20">
  标签
</span>
```

---

## 图标系统

### 使用 Heroicons SVG 图标
- **不使用 emoji 作为图标**
- 所有图标使用 SVG 格式
- 统一使用 `Icon` 组件

```tsx
import { Icon } from './components/Icon';

<Icon name="home" size={20} />
<Icon name="search" size={20} color="#A78BFA" />
```

### 可用图标
- 导航: `home`, `puzzle-piece`, `clock`, `cog`
- 状态: `check-circle`, `x-circle`, `exclamation-circle`, `information-circle`
- 操作: `search`, `chevron-down`, `chevron-right`, `external-link`, `refresh`
- 其他: `sparkles`, `cube`

---

## 无障碍设计

### Focus 状态
```css
*:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
  border-radius: 4px;
}
```

### 对比度
- 正文文本对比度 ≥ 4.5:1
- 大文本对比度 ≥ 3:1
- 交互元素有明确的视觉反馈

### 键盘导航
- 所有交互元素可通过键盘访问
- Tab 顺序符合视觉顺序
- Focus 状态清晰可见

---

## 动画原则

### 尊重用户偏好
```css
@media (prefers-reduced-motion: no-preference) {
  /* 仅在用户不偏好减少动画时应用动画 */
}
```

### 动画类型
```
fadeIn:     淡入
slideUp:    上滑进入
scaleIn:    缩放进入
```

### 骨架屏加载
```tsx
<div className="skeleton h-4 w-full" />
```

---

## 响应式断点

```
Mobile:    375px
Tablet:    768px
Desktop:   1024px
Wide:      1440px
```

### 网格布局
```tsx
// 移动端: 1 列
// 平板: 2 列
// 桌面: 3 列
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
```

---

## 最佳实践

### DO (推荐做法)
- ✅ 使用 SVG 图标而非 emoji
- ✅ 为所有可点击元素添加 `cursor-pointer` 或 `clickable` 类
- ✅ 使用平滑过渡 (150-300ms)
- ✅ 确保 Focus 状态可见
- ✅ 使用 `tabular-nums` 用于数字显示
- ✅ 提供加载状态反馈

### DON'T (避免做法)
- ❌ 使用 emoji 作为 UI 图标
- ❌ 使用 scale 变换导致布局偏移
- ❌ 过长的动画 (>500ms)
- ❌ 忽略 `prefers-reduced-motion`
- ❌ 低对比度的文本颜色
- ❌ 没有加载状态的异步操作

---

## 文件结构

```
frontend/src/
├── components/
│   ├── Icon.tsx              # 通用图标组件
│   ├── App.tsx               # 主应用（侧边栏 + 内容区）
│   ├── PluginMarket.tsx      # 插件市场
│   └── DateTimeWidget.tsx    # 日期时间组件
├── styles.css                # 全局样式和设计系统
└── plugins/                  # 插件系统
```

---

## 迁移指南

### 更新旧组件
1. 将颜色从蓝色 (#3b82f6) 更改为紫罗兰色 (#7C3AED)
2. 将 emoji 图标替换为 SVG `Icon` 组件
3. 应用 `glass-light` 或 `glass` 类到卡片
4. 添加 `clickable` 类到可点击元素
5. 确保 Focus 状态可见
6. 添加适当的过渡动画 (`transition-all duration-200`)

### 新组件开发
1. 使用设计系统中的颜色变量
2. 遵循组件规范
3. 确保无障碍性
4. 测试深色模式
5. 添加加载和错误状态

---

## 下一步优化建议

1. **添加浅色主题支持**
   - 定义浅色模式的颜色变量
   - 添加主题切换功能

2. **扩展图标库**
   - 添加更多图标到 `Icon` 组件
   - 考虑使用图标字体或 SVG sprite

3. **动画库集成**
   - 考虑使用 Framer Motion 进行复杂动画
   - 保持微动画使用 CSS

4. **组件库**
   - 提取可复用组件到独立文件
   - 创建 Storybook 进行组件文档

5. **测试**
   - 添加视觉回归测试
   - 测试不同分辨率下的显示效果
   - 验证无障碍功能

---

**设计版本**: 1.0
**最后更新**: 2026-02-07
**设计师**: Claude (UI/UX Pro Max)
