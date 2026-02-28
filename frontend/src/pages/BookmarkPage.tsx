import React, { useState, useEffect, useRef } from 'react';
import { useBookmarks, CacheStatus } from '../hooks/useBookmarks';
import { Icon } from '../components/Icon';

export const BookmarkPage: React.FC = () => {
  const { search, sync, getCacheStatus, openURL, exportHTML, exportJSON, searching, syncing, error } = useBookmarks();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<any[]>([]);
  const [cacheStatus, setCacheStatus] = useState<CacheStatus | null>(null);
  const [browserFilter, setBrowserFilter] = useState<string>('all');
  const [showExportMenu, setShowExportMenu] = useState(false);
  const exportMenuRef = useRef<HTMLDivElement>(null);
  const [selectedIndex, setSelectedIndex] = useState(0);

  // 加载缓存状态
  useEffect(() => {
    loadCacheStatus();
  }, []);

  // 点击外部关闭导出菜单
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (exportMenuRef.current && !exportMenuRef.current.contains(event.target as Node)) {
        setShowExportMenu(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const loadCacheStatus = async () => {
    const status = await getCacheStatus();
    setCacheStatus(status);
  };

  // 搜索书签
  useEffect(() => {
    const doSearch = async () => {
      if (!query.trim()) {
        setResults([]);
        return;
      }
      const searchResults = await search(query);
      setResults(searchResults);
      setSelectedIndex(0);
    };

    const debounce = setTimeout(doSearch, 200);
    return () => clearTimeout(debounce);
  }, [query, search]);

  // 手动同步
  const handleSync = async () => {
    const success = await sync();
    if (success) {
      await loadCacheStatus();
    }
  };

  // 打开书签
  const handleOpenBookmark = async (url: string) => {
    await openURL(url);
  };

  // 导出为 HTML
  const handleExportHTML = async () => {
    const defaultPath = `~/Desktop/bookmarks_${new Date().toISOString().slice(0, 10)}.html`;
    const success = await exportHTML(defaultPath);
    if (success) {
      setShowExportMenu(false);
    }
  };

  // 导出为 JSON
  const handleExportJSON = async () => {
    const defaultPath = `~/Desktop/bookmarks_${new Date().toISOString().slice(0, 10)}.json`;
    const success = await exportJSON(defaultPath);
    if (success) {
      setShowExportMenu(false);
    }
  };

  // 过滤结果 - 需要在键盘导航 useEffect 之前定义
  const filteredResults = browserFilter === 'all'
    ? results
    : results.filter(r => r.bookmark.browser === browserFilter);

  // 键盘导航
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!query.trim() || filteredResults.length === 0) return;

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIndex(prev => Math.min(prev + 1, filteredResults.length - 1));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex(prev => Math.max(prev - 1, 0));
      } else if (e.key === 'Enter') {
        e.preventDefault();
        const selected = filteredResults[selectedIndex];
        if (selected) {
          handleOpenBookmark(selected.bookmark.url);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [query, filteredResults, selectedIndex]);

  // 获取浏览器图标
  const getBrowserIcon = (browser: string) => {
    const icons: Record<string, string> = {
      chrome: '🌐',
      safari: '🧭',
      firefox: '🦊',
    };
    return icons[browser] || '🔖';
  };

  // 获取浏览器名称
  const getBrowserName = (browser: string) => {
    const names: Record<string, string> = {
      chrome: 'Chrome',
      safari: 'Safari',
      firefox: 'Firefox',
    };
    return names[browser] || browser;
  };

  return (
    <div className="h-full flex flex-col bg-[#0D0F1A]">
      {/* 头部 */}
      <div className="p-6 border-b border-white/10">
        <div className="max-w-4xl mx-auto">
          {/* 标题 */}
          <div className="flex items-center gap-4 mb-6">
            <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] flex items-center justify-center">
              <Icon name="bookmark" size={28} color="white" />
            </div>
            <div>
              <h1 className="text-3xl font-bold mb-1">书签管理</h1>
              <p className="text-white/50">搜索和管理浏览器书签</p>
              <p className="text-sm text-white/30 mt-1">v1.0.0 · by LTools</p>
            </div>
          </div>

          {/* 搜索栏 */}
          <div className="flex gap-3 mb-4">
            <div className="flex-1 relative">
              <Icon name="search" size={18} color="rgba(255,255,255,0.4)" className="absolute left-4 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="搜索书签标题、URL 或拼音..."
                className="w-full pl-11 pr-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 focus:outline-none focus:border-[#7C3AED]/50 focus:bg-white/10 transition-all"
              />
            </div>
            <select
              value={browserFilter}
              onChange={(e) => setBrowserFilter(e.target.value)}
              className="px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white focus:outline-none focus:border-[#7C3AED]/50 cursor-pointer min-w-[140px]"
            >
              <option value="all" className="bg-[#1a1a2e]">全部浏览器</option>
              <option value="chrome" className="bg-[#1a1a2e]">Chrome</option>
              <option value="safari" className="bg-[#1a1a2e]">Safari</option>
              <option value="firefox" className="bg-[#1a1a2e]">Firefox</option>
            </select>
          </div>

          {/* 操作区 */}
          <div className="flex items-center justify-between">
            <div className="flex gap-3">
              <button
                onClick={handleSync}
                disabled={syncing}
                className="px-5 py-2.5 bg-[#7C3AED] hover:bg-[#6D28D9] disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg transition-colors flex items-center gap-2 font-medium"
              >
                {syncing ? (
                  <>
                    <span className="animate-spin inline-block">⏳</span>
                    同步中...
                  </>
                ) : (
                  <>
                    <Icon name="refresh" size={16} color="white" />
                    同步书签
                  </>
                )}
              </button>

              {/* 导出按钮 */}
              <div className="relative" ref={exportMenuRef}>
                <button
                  onClick={() => setShowExportMenu(!showExportMenu)}
                  className="px-5 py-2.5 bg-white/5 hover:bg-white/10 border border-white/10 text-white rounded-lg transition-colors flex items-center gap-2 font-medium"
                >
                  <Icon name="download" size={16} color="white" />
                  导出
                  <Icon name="chevron-down" size={14} color="white" />
                </button>

                {showExportMenu && (
                  <div className="absolute top-full left-0 mt-2 w-48 glass-light rounded-lg border border-white/10 overflow-hidden z-10">
                    <button
                      onClick={handleExportHTML}
                      className="w-full px-4 py-3 text-left text-white hover:bg-white/10 transition-colors flex items-center gap-3"
                    >
                      <span className="text-lg">📄</span>
                      <div>
                        <div className="font-medium">导出为 HTML</div>
                        <div className="text-xs text-white/50">Netscape 格式</div>
                      </div>
                    </button>
                    <button
                      onClick={handleExportJSON}
                      className="w-full px-4 py-3 text-left text-white hover:bg-white/10 transition-colors flex items-center gap-3 border-t border-white/10"
                    >
                      <span className="text-lg">📋</span>
                      <div>
                        <div className="font-medium">导出为 JSON</div>
                        <div className="text-xs text-white/50">结构化数据</div>
                      </div>
                    </button>
                  </div>
                )}
              </div>
            </div>

            {/* 缓存状态 */}
            {cacheStatus && (
              <div className="flex items-center gap-4 text-sm text-white/60">
                <span>
                  总书签: <span className="text-white font-medium">{cacheStatus.total_count}</span>
                </span>
                <span>
                  上次同步: <span className="text-white">{cacheStatus.last_sync || '从未'}</span>
                </span>
                {cacheStatus.is_expired && (
                  <span className="text-yellow-400 flex items-center gap-1">
                    <Icon name="exclamation-circle" size={14} color="#F59E0B" />
                    缓存已过期
                  </span>
                )}
              </div>
            )}
          </div>

          {/* 浏览器统计 */}
          {cacheStatus?.browser_stats && Object.keys(cacheStatus.browser_stats).length > 0 && (
            <div className="flex gap-4 mt-4">
              {Object.entries(cacheStatus.browser_stats).map(([browser, count]) => (
                <div
                  key={browser}
                  className="px-4 py-2 glass-light rounded-lg border border-white/5 flex items-center gap-2"
                >
                  <span className="text-lg">{getBrowserIcon(browser)}</span>
                  <span className="text-white/60">{getBrowserName(browser)}</span>
                  <span className="text-white font-medium">{count}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="mx-6 mt-4 max-w-4xl lg:mx-auto">
          <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-xl text-red-300 flex items-center gap-3">
            <Icon name="exclamation-circle" size={20} color="#EF4444" />
            {error}
          </div>
        </div>
      )}

      {/* 搜索结果 */}
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-4xl mx-auto">
          {searching ? (
            <div className="text-center text-white/40 py-12">
              <span className="animate-spin inline-block text-2xl mb-3">⏳</span>
              <p>搜索中...</p>
            </div>
          ) : query && filteredResults.length === 0 ? (
            <div className="text-center text-white/40 py-12">
              <Icon name="search" size={48} color="rgba(255,255,255,0.2)" className="mb-4" />
              <p>未找到匹配的书签</p>
              <p className="text-sm text-white/30 mt-2">尝试其他关键词或同步书签</p>
            </div>
          ) : !query ? (
            <div className="text-center text-white/40 py-12">
              <Icon name="bookmark" size={48} color="rgba(255,255,255,0.2)" className="mb-4" />
              <p>输入关键词搜索书签</p>
              <p className="text-sm text-white/30 mt-2">支持标题、URL 和拼音搜索</p>
            </div>
          ) : (
            <div className="space-y-2">
              {filteredResults.map((result, index) => (
                <div
                  key={result.bookmark.id}
                  onClick={() => handleOpenBookmark(result.bookmark.url)}
                  className={`p-4 rounded-xl cursor-pointer transition-all ${
                    index === selectedIndex
                      ? 'bg-[#7C3AED]/20 border border-[#7C3AED]/30'
                      : 'bg-white/5 hover:bg-white/10 border border-transparent'
                  }`}
                >
                  <div className="flex items-start gap-4">
                    <span className="text-2xl flex-shrink-0 mt-0.5">
                      {getBrowserIcon(result.bookmark.browser)}
                    </span>
                    <div className="flex-1 min-w-0">
                      <div className="text-white font-medium truncate mb-1">
                        {result.bookmark.title}
                      </div>
                      <div className="text-white/50 text-sm truncate mb-1">
                        {result.bookmark.url}
                      </div>
                      <div className="flex items-center gap-3 text-xs text-white/40">
                        {result.bookmark.folder && (
                          <span className="flex items-center gap-1">
                            <Icon name="folder" size={12} color="rgba(255,255,255,0.4)" />
                            {result.bookmark.folder}
                          </span>
                        )}
                        <span className="px-2 py-0.5 rounded bg-white/5">
                          {result.match_type}
                        </span>
                        {result.score > 0 && (
                          <span className="text-[#7C3AED]">相关度: {result.score}</span>
                        )}
                      </div>
                    </div>
                    <div className="flex-shrink-0 text-white/30">
                      <Icon name="arrow-right" size={16} color="rgba(255,255,255,0.3)" />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
