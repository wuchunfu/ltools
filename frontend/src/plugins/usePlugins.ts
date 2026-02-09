import { useState, useEffect, useCallback } from 'react';
import { Events } from '@wailsio/runtime';
import { PluginService, PluginMetadata, Permission } from '../../bindings/ltools/internal/plugins';

// 定义插件状态变化事件名称
export const PLUGIN_STATE_CHANGED_EVENT = 'plugin:state-changed';

/**
 * 插件管理 Hook
 * 提供插件列表、加载、启用/禁用等功能
 */
export function usePlugins() {
  const [plugins, setPlugins] = useState<PluginMetadata[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  // 加载插件列表
  const loadPlugins = useCallback(async () => {
    try {
      setLoading(true);
      const pluginList = await PluginService.List();
      // Filter out null values
      setPlugins(pluginList.filter((p): p is PluginMetadata => p !== null));
    } catch (err) {
      setError(err as Error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 启用插件
  const enablePlugin = useCallback(async (id: string) => {
    try {
      await PluginService.Enable(id);
      await loadPlugins(); // 重新加载列表
      // 发送状态变化事件
      Events.Emit(PLUGIN_STATE_CHANGED_EVENT, { action: 'enable', pluginId: id });
    } catch (err) {
      setError(err as Error);
      throw err; // 重新抛出错误以便调用者处理
    }
  }, [loadPlugins]);

  // 禁用插件
  const disablePlugin = useCallback(async (id: string) => {
    try {
      await PluginService.Disable(id);
      await loadPlugins(); // 重新加载列表
      // 发送状态变化事件
      Events.Emit(PLUGIN_STATE_CHANGED_EVENT, { action: 'disable', pluginId: id });
    } catch (err) {
      setError(err as Error);
      throw err; // 重新抛出错误以便调用者处理
    }
  }, [loadPlugins]);

  // 搜索插件
  const searchPlugins = async (keywords: string[]) => {
    try {
      const results = await PluginService.Search(...keywords);
      return results;
    } catch (err) {
      setError(err as Error);
      return [];
    }
  };

  // 获取插件详情
  const getPlugin = async (id: string) => {
    try {
      return await PluginService.Get(id);
    } catch (err) {
      setError(err as Error);
      return null;
    }
  };

  // 检查权限
  const checkPermission = async (pluginId: string, permission: Permission) => {
    try {
      return await PluginService.CheckPermission(pluginId, permission);
    } catch (err) {
      setError(err as Error);
      return false;
    }
  };

  // 请求权限
  const requestPermission = async (
    pluginId: string,
    permission: Permission,
    granted: boolean
  ) => {
    try {
      await PluginService.RequestPermission(pluginId, permission, granted);
    } catch (err) {
      setError(err as Error);
    }
  };

  // 初始化时加载插件
  useEffect(() => {
    loadPlugins();

    // 监听插件状态变化事件，当其他组件修改插件状态时同步更新
    const unsubscribe = Events.On(PLUGIN_STATE_CHANGED_EVENT, () => {
      loadPlugins();
    });

    return () => {
      unsubscribe?.();
    };
  }, [loadPlugins]);

  return {
    plugins,
    loading,
    error,
    loadPlugins,
    enablePlugin,
    disablePlugin,
    searchPlugins,
    getPlugin,
    checkPermission,
    requestPermission,
  };
}

/**
 * 单个插件 Hook
 * @param id 插件ID
 */
export function usePlugin(id: string) {
  const [plugin, setPlugin] = useState<PluginMetadata | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const loadPlugin = async () => {
      try {
        setLoading(true);
        const data = await PluginService.Get(id);
        setPlugin(data);
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    };

    loadPlugin();
  }, [id]);

  return { plugin, loading, error };
}


/**
 * 日期时间插件 Hook
 * 专门用于日期时间插件的功能
 */
export function useDateTime() {
  const [currentTime, setCurrentTime] = useState<string>('');
  const [currentDate, setCurrentDate] = useState<string>('');
  const [weekday, setWeekday] = useState<string>('');

  // 监听日期时间事件
  useEffect(() => {
    const unsubTime = Events.On('datetime:time', (ev: { data: string }) => {
      setCurrentTime(ev.data);
    });

    const unsubDate = Events.On('datetime:date', (ev: { data: string }) => {
      setCurrentDate(ev.data);
    });

    const unsubDay = Events.On('datetime:weekday', (ev: { data: string }) => {
      setWeekday(ev.data);
    });

    return () => {
      unsubTime?.();
      unsubDate?.();
      unsubDay?.();
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return {
    currentTime,
    currentDate,
    weekday,
  };
}
