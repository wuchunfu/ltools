import { useState, useEffect, useCallback } from 'react';
import { Events } from '@wailsio/runtime';
import * as UpdateService from '../../bindings/ltools/internal/update/service';
import { Icon } from '../components/Icon';

interface UpdateInfo {
  version: string;
  size: number;
  patchSize: number;
  hasPatch: boolean;
  releaseDate: string;
  releaseNotes: string;
  mandatory: boolean;
  downloadUrl: string;
  checksum: string;
}

export function UpdateNotification() {
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null);
  const [downloading, setDownloading] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [downloaded, setDownloaded] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dismissed, setDismissed] = useState(false);

  // 先定义所有回调函数
  const handleRestart = useCallback(async () => {
    try {
      setError(null);
      await UpdateService.RestartApp();
    } catch (err) {
      console.error('Restart failed:', err);
      setError(err instanceof Error ? err.message : String(err));
    }
  }, []);

  const handleInstall = useCallback(async (filePath: string) => {
    if (!filePath) {
      setError('安装失败：文件路径不存在');
      return;
    }

    try {
      setInstalling(true);
      setError(null);
      await UpdateService.InstallUpdate(filePath);

      // macOS/Linux: 安装成功后会发送 update:installed 事件
      // Windows: 安装程序会自动退出应用
      console.log('Update installation started...');
    } catch (err) {
      console.error('Installation failed:', err);
      setInstalling(false);
      setDownloaded(false);
      setError(err instanceof Error ? err.message : String(err));
    }
  }, []);

  const handleDownload = useCallback(async () => {
    if (!updateInfo) return;

    // 参数验证
    if (!updateInfo.downloadUrl || !updateInfo.checksum) {
      setError('更新信息不完整，请稍后重试');
      return;
    }

    setDownloading(true);
    setDownloaded(false);
    setDownloadProgress(0);
    setError(null);

    try {
      const filePath = await UpdateService.DownloadUpdate(
        updateInfo.downloadUrl,
        updateInfo.checksum
      );

      console.log('Download completed:', filePath);
      setDownloaded(true);
      setDownloading(false);

      // 自动安装（macOS/Linux 会等待用户确认，Windows 会立即退出）
      await handleInstall(filePath);
    } catch (err) {
      console.error('Download failed:', err);
      setError(err instanceof Error ? err.message : String(err));
      setDownloading(false);
    }
  }, [updateInfo, handleInstall]);

  const handleDismiss = useCallback(() => {
    if (!updateInfo?.mandatory) {
      setDismissed(true);
    }
  }, [updateInfo?.mandatory]);

  const handleCheckUpdate = useCallback(async () => {
    try {
      setError(null);
      const info = await UpdateService.CheckForUpdate();

      if (info) {
        setUpdateInfo(info);
        setDismissed(false);
      } else {
        // 已经是最新版本
        setError('您已经在使用最新版本！');
        setTimeout(() => setError(null), 3000);
      }
    } catch (err) {
      console.error('Check update failed:', err);
      setError(err instanceof Error ? err.message : String(err));
    }
  }, []);

  // 然后使用 useEffect
  useEffect(() => {
    // 监听更新可用事件
    const unsubscribeUpdate = Events.On('update:available', (ev) => {
      console.log('Update available:', ev.data);
      if (ev.data) {
        setUpdateInfo(ev.data);
        setDismissed(false);
      }
    });

    // 监听下载进度事件
    const unsubscribeProgress = Events.On('update:progress', (ev) => {
      console.log('Download progress:', ev.data);
      if (ev.data !== null && ev.data !== undefined) {
        setDownloadProgress(ev.data);
      }
    });

    // 监听安装完成事件
    const unsubscribeInstalled = Events.On('update:installed', (ev) => {
      console.log('Update installed:', ev.data);
      if (ev.data) {
        setInstalling(false);
        setDownloaded(false);

        // 显示重启提示
        if (ev.data.action === 'restart') {
          // 自动重启应用
          handleRestart();
        }
      }
    });

    return () => {
      unsubscribeUpdate();
      unsubscribeProgress();
      unsubscribeInstalled();
    };
  }, [handleRestart]);

  if (!updateInfo || dismissed) {
    // 手动检查更新按钮
    return (
      <div className="fixed bottom-4 right-4">
        <button
          onClick={handleCheckUpdate}
          className="p-2 rounded-lg glass hover:bg-white/10 transition-colors"
          title="检查更新"
        >
          <Icon name="refresh" className="w-5 h-5" />
        </button>
      </div>
    );
  }

  return (
    <div className="fixed bottom-4 right-4 max-w-md bg-gradient-to-r from-purple-500/90 to-pink-500/90 backdrop-blur-lg rounded-xl shadow-2xl text-white p-4 animate-slide-up">
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0">
          <Icon name="download" className="w-6 h-6 mt-1" />
        </div>

        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-lg mb-1">
            发现新版本 {updateInfo.version}
          </h3>

          <div className="text-sm text-white/80 mb-2">
            大小: {updateInfo.hasPatch
              ? `${(updateInfo.patchSize / 1024).toFixed(0)} KB (补丁)`
              : `${(updateInfo.size / 1024 / 1024).toFixed(1)} MB`
            }
          </div>

          {updateInfo.releaseNotes && (
            <div className="text-sm text-white/70 mb-3 max-h-32 overflow-y-auto">
              <pre className="whitespace-pre-wrap font-sans">
                {updateInfo.releaseNotes}
              </pre>
            </div>
          )}

          {error && (
            <div className="bg-red-500/50 text-white px-3 py-2 rounded mb-3 text-sm">
              {error}
            </div>
          )}

          {downloading && (
            <div className="mb-3">
              <div className="bg-white/20 rounded-full h-2 mb-1">
                <div
                  className="bg-white h-2 rounded-full transition-all duration-300"
                  style={{ width: `${downloadProgress}%` }}
                />
              </div>
              <div className="text-sm text-white/80">
                下载中... {downloadProgress}%
              </div>
            </div>
          )}

          <div className="flex gap-2">
            {!downloading && !downloaded && !installing && (
              <>
                <button
                  onClick={handleDownload}
                  className="flex-1 bg-white text-purple-600 font-medium px-4 py-2 rounded-lg hover:bg-white/90 transition-colors"
                >
                  立即下载
                </button>

                {!updateInfo.mandatory && (
                  <button
                    onClick={handleDismiss}
                    className="px-4 py-2 rounded-lg hover:bg-white/20 transition-colors"
                  >
                    稍后提醒
                  </button>
                )}
              </>
            )}

            {downloading && (
              <div className="flex-1 text-center text-white/80 text-sm">
                下载中... {downloadProgress}%
              </div>
            )}

            {installing && (
              <div className="flex-1 text-center text-white/80 text-sm">
                正在安装...
              </div>
            )}
          </div>
        </div>

        {!updateInfo.mandatory && !downloading && (
          <button
            onClick={handleDismiss}
            className="flex-shrink-0 text-white/60 hover:text-white transition-colors"
          >
            <Icon name="x" className="w-5 h-5" />
          </button>
        )}
      </div>
    </div>
  );
}
