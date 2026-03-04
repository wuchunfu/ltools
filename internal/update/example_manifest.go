package update

// update.json 示例文件
// 此文件应托管在更新服务器上

/*
{
  "version": "0.2.0",
  "releaseDate": "2025-03-04T12:00:00Z",
  "releaseNotes": "## 新功能\n- 自动更新机制\n- 音乐播放器优化\n\n## 修复\n- 修复剪贴板历史 bug",
  "mandatory": false,
  "platforms": {
    "darwin-arm64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-darwin-arm64.tar.gz",
      "size": 52428800,
      "checksum": "sha256:abc123...",
      "patches": {
        "0.1.0": {
          "url": "https://updates.ltools.app/stable/patches/0.1.0-to-0.2.0-darwin-arm64.patch",
          "size": 2097152,
          "checksum": "sha256:def456..."
        }
      }
    },
    "darwin-amd64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-darwin-amd64.tar.gz",
      "size": 54525952,
      "checksum": "sha256:ghi789...",
      "patches": {}
    },
    "windows-amd64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-windows-amd64.zip",
      "size": 56623104,
      "checksum": "sha256:jkl012..."
    },
    "linux-amd64": {
      "url": "https://updates.ltools.app/stable/ltools-0.2.0-linux-amd64.tar.gz",
      "size": 53477376,
      "checksum": "sha256:mno345..."
    }
  }
}
*/
