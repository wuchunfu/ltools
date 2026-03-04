#!/bin/bash

# 音乐播放器代理诊断脚本

echo "🎵 音乐播放器代理诊断"
echo "====================="
echo ""

# 获取最近的音乐相关日志
echo "📋 最近的代理日志："
echo "---"
tail -100 /tmp/ltools-*.log 2>/dev/null | grep -E "\[ProxyManager\]|\[ServiceLX\]|Failed|Error" | tail -20
echo ""

echo "🔍 检查代理服务状态："
echo "---"
echo "1. 检查 LX Music 服务进程："
ps aux | grep "lx-music-service" | grep -v grep
echo ""

echo "2. 测试代理 URL（需要手动提供 resourceID）："
echo "   示例：curl -I http://localhost:34115/proxy/audio/<resourceID>"
echo ""

echo "📊 浏览器调试步骤："
echo "---"
echo "1. 打开浏览器开发者工具（F12）"
echo "2. 切换到 Network 标签"
echo "3. 勾选 'Disable cache'"
echo "4. 尝试播放一首歌曲"
echo "5. 查看网络请求："
echo "   - 找到 /proxy/audio/* 请求"
echo "   - 检查状态码（应该是 206）"
echo "   - 检查响应头中的 Content-Type"
echo "   - 检查响应头中的 Content-Range"
echo "   - 检查是否有 CORS 错误"
echo ""

echo "6. 查看控制台错误："
echo "   - 找到 🎵 相关的错误日志"
echo "   - 记录完整的错误信息"
echo "   - 记录错误发生的文件和行号"
echo ""

echo "🔧 常见问题排查："
echo "---"
echo "问题 1: NotSupportedError"
echo "  原因: 浏览器不支持音频格式或 CORS 配置错误"
echo "  解决: 检查响应头 Content-Type 和 CORS 头"
echo ""

echo "问题 2: 404 Not Found"
echo "  原因: 资源未注册或 URL 错误"
echo "  解决: 检查 ServiceLX 是否正确调用 RegisterAudioURL"
echo ""

echo "问题 3: Network Error"
echo "  原因: 代理服务器未启动或端口被占用"
echo "  解决: 检查 Wails Assets Server 是否正常运行"
echo ""

echo "📝 需要提供的信息："
echo "---"
echo "如果问题仍然存在，请提供："
echo "1. 浏览器控制台的完整错误信息（截图或文本）"
echo "2. Network 标签中失败请求的详情（Headers、Response）"
echo "3. 后端日志中的相关输出"
echo "4. 操作系统和浏览器版本"
