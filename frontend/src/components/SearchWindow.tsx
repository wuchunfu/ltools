import { useState, useEffect, useCallback, useRef } from 'react';
import { Icon } from './Icon';
import { Events } from '@wailsio/runtime';
import * as SearchWindowService from '../../bindings/ltools/internal/plugins/searchwindowservice';
import * as AppLauncherService from '../../bindings/ltools/plugins/applauncher/applauncherservice';
import { usePlugins } from '../plugins/usePlugins';
import { getPluginIcon } from '../utils/pluginHelpers';
import { PluginState } from '../../bindings/ltools/internal/plugins';
import './SearchWindow.css';

/**
 * 解码 Unicode 转义字符
 * 将 \uXXXX 转换为实际字符
 */
function decodeUnicode(str: string): string {
  return str.replace(/\\u[\dA-Fa-f]{4}/g, (match) => {
    const code = parseInt(match.slice(2), 16);
    return String.fromCharCode(code);
  });
}

/**
 * 搜索结果接口
 */
interface SearchResult {
  pluginId?: string;
  appId?: string;
  name: string;
  description: string;
  icon: string;
  matchedFields?: string[];
  type: string; // "plugin" or "app"
}

/**
 * 高亮搜索匹配文本
 */
function highlightMatch(text: string, query: string): JSX.Element {
  if (!query) return <>{text}</>;

  const regex = new RegExp(`(${query})`, 'gi');
  const parts = text.split(regex);

  return (
    <>
      {parts.map((part, index) =>
        regex.test(part) ? (
          <mark key={index} className="bg-[#7C3AED]/30 text-[#A78BFA] rounded px-0.5">
            {part}
          </mark>
        ) : (
          <span key={index}>{part}</span>
        )
      )}
    </>
  );
}

/**
 * 搜索窗口主组件
 */
export function SearchWindow() {
  // 窗口固定大小，使用固定的每页显示数量
  const ITEMS_PER_PAGE = 12; // 4x3 网格布局

  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [loading, setLoading] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // 分页状态
  const [currentPage, setCurrentPage] = useState(0);

  // 获取插件列表
  const { plugins } = usePlugins();

  // 过滤已启用的插件
  const enabledPlugins = plugins.filter(p =>
    p.state === PluginState.PluginStateEnabled
  );

  // 计算总页数
  const totalPages = Math.ceil(enabledPlugins.length / ITEMS_PER_PAGE);

  // 自动聚焦输入框
  useEffect(() => {
    const unsubscribeOpened = Events.On('search:opened', () => {
      console.log('[SearchWindow] Search opened event received');
      setTimeout(() => {
        inputRef.current?.focus();
        setQuery('');
        setResults([]);
        setSelectedIndex(0);
        setCurrentPage(0); // 重置到第一页
      }, 100);
    });

    const unsubscribeClosed = Events.On('search:closed', () => {
      console.log('[SearchWindow] Search closed event received');
      setQuery('');
      setResults([]);
      setSelectedIndex(0);
      setCurrentPage(0); // 重置到第一页
    });

    // Initial focus
    inputRef.current?.focus();

    return () => {
      unsubscribeOpened();
      unsubscribeClosed();
    };
  }, []);

  // 滚轮切换页面
  useEffect(() => {
    const handleWheel = (e: Event) => {
      // 只有在没有搜索输入时才启用滚轮翻页
      if (query || loading || enabledPlugins.length === 0) return;

      const wheelEvent = e as WheelEvent;
      // 防抖处理
      e.preventDefault();

      // 垂直滚轮或水平滚轮都可以切换页面
      const delta = wheelEvent.deltaY !== 0 ? wheelEvent.deltaY : wheelEvent.deltaX;

      if (delta > 0) {
        // 向下/向右滚动，下一页
        setCurrentPage(prev => Math.min(prev + 1, totalPages - 1));
      } else if (delta < 0) {
        // 向上/向左滚动，上一页
        setCurrentPage(prev => Math.max(prev - 1, 0));
      }
    };

    const searchResultsEl = document.querySelector('.search-results');
    searchResultsEl?.addEventListener('wheel', handleWheel, { passive: false });

    return () => {
      searchResultsEl?.removeEventListener('wheel', handleWheel);
    };
  }, [query, loading, enabledPlugins.length, totalPages]);

  // 页面切换函数
  const goToPage = useCallback((page: number) => {
    if (page >= 0 && page < totalPages) {
      setCurrentPage(page);
    }
  }, [totalPages]);

  // 搜索功能（防抖 150ms）
  const performSearch = useCallback(async (searchQuery: string) => {
    if (!searchQuery || searchQuery.trim() === '') {
      setResults([]);
      setSelectedIndex(0);
      return;
    }

    setLoading(true);
    try {
      console.log('[SearchWindow] Searching for:', searchQuery);

      // 同时调用插件搜索和应用搜索
      const [pluginResults, appResults] = await Promise.all([
        SearchWindowService.Search(searchQuery),
        AppLauncherService.Search(searchQuery).catch(() => [])
      ]);

      console.log('[SearchWindow] Plugin results:', pluginResults);
      console.log('[SearchWindow] App results:', appResults);

      // 转换插件结果格式，并解码 Unicode 转义字符
      const searchResults: SearchResult[] = pluginResults.map((item: any) => ({
        pluginId: item.pluginId,
        appId: item.appId,
        name: decodeUnicode(item.name || ''),
        description: decodeUnicode(item.description || ''),
        icon: item.icon || '',
        matchedFields: item.matchedFields || [],
        type: item.type || 'plugin',
      }));

      // 转换应用结果格式，并解码 Unicode 转义字符
      const appResultsFormatted: SearchResult[] = appResults.map((item: any) => {
        // 使用 iconData，如果没有则使用默认图标
        const icon = item.iconData || '🚀';
        console.log('[SearchWindow] App icon for', item.name, ':', icon.substring(0, 50) + '...');
        return {
          appId: item.id,
          name: decodeUnicode(item.name || ''),
          description: decodeUnicode(item.description || ''),
          icon: icon,
          type: 'app',
        };
      });

      // 合并结果
      const allResults = [...searchResults, ...appResultsFormatted];

      console.log('[SearchWindow] Total results:', allResults.length);
      setResults(allResults);
      setSelectedIndex(Math.min(selectedIndex, Math.max(0, allResults.length - 1)));
    } catch (error) {
      console.error('[SearchWindow] Search failed:', error);
      setResults([]);
    } finally {
      setLoading(false);
    }
  }, []);

  // 防抖搜索
  useEffect(() => {
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }

    searchTimeoutRef.current = setTimeout(() => {
      performSearch(query);
    }, 150);

    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, [query, performSearch]);

  // 键盘导航
  useEffect(() => {
    const handleKeyDown = async (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(prev => Math.min(prev + 1, results.length - 1));
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(prev => Math.max(prev - 1, 0));
          break;
        case 'ArrowLeft':
          // 在默认视图下，左箭头切换到上一页
          if (!query && !loading && enabledPlugins.length > 0) {
            e.preventDefault();
            setCurrentPage(prev => Math.max(prev - 1, 0));
          }
          break;
        case 'ArrowRight':
          // 在默认视图下，右箭头切换到下一页
          if (!query && !loading && enabledPlugins.length > 0) {
            e.preventDefault();
            setCurrentPage(prev => Math.min(prev + 1, totalPages - 1));
          }
          break;
        case 'Enter':
          e.preventDefault();
          if (results.length > 0 && selectedIndex >= 0 && selectedIndex < results.length) {
            await openItem(results[selectedIndex]);
          }
          break;
        case 'Escape':
          e.preventDefault();
          // 如果有搜索内容，清空搜索；如果为空，关闭窗口
          if (query.trim() !== '') {
            setQuery('');
            setResults([]);
            setSelectedIndex(0);
          } else {
            try {
              await SearchWindowService.Hide();
            } catch (error) {
              console.error('[SearchWindow] Failed to hide window:', error);
            }
          }
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [results, selectedIndex, query, loading, enabledPlugins.length, totalPages]);

  // 打开插件
  const openPlugin = async (pluginId: string) => {
    console.log('[SearchWindow] Opening plugin:', pluginId);
    try {
      await SearchWindowService.OpenPlugin(pluginId);
      // 窗口会在 OpenPlugin 后自动隐藏
    } catch (error) {
      console.error('[SearchWindow] Failed to open plugin:', error);
    }
  };

  // 打开应用
  const openApp = async (appId: string) => {
    console.log('[SearchWindow] Opening app:', appId);
    try {
      await SearchWindowService.OpenApp(appId);
      // 窗口会在 OpenApp 后自动隐藏
    } catch (error) {
      console.error('[SearchWindow] Failed to open app:', error);
    }
  };

  // 打开结果项（插件或应用）
  const openItem = async (result: SearchResult) => {
    if (result.type === 'app' && result.appId) {
      await openApp(result.appId);
    } else if (result.type === 'plugin' && result.pluginId) {
      await openPlugin(result.pluginId);
    }
  };

  // 点击结果项
  const handleResultClick = (result: SearchResult) => {
    openItem(result);
  };

  // 获取匹配字段名称
  const getMatchedFieldLabel = (field: string): string => {
    const labels: Record<string, string> = {
      name: '名称',
      description: '描述',
      keyword: '关键词',
      author: '作者',
    };
    return labels[field] || field;
  };

  return (
    <div className="search-window">
      {/* 窗口头部 - 可拖动区域 */}
      <div className="search-window-header" data-wails-draggable>
        <div className="search-header-content">
          <Icon name="search" size={18} color="#A78BFA" />
          <input
            ref={inputRef}
            type="text"
            className="search-input"
            placeholder="搜索插件..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            data-wails-drag-draggable="false"
          />
          <button
            className="close-button"
            onClick={() => {
              setQuery('');
              setResults([]);
              setSelectedIndex(0);
              inputRef.current?.focus();
            }}
            data-wails-drag-draggable="false"
            title="清空搜索"
          >
            <Icon name="x-circle" size={18} color="rgba(255,255,255,0.5)" />
          </button>
        </div>
      </div>

      {/* 搜索结果列表 */}
      <div className="search-results">
        {loading && (
          <div className="search-loading">
            <div className="loading-spinner" />
            <p className="text-white/50 text-sm">搜索中...</p>
          </div>
        )}

        {/* 默认插件卡片视图 - 分页模式 */}
        {!query && !loading && enabledPlugins.length > 0 && (
          <div className="plugin-launchpad">
            <div className="plugin-pages-container">
              <div
                className="plugin-pages-slider"
                style={{
                  transform: `translateX(-${currentPage * 100}%)`,
                  transition: 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                }}
              >
                {Array.from({ length: totalPages }).map((_, pageIndex) => (
                  <div key={pageIndex} className="plugin-page">
                    <div className="plugin-grid">
                      {enabledPlugins
                        .slice(pageIndex * ITEMS_PER_PAGE, (pageIndex + 1) * ITEMS_PER_PAGE)
                        .map((plugin) => (
                          <div
                            key={plugin.id}
                            className="plugin-card"
                            onClick={() => openPlugin(plugin.id)}
                          >
                            <div className="plugin-card-icon">
                              {getPluginIcon(plugin)}
                            </div>
                            <div className="plugin-card-name">{plugin.name}</div>
                          </div>
                        ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* 页面指示器 */}
            {totalPages > 1 && (
              <div className="page-indicators">
                {Array.from({ length: totalPages }).map((_, index) => (
                  <button
                    key={index}
                    className={`page-dot ${index === currentPage ? 'page-dot-active' : ''}`}
                    onClick={() => goToPage(index)}
                    aria-label={`Page ${index + 1}`}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {!loading && query && results.length === 0 && (
          <div className="search-empty">
            <Icon name="search" size={48} color="rgba(167, 139, 250, 0.2)" />
            <p className="text-white/40 mt-3">未找到匹配的结果</p>
            <p className="text-white/30 text-sm mt-1">尝试其他关键词</p>
          </div>
        )}

        {!loading && query && results.length > 0 && (
          <div className="results-list">
            {results.map((result, index) => (
              <div
                key={result.pluginId || result.appId}
                className={`result-item ${
                  index === selectedIndex ? 'result-item-selected' : ''
                }`}
                onClick={() => handleResultClick(result)}
                onMouseEnter={() => setSelectedIndex(index)}
              >
                <div className="result-icon">
                  {result.type === 'app' && typeof result.icon === 'string' && result.icon.startsWith('data:') ? (
                    <img src={result.icon} alt={result.name} className="w-8 h-8 rounded" />
                  ) : (
                    result.icon
                  )}
                </div>
                <div className="result-content">
                  <div className="result-name">
                    {highlightMatch(result.name, query)}
                  </div>
                  <div className="result-description">
                    {highlightMatch(result.description, query)}
                  </div>
                  {result.matchedFields && result.matchedFields.length > 0 && (
                    <div className="result-match-info">
                      {result.matchedFields.map(field => (
                        <span key={field} className="match-badge">
                          {getMatchedFieldLabel(field)}
                        </span>
                      ))}
                    </div>
                  )}
                  <div className="result-type-badge">
                    {result.type === 'app' ? '应用' : '插件'}
                  </div>
                </div>
                {index === selectedIndex && (
                  <div className="result-indicator">
                    <Icon name="chevron-right" size={16} color="#A78BFA" />
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 状态栏 */}
      {results.length > 0 && (
        <div className="search-statusbar">
          <div className="statusbar-info">
            <span className="statusbar-count">
              找到 {results.length} 个插件
            </span>
          </div>
          <div className="statusbar-shortcuts">
            <span className="shortcut-hint-inline">
              ↑↓ 导航 • Enter 打开 • Esc 关闭
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
