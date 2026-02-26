import { useState, useEffect, useCallback, useRef } from 'react';
import * as TunnelService from '../../bindings/ltools/plugins/tunnel/tunnelservice';
import {
  Tunnel,
  TunnelRuntimeInfo,
  InstallationStatus,
  CreateTunnelRequest,
  UpdateTunnelRequest,
  ProtocolType,
  ProxyType,
  FRPServerConfig,
  GlobalOptions
} from '../../bindings/ltools/plugins/tunnel/models';
import { Icon } from './Icon';
import { useToast } from '../hooks/useToast';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import { Events } from '@wailsio/runtime';

type View = 'tunnels' | 'create' | 'edit' | 'settings';

interface TunnelFormData {
  name: string;
  protocol: ProtocolType;
  localHost: string;
  localPort: string;
  proxyType?: ProxyType;
  subdomain: string;
  frpServerAddress: string;
  frpServerToken: string;
  autoStart: boolean;
  enabled: boolean;
}

interface TunnelWidgetProps {
  onBack: () => void;
}

export function TunnelWidget({ onBack }: TunnelWidgetProps): JSX.Element {
  const [view, setView] = useState<View>('tunnels');
  const [tunnels, setTunnels] = useState<Tunnel[]>([]);
  const [statuses, setStatuses] = useState<TunnelRuntimeInfo[]>([]);
  const [installStatus, setInstallStatus] = useState<InstallationStatus | null>(null);
  const [globalOptions, setGlobalOptions] = useState<GlobalOptions | null>(null);
  const [editingTunnel, setEditingTunnel] = useState<Tunnel | null>(null);
  const [showLogModal, setShowLogModal] = useState(false);
  const [currentLog, setCurrentLog] = useState<string>('');
  const [currentLogTitle, setCurrentLogTitle] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const toast = useToast();

  // 表单状态
  const [formData, setFormData] = useState<TunnelFormData>({
    name: '',
    protocol: ProtocolType.ProtocolFRP,
    localHost: '127.0.0.1',
    localPort: '3000',
    proxyType: ProxyType.ProxyTypeHTTP,
    subdomain: '',
    frpServerAddress: '',
    frpServerToken: '',
    autoStart: false,
    enabled: true
  });

  // 设置表单状态
  const [settingsForm, setSettingsForm] = useState({
    frpServerAddress: '',
    frpServerToken: ''
  });

  // 加载隧道列表
  const loadTunnels = useCallback(async () => {
    try {
      const result = await TunnelService.GetTunnels();
      setTunnels(result || []);
    } catch (error) {
      console.error('Failed to load tunnels:', error);
    }
  }, []);

  // 加载状态和安装信息
  const loadStatuses = useCallback(async () => {
    try {
      const [status, install] = await Promise.all([
        TunnelService.GetAllTunnelStatuses(),
        TunnelService.GetInstallationStatus()
      ]);
      setStatuses(status || []);
      setInstallStatus(install || null);
    } catch (error) {
      console.error('Failed to load statuses:', error);
    }
  }, []);

  // 加载全局配置
  const loadGlobalOptions = useCallback(async () => {
    try {
      const opts = await TunnelService.GetGlobalOptions();
      setGlobalOptions(opts);
      if (opts?.frpServer) {
        setSettingsForm({
          frpServerAddress: opts.frpServer.address || '',
          frpServerToken: opts.frpServer.token || ''
        });
      }
    } catch (error) {
      console.error('Failed to load global options:', error);
    }
  }, []);

  // 初始加载
  useEffect(() => {
    loadTunnels();
    loadStatuses();
    loadGlobalOptions();
  }, [loadTunnels, loadStatuses, loadGlobalOptions]);

  // 定时刷新状态
  useEffect(() => {
    const interval = setInterval(() => {
      loadStatuses();
    }, 3000);
    return () => clearInterval(interval);
  }, [loadStatuses]);

  // 监听后端事件
  useEffect(() => {
    const unsubscribers: (() => void)[] = [];

    // 隧道启动成功
    unsubscribers.push(Events.On('tunnel:started', (ev: { data: string }) => {
      toast.success(`隧道 "${ev.data}" 已启动`);
      loadStatuses();
    }));

    // 隧道停止
    unsubscribers.push(Events.On('tunnel:stopped', (ev: { data: string }) => {
      toast.info(`隧道 "${ev.data}" 已停止`);
      loadStatuses();
    }));

    // 隧道错误
    unsubscribers.push(Events.On('tunnel:error', (ev: { data: { tunnelId: string; error: string } }) => {
      toast.error(`隧道 "${ev.data.tunnelId}" 错误: ${ev.data.error}`);
      loadStatuses();
    }));

    // 隧道 URL 更新
    unsubscribers.push(Events.On('tunnel:url', (ev: { data: { tunnelId: string; url: string } }) => {
      toast.success(`隧道 "${ev.data.tunnelId}" 公网地址: ${ev.data.url}`);
      loadStatuses();
    }));

    // 安装进度
    unsubscribers.push(Events.On('tunnel:install:progress', (ev: { data: string }) => {
      console.log('[FRP Install]', ev.data);
    }));

    return () => {
      unsubscribers.forEach(unsub => unsub());
    };
  }, [loadStatuses, toast]);

  const handleDeleteTunnel = async (id: string) => {
    if (!confirm('确定要删除此隧道吗？')) return;

    setIsLoading(true);
    try {
      const result = await TunnelService.DeleteTunnel(id);
      if (result?.success) {
        toast.success('隧道删除成功');
        await loadTunnels();
      } else {
        toast.error(result?.error || '删除失败');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleStartTunnel = async (id: string) => {
    setIsLoading(true);
    try {
      const result = await TunnelService.StartTunnel(id);
      if (result?.success) {
        toast.success('隧道启动成功');
        await loadStatuses();
      } else {
        toast.error(result?.error || '启动失败');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleStopTunnel = async (id: string) => {
    setIsLoading(true);
    try {
      const result = await TunnelService.StopTunnel(id);
      if (result?.success) {
        toast.success('隧道停止成功');
        await loadStatuses();
      } else {
        toast.error(result?.error || '停止失败');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleRestartTunnel = async (id: string) => {
    setIsLoading(true);
    try {
      const result = await TunnelService.RestartTunnel(id);
      if (result?.success) {
        toast.success('隧道重启成功');
        await loadStatuses();
      } else {
        toast.error(result?.error || '重启失败');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleInstallFRP = async () => {
    setIsLoading(true);
    try {
      const result = await TunnelService.InstallFRP();
      if (result?.success) {
        toast.success('FRP 安装成功');
        await loadStatuses();
      } else {
        toast.error(result?.error || '安装失败');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleSaveSettings = async () => {
    setIsLoading(true);
    try {
      const opts = new GlobalOptions({
        defaultProtocol: ProtocolType.ProtocolFRP,
        frpServer: new FRPServerConfig({
          address: settingsForm.frpServerAddress,
          token: settingsForm.frpServerToken
        })
      });

      const result = await TunnelService.SetGlobalOptions(opts);
      if (result?.success) {
        toast.success('设置保存成功');
        await loadGlobalOptions();
      } else {
        toast.error(result?.error || '保存失败');
      }
    } finally {
      setIsLoading(false);
    }
  };

  // 复制到剪贴板
  const handleCopy = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast.success(`${label}已复制到剪贴板`);
    } catch (err) {
      toast.error('复制失败');
    }
  };

  // 查看日志
  const [currentLogTunnelId, setCurrentLogTunnelId] = useState<string>('');
  const [autoRefreshLog, setAutoRefreshLog] = useState<boolean>(true);
  const logRefreshIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const handleViewLog = async (tunnel: Tunnel) => {
    const status = getTunnelStatus(tunnel.id);
    if (!status?.logPath) {
      toast.info('暂无日志文件');
      return;
    }

    setCurrentLogTunnelId(tunnel.id);
    setCurrentLogTitle(`隧道 "${tunnel.name}" 日志`);
    setShowLogModal(true);
    setAutoRefreshLog(true);

    // 立即加载日志
    await loadLogContent(tunnel.id);
  };

  const loadLogContent = async (tunnelId: string, lineCount: number = 100) => {
    try {
      const logContent = await TunnelService.GetTunnelLog(tunnelId, lineCount);
      setCurrentLog(logContent || '(暂无日志内容)');
    } catch (error: any) {
      setCurrentLog(`加载日志失败: ${error?.message || '未知错误'}`);
    }
  };

  // 日志自动刷新
  useEffect(() => {
    if (showLogModal && autoRefreshLog && currentLogTunnelId) {
      logRefreshIntervalRef.current = setInterval(() => {
        loadLogContent(currentLogTunnelId, 100);
      }, 2000);
    }

    return () => {
      if (logRefreshIntervalRef.current) {
        clearInterval(logRefreshIntervalRef.current);
        logRefreshIntervalRef.current = null;
      }
    };
  }, [showLogModal, autoRefreshLog, currentLogTunnelId]);

  const handleCloseLogModal = () => {
    setShowLogModal(false);
    setCurrentLogTunnelId('');
    setCurrentLog('');
    if (logRefreshIntervalRef.current) {
      clearInterval(logRefreshIntervalRef.current);
      logRefreshIntervalRef.current = null;
    }
  };

  // 重置表单
  const resetForm = () => {
    setFormData({
      name: '',
      protocol: ProtocolType.ProtocolFRP,
      localHost: '127.0.0.1',
      localPort: '3000',
      proxyType: ProxyType.ProxyTypeHTTP,
      subdomain: '',
      frpServerAddress: globalOptions?.frpServer?.address || '',
      frpServerToken: globalOptions?.frpServer?.token || '',
      autoStart: false,
      enabled: true
    });
  };

  // 初始化编辑表单
  const initEditForm = (tunnel: Tunnel) => {
    setFormData({
      name: tunnel.name,
      protocol: tunnel.protocol,
      localHost: tunnel.localHost,
      localPort: tunnel.localPort.toString(),
      proxyType: tunnel.proxyType || ProxyType.ProxyTypeHTTP,
      subdomain: tunnel.subdomain || '',
      frpServerAddress: tunnel.frpServer?.address || '',
      frpServerToken: tunnel.frpServer?.token || '',
      autoStart: tunnel.autoStart,
      enabled: tunnel.enabled
    });
  };

  // 切换到创建视图时重置表单
  const handleSwitchToCreate = () => {
    resetForm();
    setView('create');
  };

  // 切换到编辑视图时初始化表单
  const handleSwitchToEdit = (tunnel: Tunnel) => {
    setEditingTunnel(tunnel);
    initEditForm(tunnel);
    setView('edit');
  };

  // 验证表单
  const validateForm = (isEdit: boolean = false): { valid: boolean; error?: string } => {
    if (!formData.name.trim()) {
      return { valid: false, error: '请输入隧道名称' };
    }
    if (!formData.localHost.trim()) {
      return { valid: false, error: '请输入本地地址' };
    }
    const port = parseInt(formData.localPort);
    if (isNaN(port) || port < 1 || port > 65535) {
      return { valid: false, error: '端口号必须在 1-65535 之间' };
    }
    if (!isEdit || view === 'edit') {
      if (formData.protocol === ProtocolType.ProtocolFRP) {
        if (!formData.frpServerAddress.trim()) {
          return { valid: false, error: '请输入 FRP 服务器地址' };
        }
      }
    }
    return { valid: true };
  };

  // 处理创建隧道
  const handleCreateTunnel = async () => {
    const validation = validateForm(false);
    if (!validation.valid) {
      toast.error(validation.error || '表单验证失败');
      return;
    }

    setIsLoading(true);
    try {
      const request = new CreateTunnelRequest({
        name: formData.name,
        protocol: formData.protocol,
        localHost: formData.localHost,
        localPort: parseInt(formData.localPort),
        autoStart: formData.autoStart
      });

      const result = await TunnelService.CreateTunnel(request);

      if (result?.success) {
        // 如果是 FRP 协议，创建后需要更新以添加 FRP 配置
        if (formData.protocol === ProtocolType.ProtocolFRP) {
          await loadTunnels();
          const newTunnel = [...tunnels].find(t => t.name === formData.name);
          if (newTunnel) {
            const updateReq = new UpdateTunnelRequest({
              name: formData.name,
              protocol: formData.protocol,
              localHost: formData.localHost,
              localPort: parseInt(formData.localPort),
              enabled: true,
              autoStart: formData.autoStart,
              frpServer: new FRPServerConfig({
                address: formData.frpServerAddress,
                token: formData.frpServerToken
              }),
              proxyType: formData.proxyType,
              subdomain: formData.subdomain
            });

            await TunnelService.UpdateTunnel(newTunnel.id, updateReq);
          }
        }

        toast.success('隧道创建成功');
        await loadTunnels();
        setView('tunnels');
      } else {
        toast.error(result?.error || '创建失败');
      }
    } finally {
      setIsLoading(false);
    }
  };

  // 处理更新隧道
  const handleUpdateTunnel = async () => {
    if (!editingTunnel) return;

    const validation = validateForm(true);
    if (!validation.valid) {
      toast.error(validation.error || '表单验证失败');
      return;
    }

    setIsLoading(true);
    try {
      const request = new UpdateTunnelRequest({
        name: formData.name,
        protocol: formData.protocol,
        localHost: formData.localHost,
        localPort: parseInt(formData.localPort),
        enabled: formData.enabled,
        autoStart: formData.autoStart
      });

      // 如果是 FRP 协议，添加 FRP 相关配置
      if (formData.protocol === ProtocolType.ProtocolFRP) {
        request.frpServer = new FRPServerConfig({
          address: formData.frpServerAddress,
          token: formData.frpServerToken
        });
        request.proxyType = formData.proxyType;
        request.subdomain = formData.subdomain;
      }

      const result = await TunnelService.UpdateTunnel(editingTunnel.id, request);

      if (result?.success) {
        toast.success('隧道更新成功');
        await loadTunnels();
        setView('tunnels');
      } else {
        toast.error(result?.error || '更新失败');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const getTunnelStatus = (tunnelId: string): TunnelRuntimeInfo | undefined => {
    return statuses.find(s => s.tunnelId === tunnelId);
  };

  const getStatusDisplay = (status?: TunnelRuntimeInfo) => {
    if (!status) return { text: '已停止', color: 'bg-white/10 text-white/60', icon: 'x-circle' };
    switch (status.status) {
      case 'running':
        return { text: '运行中', color: 'bg-[#22C55E] text-white', icon: 'check-circle' };
      case 'starting':
        return { text: '启动中', color: 'bg-[#F59E0B] text-white', icon: 'refresh' };
      case 'error':
        return { text: '错误', color: 'bg-[#EF4444] text-white', icon: 'alert-circle' };
      default:
        return { text: '已停止', color: 'bg-white/10 text-white/60', icon: 'x-circle' };
    }
  };

  const renderTunnelsView = () => (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-xl font-semibold text-white">隧道管理</h2>
          <p className="text-sm text-white/50 mt-1">管理 FRP 内网穿透隧道</p>
        </div>
        <button
          className="px-4 py-2 bg-[#7C3AED] hover:bg-[#6D28D9] rounded-lg text-white font-medium clickable flex items-center"
          onClick={handleSwitchToCreate}
          disabled={isLoading}
        >
          <Icon name="plus" size={16} />
          <span className="ml-2">创建隧道</span>
        </button>
      </div>

      {/* FRP 安装状态提示 */}
      {installStatus && !installStatus.frpInstalled && (
        <div className="glass-light p-4 rounded-xl mb-4 border border-[#F59E0B]/30">
          <div className="flex items-center gap-3">
            <Icon name="alert-circle" size={20} color="#F59E0B" />
            <div className="flex-1">
              <p className="text-white/80">FRP 未安装</p>
              <p className="text-sm text-white/50">请先安装 FRP 或配置 FRP 路径</p>
            </div>
            <button
              className="px-3 py-1.5 bg-[#7C3AED] hover:bg-[#6D28D9] rounded-lg text-white text-sm clickable"
              onClick={() => setView('settings')}
            >
              去设置
            </button>
          </div>
        </div>
      )}

      {tunnels.length === 0 ? (
        <div className="glass-light p-8 rounded-xl text-center text-white/60">
          <Icon name="network" size={48} className="mx-auto mb-4 text-white/40" />
          <p className="text-lg">暂无隧道配置</p>
          <p className="text-sm mt-2">点击上方按钮创建您的第一个隧道</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {tunnels.map(tunnel => {
            const status = getTunnelStatus(tunnel.id);
            const statusDisplay = getStatusDisplay(status);
            const isRunning = status?.status === 'running';
            const isStarting = status?.status === 'starting';

            return (
              <div key={tunnel.id} className="glass-light p-4 rounded-xl">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="font-semibold text-white text-lg">{tunnel.name}</h3>
                    <p className="text-xs text-white/40 mt-0.5">{tunnel.id}</p>
                  </div>
                  <span className={`px-2 py-0.5 rounded text-xs font-medium flex items-center gap-1 ${statusDisplay.color}`}>
                    <Icon name={statusDisplay.icon as any} size={12} />
                    {statusDisplay.text}
                  </span>
                </div>

                <div className="space-y-1.5 text-sm text-white/70">
                  <div className="flex items-center gap-2">
                    <span className="text-white/40 w-12">类型:</span>
                    {tunnel.proxyType ? (
                      <span className="px-1.5 py-0.5 rounded bg-[#7C3AED]/20 text-[#A78BFA] text-xs">
                        {tunnel.proxyType}
                      </span>
                    ) : (
                      <span className="px-1.5 py-0.5 rounded bg-white/10 text-xs">FRP</span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-white/40 w-12">本地:</span>
                    <code className="text-xs bg-white/5 px-1.5 py-0.5 rounded">{tunnel.localHost}:{tunnel.localPort}</code>
                  </div>
                  {tunnel.subdomain && (
                    <div className="flex items-center gap-2">
                      <span className="text-white/40 w-12">子域名:</span>
                      <span>{tunnel.subdomain}</span>
                    </div>
                  )}
                  {status?.publicUrl && (
                    <div className="flex items-center gap-2">
                      <span className="text-white/40 w-12">公网:</span>
                      <a href={status.publicUrl} target="_blank" rel="noopener noreferrer"
                         className="text-[#7C3AED] hover:underline truncate flex-1">
                        {status.publicUrl}
                      </a>
                      <button
                        className="p-1 rounded hover:bg-white/10 clickable"
                        onClick={() => handleCopy(status.publicUrl!, '公网地址')}
                        title="复制地址"
                      >
                        <Icon name="copy" size={14} />
                      </button>
                    </div>
                  )}
                  {status?.lastError && (
                    <div className="text-[#EF4444] text-xs bg-[#EF4444]/10 p-2 rounded mt-2">
                      <span className="font-medium">错误:</span> {status.lastError}
                    </div>
                  )}
                </div>

                <div className="flex items-center justify-between mt-4 pt-4 border-t border-white/10">
                  <div className="flex gap-1.5">
                    {isRunning ? (
                      <button
                        className="p-2 rounded-lg bg-[#EF4444]/20 hover:bg-[#EF4444]/30 clickable text-[#EF4444]"
                        onClick={() => handleStopTunnel(tunnel.id)}
                        disabled={isLoading}
                        title="停止"
                      >
                        <Icon name="stop" size={16} />
                      </button>
                    ) : (
                      <button
                        className="p-2 rounded-lg bg-[#22C55E]/20 hover:bg-[#22C55E]/30 clickable text-[#22C55E]"
                        onClick={() => handleStartTunnel(tunnel.id)}
                        disabled={isLoading || isStarting}
                        title="启动"
                      >
                        <Icon name="play" size={16} />
                      </button>
                    )}
                    <button
                      className="p-2 rounded-lg bg-white/10 hover:bg-white/20 clickable text-white/70"
                      onClick={() => handleRestartTunnel(tunnel.id)}
                      disabled={isLoading || isStarting}
                      title="重启"
                    >
                      <Icon name="refresh" size={16} />
                    </button>
                    <button
                      className="p-2 rounded-lg bg-white/10 hover:bg-white/20 clickable text-white/70"
                      onClick={() => handleViewLog(tunnel)}
                      title="查看日志"
                    >
                      <Icon name="log" size={16} />
                    </button>
                    <button
                      className="p-2 rounded-lg bg-white/10 hover:bg-white/20 clickable text-white/70"
                      onClick={() => handleSwitchToEdit(tunnel)}
                      title="编辑"
                    >
                      <Icon name="pencil" size={16} />
                    </button>
                    <button
                      className="p-2 rounded-lg bg-[#EF4444]/20 hover:bg-[#EF4444]/30 clickable text-[#EF4444]"
                      onClick={() => handleDeleteTunnel(tunnel.id)}
                      disabled={isLoading}
                      title="删除"
                    >
                      <Icon name="trash" size={16} />
                    </button>
                  </div>
                  {tunnel.autoStart && (
                    <span className="text-xs text-white/40 flex items-center gap-1">
                      <Icon name="check" size={12} />
                      自启动
                    </span>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );

  const renderSettingsView = () => (
    <div className="space-y-4">
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-white">FRP 设置</h2>
        <p className="text-sm text-white/50 mt-1">配置默认 FRP 服务器和安装选项</p>
      </div>

      {/* 安装状态 */}
      <div className="glass-light p-4 rounded-xl">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
              installStatus?.frpInstalled ? 'bg-[#22C55E]/20' : 'bg-[#F59E0B]/20'
            }`}>
              <Icon name={installStatus?.frpInstalled ? 'check-circle' : 'alert-circle'} size={20}
                color={installStatus?.frpInstalled ? '#22C55E' : '#F59E0B'} />
            </div>
            <div>
              <h3 className="font-semibold text-white">FRP 安装状态</h3>
              <p className="text-sm text-white/50">
                {installStatus?.frpInstalled
                  ? `已安装${installStatus.frpVersion ? ` (${installStatus.frpVersion})` : ''}`
                  : '未安装'}
              </p>
            </div>
          </div>
          {!installStatus?.frpInstalled && (
            <button
              className="px-4 py-2 bg-[#7C3AED] hover:bg-[#6D28D9] rounded-lg text-white font-medium clickable"
              onClick={handleInstallFRP}
              disabled={isLoading}
            >
              {isLoading ? '安装中...' : '安装 FRP'}
            </button>
          )}
        </div>

        {!installStatus?.frpInstalled && (
          <div className="text-sm text-white/50 bg-white/5 p-3 rounded-lg">
            <p>FRP 是一个高性能的反向代理应用，用于内网穿透。</p>
            <p className="mt-1">也可以手动从 <a href="https://github.com/fatedier/frp/releases" target="_blank" rel="noopener noreferrer"
                            className="text-[#7C3AED] hover:underline">GitHub Releases</a> 下载并安装到系统 PATH。</p>
          </div>
        )}
      </div>

      {/* 默认服务器配置 */}
      <div className="glass-light p-4 rounded-xl">
        <h3 className="font-semibold text-white mb-4 flex items-center gap-2">
          <Icon name="server" size={18} />
          默认 FRP 服务器
        </h3>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-white/70 mb-2">服务器地址</label>
            <input
              type="text"
              value={settingsForm.frpServerAddress}
              onChange={(e) => setSettingsForm({ ...settingsForm, frpServerAddress: e.target.value })}
              className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
              placeholder="例如: frp.example.com:7000"
            />
            <p className="text-xs text-white/40 mt-1">FRP 服务器的地址和端口</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-white/70 mb-2">认证 Token</label>
            <input
              type="password"
              value={settingsForm.frpServerToken}
              onChange={(e) => setSettingsForm({ ...settingsForm, frpServerToken: e.target.value })}
              className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
              placeholder="输入服务器 Token"
            />
            <p className="text-xs text-white/40 mt-1">用于连接 FRP 服务器的认证令牌</p>
          </div>

          <div className="pt-4 border-t border-white/10">
            <button
              className="px-6 py-2 bg-[#7C3AED] hover:bg-[#6D28D9] rounded-lg text-white font-medium clickable"
              onClick={handleSaveSettings}
              disabled={isLoading}
            >
              {isLoading ? '保存中...' : '保存设置'}
            </button>
          </div>
        </div>
      </div>

      {/* 使用说明 */}
      <div className="glass-light p-4 rounded-xl">
        <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
          <Icon name="information-circle" size={18} />
          使用说明
        </h3>
        <div className="text-sm text-white/60 space-y-2">
          <p>1. 确保 FRP 服务端 (frps) 已部署并运行</p>
          <p>2. 在设置中配置默认 FRP 服务器地址和 Token</p>
          <p>3. 创建隧道时选择代理类型（HTTP、HTTPS、TCP 等）</p>
          <p>4. 启动隧道后，系统会分配公网访问地址</p>
          <p>5. 支持子域名配置（需要服务端支持）</p>
        </div>
      </div>
    </div>
  );

  // 渲染创建表单
  const renderCreateForm = () => (
    <div className="glass-light p-6 rounded-xl">
      <h2 className="text-xl font-semibold text-white mb-6">创建新隧道</h2>

      <div className="space-y-4">
        {/* 基本信息 */}
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-white/70 mb-2">隧道名称 *</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
              placeholder="例如: 我的 Web 服务"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">本地地址 *</label>
              <input
                type="text"
                value={formData.localHost}
                onChange={(e) => setFormData({ ...formData, localHost: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
                placeholder="127.0.0.1"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">本地端口 *</label>
              <input
                type="number"
                value={formData.localPort}
                onChange={(e) => setFormData({ ...formData, localPort: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
                placeholder="3000"
                min="1"
                max="65535"
              />
            </div>
          </div>
        </div>

        {/* FRP 配置 */}
        <div className="space-y-3 pt-4 border-t border-white/10">
          <h3 className="text-lg font-medium text-white mb-2">FRP 配置</h3>

            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">代理类型</label>
              <Select
                value={formData.proxyType}
                onValueChange={(value) => setFormData({ ...formData, proxyType: value as ProxyType })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择代理类型" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ProxyType.ProxyTypeHTTP}>HTTP</SelectItem>
                  <SelectItem value={ProxyType.ProxyTypeHTTPS}>HTTPS</SelectItem>
                  <SelectItem value={ProxyType.ProxyTypeTCP}>TCP</SelectItem>
                  <SelectItem value={ProxyType.ProxyTypeSTCP}>STCP (秘密 TCP)</SelectItem>
                  <SelectItem value={ProxyType.ProxyTypeXTCP}>XTCP (P2P TCP)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">子域名 (可选)</label>
              <input
                type="text"
                value={formData.subdomain}
                onChange={(e) => setFormData({ ...formData, subdomain: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
                placeholder="例如: myapp"
              />
              <p className="text-xs text-white/40 mt-1">需要服务端支持自定义域名</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">服务器地址 *</label>
              <input
                type="text"
                value={formData.frpServerAddress}
                onChange={(e) => setFormData({ ...formData, frpServerAddress: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
                placeholder="例如: frp.example.com:7000"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">认证 Token *</label>
              <input
                type="password"
                value={formData.frpServerToken}
                onChange={(e) => setFormData({ ...formData, frpServerToken: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
                placeholder="输入服务器 Token"
              />
            </div>
        </div>

        {/* 其他选项 */}
        <div className="space-y-3 pt-4 border-t border-white/10">
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={formData.autoStart}
              onChange={(e) => setFormData({ ...formData, autoStart: e.target.checked })}
              className="w-5 h-5 rounded border-white/20 accent-[#7C3AED]"
            />
            <span className="text-sm text-white/70">应用启动时自动启动此隧道</span>
          </label>
        </div>
      </div>

      {/* 按钮组 */}
      <div className="flex gap-3 mt-6 pt-6 border-t border-white/10">
        <button
          className="px-6 py-2 bg-[#7C3AED] hover:bg-[#6D28D9] rounded-lg text-white font-medium clickable disabled:opacity-50"
          onClick={handleCreateTunnel}
          disabled={isLoading}
        >
          {isLoading ? '创建中...' : '创建隧道'}
        </button>
        <button
          className="px-6 py-2 bg-white/10 hover:bg-white/20 rounded-lg text-white font-medium clickable"
          onClick={() => setView('tunnels')}
          disabled={isLoading}
        >
          取消
        </button>
      </div>
    </div>
  );

  // 渲染编辑表单
  const renderEditForm = () => (
    <div className="glass-light p-6 rounded-xl">
      <h2 className="text-xl font-semibold text-white mb-6">编辑隧道: {editingTunnel?.name}</h2>

      <div className="space-y-4">
        {/* 基本信息 */}
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-white/70 mb-2">隧道名称</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">本地地址</label>
              <input
                type="text"
                value={formData.localHost}
                onChange={(e) => setFormData({ ...formData, localHost: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">本地端口</label>
              <input
                type="number"
                value={formData.localPort}
                onChange={(e) => setFormData({ ...formData, localPort: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
                min="1"
                max="65535"
              />
            </div>
          </div>
        </div>

        {/* FRP 专用配置 */}
        <div className="space-y-3 pt-4 border-t border-white/10">
          <h3 className="text-lg font-medium text-white mb-2">FRP 服务器配置</h3>

            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">代理类型</label>
              <Select
                value={formData.proxyType}
                onValueChange={(value) => setFormData({ ...formData, proxyType: value as ProxyType })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择代理类型" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ProxyType.ProxyTypeHTTP}>HTTP</SelectItem>
                  <SelectItem value={ProxyType.ProxyTypeHTTPS}>HTTPS</SelectItem>
                  <SelectItem value={ProxyType.ProxyTypeTCP}>TCP</SelectItem>
                  <SelectItem value={ProxyType.ProxyTypeSTCP}>STCP (秘密 TCP)</SelectItem>
                  <SelectItem value={ProxyType.ProxyTypeXTCP}>XTCP (P2P TCP)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">子域名 (可选)</label>
              <input
                type="text"
                value={formData.subdomain}
                onChange={(e) => setFormData({ ...formData, subdomain: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
                placeholder="例如: myapp"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">服务器地址</label>
              <input
                type="text"
                value={formData.frpServerAddress}
                onChange={(e) => setFormData({ ...formData, frpServerAddress: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-white/70 mb-2">认证 Token</label>
              <input
                type="password"
                value={formData.frpServerToken}
                onChange={(e) => setFormData({ ...formData, frpServerToken: e.target.value })}
                className="w-full px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]"
              />
            </div>
        </div>

        {/* 其他选项 */}
        <div className="space-y-3 pt-4 border-t border-white/10">
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={formData.autoStart}
              onChange={(e) => setFormData({ ...formData, autoStart: e.target.checked })}
              className="w-5 h-5 rounded border-white/20 accent-[#7C3AED]"
            />
            <span className="text-sm text-white/70">应用启动时自动启动此隧道</span>
          </label>
        </div>
      </div>

      {/* 按钮组 */}
      <div className="flex gap-3 mt-6 pt-6 border-t border-white/10">
        <button
          className="px-6 py-2 bg-[#7C3AED] hover:bg-[#6D28D9] rounded-lg text-white font-medium clickable disabled:opacity-50"
          onClick={handleUpdateTunnel}
          disabled={isLoading}
        >
          {isLoading ? '保存中...' : '保存更改'}
        </button>
        <button
          className="px-6 py-2 bg-white/10 hover:bg-white/20 rounded-lg text-white font-medium clickable"
          onClick={() => setView('tunnels')}
          disabled={isLoading}
        >
          取消
        </button>
      </div>
    </div>
  );

  // 日志查看弹窗
  const renderLogModal = () => {
    if (!showLogModal) return null;

    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
        <div className="glass-heavy w-full max-w-4xl max-h-[85vh] rounded-xl overflow-hidden flex flex-col">
          <div className="flex items-center justify-between p-4 border-b border-white/10">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Icon name="terminal" size={18} />
              {currentLogTitle}
            </h3>
            <div className="flex items-center gap-2">
              <button
                className="p-2 rounded-lg hover:bg-white/10 clickable text-white/70"
                onClick={() => loadLogContent(currentLogTunnelId, 100)}
                title="手动刷新"
              >
                <Icon name="refresh" size={16} />
              </button>
              <button
                className="p-2 rounded-lg hover:bg-white/10 clickable"
                onClick={handleCloseLogModal}
              >
                <Icon name="close" size={18} />
              </button>
            </div>
          </div>

          <div className="flex items-center justify-between px-4 py-2 border-b border-white/10 bg-white/5">
            <div className="flex items-center gap-2">
              <span className="text-xs text-white/50">显示最后 100 行</span>
              {autoRefreshLog && (
                <span className="flex items-center gap-1 text-xs text-[#22C55E]">
                  <span className="w-1.5 h-1.5 rounded-full bg-[#22C55E] animate-pulse"></span>
                  实时刷新中
                </span>
              )}
            </div>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={autoRefreshLog}
                onChange={(e) => setAutoRefreshLog(e.target.checked)}
                className="w-4 h-4 rounded border-white/20 accent-[#7C3AED]"
              />
              <span className="text-sm text-white/70">自动刷新</span>
            </label>
          </div>

          <div className="flex-1 overflow-auto p-4 bg-black/30">
            <pre className="text-sm text-white/80 font-mono whitespace-pre-wrap leading-relaxed">{currentLog || '(暂无日志)'}</pre>
          </div>

          <div className="p-4 border-t border-white/10 flex justify-end gap-2">
            <button
              className="px-4 py-2 bg-white/10 hover:bg-white/20 rounded-lg text-white clickable flex items-center gap-2"
              onClick={() => loadLogContent(currentLogTunnelId, 100)}
            >
              <Icon name="refresh" size={14} />
              刷新
            </button>
            <button
              className="px-4 py-2 bg-[#7C3AED] hover:bg-[#6D28D9] rounded-lg text-white clickable"
              onClick={handleCloseLogModal}
            >
              关闭
            </button>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-[#0D0F1A] to-[#1A1D2E] p-6">
      {renderLogModal()}

      <button
        className="mb-6 flex items-center gap-2 text-white/60 hover:text-white clickable"
        onClick={onBack}
      >
        <Icon name="arrow-left" size={16} />
        <span>返回</span>
      </button>

      <div className="flex gap-4 mb-6">
        <button
          className={`px-4 py-2 rounded-lg font-medium transition-all ${
            view === 'tunnels'
              ? 'bg-[#7C3AED] text-white'
              : 'bg-white/10 text-white/60 hover:bg-white/20'
          }`}
          onClick={() => setView('tunnels')}
        >
          隧道列表
        </button>
        <button
          className={`px-4 py-2 rounded-lg font-medium transition-all ${
            view === 'create' || view === 'edit'
              ? 'bg-[#7C3AED] text-white'
              : 'bg-white/10 text-white/60 hover:bg-white/20'
          }`}
          onClick={handleSwitchToCreate}
        >
          创建隧道
        </button>
        <button
          className={`px-4 py-2 rounded-lg font-medium transition-all ${
            view === 'settings'
              ? 'bg-[#7C3AED] text-white'
              : 'bg-white/10 text-white/60 hover:bg-white/20'
          }`}
          onClick={() => setView('settings')}
        >
          设置
        </button>
      </div>

      {view === 'tunnels' && renderTunnelsView()}
      {view === 'create' && renderCreateForm()}
      {view === 'edit' && editingTunnel && renderEditForm()}
      {view === 'settings' && renderSettingsView()}
    </div>
  );
}
