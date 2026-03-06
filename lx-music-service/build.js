const esbuild = require('esbuild');
const fs = require('fs');
const path = require('path');

// 打包主服务
async function build() {
  console.log('Building lx-music-service bundle...');

  // 确保 dist 目录存在
  if (!fs.existsSync('dist')) {
    fs.mkdirSync('dist', { recursive: true });
  }

  // 使用 esbuild 打包
  await esbuild.build({
    entryPoints: ['./server.js'],
    bundle: true,
    platform: 'node',
    target: 'node16',
    outfile: 'dist/server.bundle.js',
    minify: true,
    sourcemap: false,
    // 所有依赖都打包进去，不需要 external
    external: [],
    // 处理 Node.js 内置模块
    define: {
      'process.env.NODE_ENV': '"production"',
    },
  });

  // 验证输出文件
  const bundlePath = 'dist/server.bundle.js';
  if (fs.existsSync(bundlePath)) {
    const stats = fs.statSync(bundlePath);
    console.log(`✅ Bundle created: ${bundlePath}`);
    console.log(`   Size: ${(stats.size / 1024).toFixed(2)} KB`);
  } else {
    throw new Error('Bundle file was not created');
  }

  // 复制插件文件到 dist
  fs.mkdirSync('dist/sources', { recursive: true });

  const sourceFiles = [
    'sources/元力kw1.1.0.js',
    'sources/开心汽水_0.1.5_鸿蒙.js'
  ];

  for (const file of sourceFiles) {
    const sourcePath = path.join(__dirname, file);
    const targetPath = path.join(__dirname, 'dist/sources', path.basename(file));

    if (fs.existsSync(sourcePath)) {
      fs.copyFileSync(sourcePath, targetPath);
      console.log(`✅ Copied: ${file} -> dist/sources/`);
    } else {
      console.warn(`⚠️  Source file not found: ${file}`);
    }
  }

  console.log('\n🎉 Build complete!');
  console.log('   Output: dist/server.bundle.js');
  console.log('   To run: node dist/server.bundle.js');
}

// 错误处理
build().catch((err) => {
  console.error('❌ Build failed:', err);
  process.exit(1);
});
