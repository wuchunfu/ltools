/**
 * LTools 图标生成器
 * 使用 Node.js Canvas 生成应用图标
 */

const { createCanvas } = require('canvas');
const fs = require('fs');
const path = require('path');

// 颜色配置
const colors = {
  primary: '#7C3AED',
  secondary: '#A78BFA',
  accent: '#F59E0B',
  white: '#FFFFFF',
  dark: '#0D0F1A'
};

function createLinearGradient(ctx, x0, y0, x1, y1, stops) {
  const gradient = ctx.createLinearGradient(x0, y0, x1, y1);
  stops.forEach(stop => gradient.addColorStop(stop.pos, stop.color));
  return gradient;
}

function createRadialGradient(ctx, x0, y0, r0, x1, y1, r1, stops) {
  const gradient = ctx.createRadialGradient(x0, y0, r0, x1, y1, r1);
  stops.forEach(stop => gradient.addColorStop(stop.pos, stop.color));
  return gradient;
}

function drawIcon(size = 1024) {
  const canvas = createCanvas(size, size);
  const ctx = canvas.getContext('2d');
  const center = size / 2;

  // 清空画布
  ctx.clearRect(0, 0, size, size);

  // 背景渐变
  const bgGradient = createLinearGradient(ctx, 0, 0, size, size, [
    { pos: 0, color: colors.primary },
    { pos: 1, color: colors.secondary }
  ]);

  // 圆角矩形背景
  const radius = size * 0.18;
  ctx.beginPath();
  ctx.roundRect(0, 0, size, size, radius);
  ctx.fillStyle = bgGradient;
  ctx.fill();

  // 计算立方体尺寸
  const cubeSize = size * 0.45;
  const cubeY = center - cubeSize * 0.1;
  const isometric = Math.cos(Math.PI / 6);
  const faceHeight = cubeSize * 0.577;

  // 绘制立方体阴影
  ctx.save();
  ctx.translate(center, cubeY + faceHeight * 0.5);
  ctx.beginPath();
  ctx.ellipse(0, cubeSize * 0.15, cubeSize * 0.5, cubeSize * 0.15, 0, 0, Math.PI * 2);
  ctx.fillStyle = 'rgba(0, 0, 0, 0.2)';
  ctx.shadowColor = 'rgba(0, 0, 0, 0.3)';
  ctx.shadowBlur = 20;
  ctx.fill();
  ctx.restore();

  // 立方体顶点
  const cx = center;
  const cy = cubeY - faceHeight * 0.3;
  const hw = cubeSize * 0.5;
  const hh = faceHeight * 0.5;

  // 顶面（菱形）- 最亮
  ctx.beginPath();
  ctx.moveTo(cx, cy - hh);
  ctx.lineTo(cx + hw * isometric, cy);
  ctx.lineTo(cx, cy + hh);
  ctx.lineTo(cx - hw * isometric, cy);
  ctx.closePath();
  ctx.fillStyle = 'rgba(255, 255, 255, 0.95)';
  ctx.fill();
  ctx.strokeStyle = 'rgba(167, 139, 250, 0.3)';
  ctx.lineWidth = size * 0.01;
  ctx.stroke();

  // 右面 - 中等亮度
  ctx.beginPath();
  ctx.moveTo(cx + hw * isometric, cy);
  ctx.lineTo(cx + hw * isometric, cy + faceHeight);
  ctx.lineTo(cx, cy + hh + faceHeight);
  ctx.lineTo(cx, cy + hh);
  ctx.closePath();
  const rightGradient = createLinearGradient(ctx, cx + hw * isometric, cy, cx, cy + hh + faceHeight, [
    { pos: 0, color: 'rgba(255, 255, 255, 0.85)' },
    { pos: 1, color: 'rgba(200, 200, 220, 0.8)' }
  ]);
  ctx.fillStyle = rightGradient;
  ctx.fill();
  ctx.strokeStyle = 'rgba(167, 139, 250, 0.3)';
  ctx.lineWidth = size * 0.01;
  ctx.stroke();

  // 左面 - 最暗
  ctx.beginPath();
  ctx.moveTo(cx - hw * isometric, cy);
  ctx.lineTo(cx - hw * isometric, cy + faceHeight);
  ctx.lineTo(cx, cy + hh + faceHeight);
  ctx.lineTo(cx, cy + hh);
  ctx.closePath();
  const leftGradient = createLinearGradient(ctx, cx - hw * isometric, cy, cx, cy + hh + faceHeight, [
    { pos: 0, color: 'rgba(230, 230, 250, 0.75)' },
    { pos: 1, color: 'rgba(180, 180, 210, 0.7)' }
  ]);
  ctx.fillStyle = leftGradient;
  ctx.fill();
  ctx.strokeStyle = 'rgba(167, 139, 250, 0.3)';
  ctx.lineWidth = size * 0.01;
  ctx.stroke();

  // 添加插件网格纹理到顶面
  ctx.save();
  ctx.beginPath();
  ctx.moveTo(cx, cy - hh);
  ctx.lineTo(cx + hw * isometric, cy);
  ctx.lineTo(cx, cy + hh);
  ctx.lineTo(cx - hw * isometric, cy);
  ctx.closePath();
  ctx.clip();

  // 绘制 2x2 网格线
  ctx.strokeStyle = 'rgba(124, 58, 237, 0.15)';
  ctx.lineWidth = size * 0.008;

  const gridTop = cy - hh;
  const gridBottom = cy + hh;
  const gridLeft = cx - hw * isometric;
  const gridRight = cx + hw * isometric;

  // 水平中线
  ctx.beginPath();
  ctx.moveTo(gridLeft, cy);
  ctx.lineTo(gridRight, cy);
  ctx.stroke();

  // 垂直中线
  ctx.beginPath();
  ctx.moveTo(cx, gridTop);
  ctx.lineTo(cx, gridBottom);
  ctx.stroke();

  ctx.restore();

  // 高光效果
  ctx.save();
  ctx.globalAlpha = 0.3;
  const highlightGradient = createRadialGradient(
    ctx,
    cx - cubeSize * 0.15, cy - faceHeight * 0.3, 0,
    cx - cubeSize * 0.15, cy - faceHeight * 0.3, cubeSize * 0.4,
    [
      { pos: 0, color: 'rgba(255, 255, 255, 0.4)' },
      { pos: 1, color: 'rgba(255, 255, 255, 0)' }
    ]
  );
  ctx.fillStyle = highlightGradient;
  ctx.fillRect(0, 0, size, size);
  ctx.restore();

  // 边缘发光效果
  ctx.save();
  ctx.globalAlpha = 0.1;
  ctx.strokeStyle = colors.white;
  ctx.lineWidth = size * 0.02;
  ctx.beginPath();
  ctx.roundRect(size * 0.02, size * 0.02, size * 0.96, size * 0.96, radius * 0.9);
  ctx.stroke();
  ctx.restore();

  return canvas;
}

// 生成图标
function main() {
  const sizes = [1024, 512, 256, 128, 64, 32, 16];
  const buildDir = path.join(__dirname);

  console.log('正在生成 LTools 图标...\n');

  sizes.forEach(size => {
    const canvas = drawIcon(size);
    const filename = size === 1024 ? 'appicon.png' : `appicon-${size}x${size}.png`;
    const filepath = path.join(buildDir, filename);

    const buffer = canvas.toBuffer('image/png');
    fs.writeFileSync(filepath, buffer);

    console.log(`✓ 生成 ${filename}`);
  });

  console.log('\n图标生成完成！');
  console.log(`主图标: ${path.join(buildDir, 'appicon.png')}`);
  console.log('\n请将 appicon.png 复制到项目目录并重新构建应用。');
}

main();
