import { useState, useCallback } from 'react';
import { PluginService, Permission } from '../../bindings/ltools/internal/plugins';

/**
 * 权限请求状态
 */
export interface PermissionRequest {
  pluginId: string;
  pluginName: string;
  permissions: Permission[];
}

/**
 * 权限管理 Hook
 * 提供权限检查、请求和授权功能
 */
export function usePermissions() {
  const [pendingRequest, setPendingRequest] = useState<PermissionRequest | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  /**
   * 检查插件是否有特定权限
   */
  const checkPermission = useCallback(
    async (pluginId: string, permission: Permission): Promise<boolean> => {
      try {
        return await PluginService.CheckPermission(pluginId, permission);
      } catch (error) {
        console.error(`Failed to check permission ${permission} for plugin ${pluginId}:`, error);
        return false;
      }
    },
    []
  );

  /**
   * 检查插件是否有多个权限
   */
  const checkPermissions = useCallback(
    async (pluginId: string, permissions: Permission[]): Promise<Record<Permission, boolean>> => {
      const results: Record<Permission, boolean> = {} as any;

      await Promise.all(
        permissions.map(async (permission) => {
          results[permission] = await checkPermission(pluginId, permission);
        })
      );

      return results;
    },
    [checkPermission]
  );

  /**
   * 获取插件的所有权限
   */
  const getPluginPermissions = useCallback(async (pluginId: string): Promise<Permission[]> => {
    try {
      return await PluginService.GetPermissions(pluginId);
    } catch (error) {
      console.error(`Failed to get permissions for plugin ${pluginId}:`, error);
      return [];
    }
  }, []);

  /**
   * 请求权限（显示对话框）
   */
  const requestPermissions = useCallback(
    (pluginId: string, pluginName: string, permissions: Permission[]) => {
      setPendingRequest({ pluginId, pluginName, permissions });
      setIsDialogOpen(true);
    },
    []
  );

  /**
   * 授予权限（用户在对话框中确认）
   */
  const grantPermissions = useCallback(async (permissions: Permission[]) => {
    if (!pendingRequest) return;

    try {
      for (const permission of permissions) {
        await PluginService.RequestPermission(
          pendingRequest.pluginId,
          permission,
          true
        );
      }
    } catch (error) {
      console.error('Failed to grant permissions:', error);
      throw error;
    } finally {
      setIsDialogOpen(false);
      setPendingRequest(null);
    }
  }, [pendingRequest]);

  /**
   * 拒绝权限（用户在对话框中拒绝）
   */
  const denyPermissions = useCallback(async () => {
    if (!pendingRequest) return;

    try {
      for (const permission of pendingRequest.permissions) {
        await PluginService.RequestPermission(
          pendingRequest.pluginId,
          permission,
          false
        );
      }
    } catch (error) {
      console.error('Failed to deny permissions:', error);
    } finally {
      setIsDialogOpen(false);
      setPendingRequest(null);
    }
  }, [pendingRequest]);

  /**
   * 获取所有可用的权限类型
   */
  const getAvailablePermissions = useCallback(async (): Promise<Permission[]> => {
    return await PluginService.GetAvailablePermissions();
  }, []);

  return {
    pendingRequest,
    isDialogOpen,
    checkPermission,
    checkPermissions,
    getPluginPermissions,
    requestPermissions,
    grantPermissions,
    denyPermissions,
    getAvailablePermissions,
    closeDialog: () => setIsDialogOpen(false),
  };
}

/**
 * 单个插件权限 Hook
 * @param pluginId 插件ID
 */
export function usePluginPermissions(pluginId: string) {
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const { checkPermission, getPluginPermissions } = usePermissions();

  const loadPermissions = useCallback(async () => {
    try {
      setLoading(true);
      const perms = await getPluginPermissions(pluginId);
      setPermissions(perms);
    } catch (err) {
      setError(err as Error);
    } finally {
      setLoading(false);
    }
  }, [pluginId, getPluginPermissions]);

  const hasPermission = useCallback(
    async (permission: Permission): Promise<boolean> => {
      return await checkPermission(pluginId, permission);
    },
    [pluginId, checkPermission]
  );

  return {
    permissions,
    loading,
    error,
    loadPermissions,
    hasPermission,
  };
}
