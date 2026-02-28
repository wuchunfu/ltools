import React, { useState, useEffect } from 'react';
import { useBookmarks, CacheStatus } from '../hooks/useBookmarks';

export const BookmarkPage: React.FC = () => {
  const { search, sync, getCacheStatus, searching, syncing, error } = useBookmarks();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<any[]>([]);
  const [cacheStatus, setCacheStatus] = useState<CacheStatus | null>(null);
  const [browserFilter, setBrowserFilter] = useState<string>('all');

  // 加载缓存状态
  useEffect(() => {
    loadCacheStatus();
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

  // 获取浏览器图标
  const getBrowserIcon = (browser: string) => {
    const icons: Record<string, string> = {
      chrome: '🌐',
      safari: '🧭',
      firefox: '🦊',
    };
    return icons[browser] || '🔖';
  };

  // 过滤结果
  const filteredResults = browserFilter === 'all'
    ? results
    : results.filter(r => r.bookmark.browser === browserFilter);

  return (
    <div className="h-full flex flex-col">
      {/* 头部 */}
      <div className="p-4 border-b border-white/10">
        <h1 className="text-2xl font-bold text-white mb-4">书签管理</h1>

        {/* 搜索栏 */}
        <div className="flex gap-2 mb-4">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="搜索书签..."
            className="flex-1 px-4 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:border-purple-500"
          />
          <select
            value={browserFilter}
            onChange={(e) => setBrowserFilter(e.target.value)}
            className="px-4 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:border-purple-500"
          >
            <option value="all">全部浏览器</option>
            <option value="chrome">Chrome</option>
            <option value="safari">Safari</option>
            <option value="firefox">Firefox</option>
          </select>
        </div>

        {/* 操作区 */}
        <div className="flex items-center justify-between">
          <div className="flex gap-2">
            <button
              onClick={handleSync}
              disabled={syncing}
              className="px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:opacity-50 text-white rounded-lg transition-colors flex items-center gap-2"
            >
              {syncing ? (
                <>
                  <span className="animate-spin">⏳</span>
                  同步中...
                </>
              ) : (
                <>
                  🔄 同步书签
                </>
              )}
            </button>
          </div>

          {/* 缓存状态 */}
          {cacheStatus && (
            <div className="text-sm text-gray-400">
              <span className="mr-4">
                总书签: <span className="text-white">{cacheStatus.total_count}</span>
              </span>
              <span className="mr-4">
                上次同步: <span className="text-white">{cacheStatus.last_sync || '从未'}</span>
              </span>
              {cacheStatus.is_expired && (
                <span className="text-yellow-400">⚠️ 缓存已过期</span>
              )}
            </div>
          )}
        </div>

        {/* 浏览器统计 */}
        {cacheStatus?.browser_stats && (
          <div className="flex gap-4 mt-3 text-sm">
            {Object.entries(cacheStatus.browser_stats).map(([browser, count]) => (
              <span key={browser} className="text-gray-400">
                {getBrowserIcon(browser)} {browser}: <span className="text-white">{count}</span>
              </span>
            ))}
          </div>
        )}
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="mx-4 mt-4 p-3 bg-red-500/20 border border-red-500/50 rounded-lg text-red-300">
          {error}
        </div>
      )}

      {/* 搜索结果 */}
      <div className="flex-1 overflow-y-auto p-4">
        {searching ? (
          <div className="text-center text-gray-400 py-8">
            <span className="animate-spin inline-block">⏳</span> 搜索中...
          </div>
        ) : query && filteredResults.length === 0 ? (
          <div className="text-center text-gray-400 py-8">
            未找到匹配的书签
          </div>
        ) : (
          <div className="space-y-2">
            {filteredResults.map((result) => (
              <div
                key={result.bookmark.id}
                onClick={() => window.open(result.bookmark.url, '_blank')}
                className="p-3 bg-white/5 hover:bg-white/10 rounded-lg cursor-pointer transition-colors"
              >
                <div className="flex items-start gap-3">
                  <span className="text-xl flex-shrink-0">
                    {getBrowserIcon(result.bookmark.browser)}
                  </span>
                  <div className="flex-1 min-w-0">
                    <div className="text-white font-medium truncate">
                      {result.bookmark.title}
                    </div>
                    <div className="text-gray-400 text-sm truncate">
                      {result.bookmark.url}
                    </div>
                    {result.bookmark.folder && (
                      <div className="text-gray-500 text-xs mt-1 truncate">
                        📁 {result.bookmark.folder}
                      </div>
                    )}
                  </div>
                  <div className="text-xs text-gray-500 flex-shrink-0">
                    {result.match_type}
                    {result.score > 0 && <span className="ml-1">({result.score})</span>}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
