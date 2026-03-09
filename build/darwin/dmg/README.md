# DMG 自定义素材说明

此目录用于存放 macOS DMG 安装包的自定义素材。

## 文件结构

```
build/darwin/dmg/
├── background.png      # DMG 窗口背景 (必需)
├── background@2x.png   # Retina 显示屏背景 (可选但推荐)
└── README.md           # 本说明文件
```

## 设计规范

### 尺寸要求

- **标准背景**: 800 x 400 像素 (与 DMG 窗口大小匹配)
- **Retina 背景**: 1600 x 800 像素 (@2x 版本)

### 设计建议

1. **布局区域**
   - 应用图标位置: 左侧 (约 200, 185 像素坐标)
   - 应用程序文件夹: 右侧 (约 600, 185 像素坐标)
   - 拖拽箭头/指示: 中间连接两个图标

2. **视觉风格**
   - 使用 LTools 品牌色调 (#7C3AED 紫色)
   - 保持简洁专业的设计风格
   - 添加简洁的文字说明，如 "Drag to Applications"

3. **格式要求**
   - PNG 格式，支持透明度
   - 颜色模式: RGB
   - 建议添加圆角或阴影以增强视觉效果

### 示例设计

```
┌────────────────────────────────────────┐
│                                        │
│     [LTools Icon]    →    [Apps Icon]  │
│                                        │
│       LTools.app       Applications    │
│                                        │
│         Drag to install                │
│                                        │
└────────────────────────────────────────┘
```

## 启用自定义背景

只需将 `background.png` 放入此目录，DMG 创建脚本会自动检测并使用它。

如果没有自定义背景，DMG 将使用默认布局（无背景图）。

## 创建背景图

可以使用以下工具创建:
- Figma
- Sketch
- Adobe Photoshop/Illustrator
- Affinity Designer

或使用命令行工具:

```bash
# 使用 ImageMagick 创建简单背景
convert -size 800x400 xc:'#1a1a2e' \
  -pointsize 24 -fill white -gravity center \
  -annotate +0+0 "Drag LTools to Applications" \
  background.png
```
