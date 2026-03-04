/**
 * MusicFree 适配器
 *
 * 将 MusicFree 格式的插件适配为 LX Music 格式
 */

const fs = require('fs');
const path = require('path');

class MusicFreeAdapter {
  constructor() {
    this.plugins = new Map();
    // 缓存原始 musicItem 数据（用于歌词/图片获取）
    // key: `${pluginId}:${songId}`, value: 原始 item
    this.musicItemCache = new Map();
    // 多音源缓存：同一首歌在不同音源的数据
    // key: songName:artist, value: Array<{pluginId, item, songId}>
    this.multiSourceCache = new Map();
    // 音源成功率统计
    // key: pluginId, value: {success: number, fail: number}
    this.sourceStats = new Map();
  }

  /**
   * 加载 MusicFree 插件
   */
  async loadPlugin(scriptPath) {
    try {
      const absolutePath = path.resolve(scriptPath);

      // 加载插件模块
      const plugin = require(absolutePath);

      if (!plugin || !plugin.platform) {
        throw new Error('Invalid MusicFree plugin format');
      }

      const pluginId = plugin.platform;
      this.plugins.set(pluginId, plugin);

      console.error(`[MusicFreeAdapter] Loaded plugin: ${pluginId} v${plugin.version}`);

      return pluginId;
    } catch (err) {
      console.error(`[MusicFreeAdapter] Failed to load plugin ${scriptPath}:`, err.message);
      throw err;
    }
  }

  /**
   * 转换为 LX Music 源信息
   */
  getLXSourceInfo(pluginId) {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) return null;

    return {
      name: plugin.platform,
      id: pluginId,
      type: 'music',
      actions: ['musicUrl', 'lyric', 'pic'],
      qualitys: ['128k', '320k', 'flac'],
    };
  }

  /**
   * 搜索（转换为 LX Music 格式）
   */
  async search(pluginId, keyword, page = 1, limit = 20) {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      throw new Error(`Plugin ${pluginId} not found`);
    }

    try {
      // 调用 MusicFree 的搜索方法
      const result = await plugin.search(keyword, page, 'music');

      // 调试：输出原始返回数据的类型和结构
      console.error(`[MusicFreeAdapter] [DEBUG] Plugin ${pluginId} search result type: ${typeof result}`);
      console.error(`[MusicFreeAdapter] [DEBUG] Plugin ${pluginId} search result is array: ${Array.isArray(result)}`);
      if (result && typeof result === 'object') {
        console.error(`[MusicFreeAdapter] [DEBUG] Plugin ${pluginId} result keys: ${JSON.stringify(Object.keys(result))}`);
        if (result.data !== undefined) {
          console.error(`[MusicFreeAdapter] [DEBUG] result.data type: ${typeof result.data}, is array: ${Array.isArray(result.data)}`);
        }
        if (result.songs !== undefined) {
          console.error(`[MusicFreeAdapter] [DEBUG] result.songs type: ${typeof result.songs}, is array: ${Array.isArray(result.songs)}`);
        }
      }

      // 检查返回数据
      if (!result) {
        console.error(`[MusicFreeAdapter] Plugin ${pluginId} returned null/undefined`);
        return { songs: [], total: 0, page };
      }

      // 处理不同的数据格式
      let songs = [];
      if (Array.isArray(result)) {
        // 直接返回数组
        songs = result;
      } else if (result.data && Array.isArray(result.data)) {
        // 返回 { data: [...] } 格式
        songs = result.data;
      } else if (result.songs && Array.isArray(result.songs)) {
        // 返回 { songs: [...] } 格式
        songs = result.songs;
      } else {
        console.error(`[MusicFreeAdapter] Plugin ${pluginId} returned unexpected format:`, Object.keys(result));
        console.error(`[MusicFreeAdapter] [DEBUG] Full result: ${JSON.stringify(result).substring(0, 500)}`);
        return { songs: [], total: 0, page };
      }

      // 转换为 LX Music 格式（同时保留原始数据）
      const convertedSongs = songs.slice(0, limit).map((item) => {
        const songId = item.id || item.songid;

        // 缓存原始数据（用于后续的歌词/图片获取）
        const cacheKey = `${pluginId}:${songId}`;
        this.musicItemCache.set(cacheKey, item);

        // 建立多音源索引（用于降级）
        const songName = (item.title || item.name || '').trim().toLowerCase();
        const artist = (item.artist || item.author || '').trim().toLowerCase();
        const multiSourceKey = `${songName}:${artist}`;

        if (!this.multiSourceCache.has(multiSourceKey)) {
          this.multiSourceCache.set(multiSourceKey, []);
        }
        this.multiSourceCache.get(multiSourceKey).push({
          pluginId,
          item,
          songId,
        });

        return {
          id: songId,
          name: item.title || item.name,
          singer: item.artist || item.author || '',
          source: pluginId,
          interval: item.interval || '',
          album: item.album || '',
          meta: {
            songId: songId,
            albumName: item.album || '',
            picUrl: item.artwork || item.pic || '',
          },
        };
      });

      return {
        songs: convertedSongs,
        total: result.isEnd ? convertedSongs.length : -1,
        page,
      };
    } catch (err) {
      console.error(`[MusicFreeAdapter] Search error for ${pluginId}:`, err.message);
      throw err;
    }
  }

  /**
   * 获取播放 URL
   */
  async getMusicUrl(pluginId, musicItem, quality = '320k') {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      throw new Error(`Plugin ${pluginId} not found`);
    }

    // 转换质量标识
    const qualityMap = {
      '128k': 'standard',
      '320k': 'high',
      'flac': 'lossless',
    };

    const result = await plugin.getMediaSource(musicItem, qualityMap[quality] || 'high');

    if (!result || !result.url) {
      throw new Error('Failed to get music URL');
    }

    return {
      url: result.url,
      quality,
      size: 0,
    };
  }

  /**
   * 根据成功率获取音源优先级
   */
  getSourcePriority(pluginId) {
    const stats = this.sourceStats.get(pluginId);
    if (!stats || stats.success + stats.fail === 0) {
      return 0.5; // 新音源，默认中等优先级
    }
    return stats.success / (stats.success + stats.fail);
  }

  /**
   * 更新音源统计
   */
  updateSourceStats(pluginId, success) {
    if (!this.sourceStats.has(pluginId)) {
      this.sourceStats.set(pluginId, { success: 0, fail: 0 });
    }
    const stats = this.sourceStats.get(pluginId);
    if (success) {
      stats.success++;
    } else {
      stats.fail++;
    }
  }

  /**
   * 查找同一首歌的其他音源
   */
  findAlternativeSources(musicItem) {
    const songName = (musicItem.name || musicItem.title || '').trim().toLowerCase();
    const artist = (musicItem.singer || musicItem.artist || musicItem.author || '').trim().toLowerCase();
    const key = `${songName}:${artist}`;

    const alternatives = this.multiSourceCache.get(key) || [];

    // 按成功率排序，并过滤掉当前音源
    return alternatives
      .filter(alt => alt.pluginId !== musicItem.source && alt.songId !== musicItem.id)
      .sort((a, b) => this.getSourcePriority(b.pluginId) - this.getSourcePriority(a.pluginId));
  }

  /**
   * 更新音源统计
   */
  updateSourceStats(pluginId, success) {
    if (!this.sourceStats.has(pluginId)) {
      this.sourceStats.set(pluginId, { success: 0, fail: 0 });
    }
    const stats = this.sourceStats.get(pluginId);
    if (success) {
      stats.success++;
    } else {
      stats.fail++;
    }
  }

  /**
   * 获取歌词（带重试和多音源降级）
   */
  async getLyric(pluginId, musicItem) {
    // 尝试从缓存中获取原始 musicItem
    const cacheKey = `${pluginId}:${musicItem.id}`;
    const cachedItem = this.musicItemCache.get(cacheKey);

    // 优先使用缓存的原始数据，否则使用传入的 musicItem
    let itemToUse = cachedItem || musicItem;

    // 如果没有缓存，构造兼容格式
    if (!cachedItem) {
      itemToUse = {
        ...musicItem,
        songid: musicItem.id,
        title: musicItem.name || musicItem.title,
        artist: musicItem.singer || musicItem.artist || musicItem.author,
        author: musicItem.singer || musicItem.artist || musicItem.author,
      };
    }

    // 收集所有候选音源（包括原始音源）
    const alternatives = this.findAlternativeSources(musicItem);
    const allSources = [
      { pluginId, item: itemToUse, isOriginal: true },
      ...alternatives.map(alt => ({
        pluginId: alt.pluginId,
        item: alt.item,
        isOriginal: false
      }))
    ];

    if (allSources.length > 1) {
      console.error(`[MusicFreeAdapter] Multi-source fallback: ${allSources.length} sources available (1 original + ${alternatives.length} alternatives)`);
    }

    // 尝试每个音源（最多 3 个）
    const maxSources = Math.min(3, allSources.length);

    for (let sourceIdx = 0; sourceIdx < maxSources; sourceIdx++) {
      const source = allSources[sourceIdx];
      const plugin = this.plugins.get(source.pluginId);

      if (!plugin || !plugin.getLyric) {
        continue;
      }

      // 对每个音源最多重试 2 次
      const maxRetries = 2;
      for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
          const result = await plugin.getLyric(source.item);

          // 处理不同的返回格式
          let lyric = '';
          let tlyric = '';

          if (!result) {
            if (attempt < maxRetries) {
              await new Promise(resolve => setTimeout(resolve, 300 * attempt));
              continue;
            }
          } else if (typeof result === 'string') {
            lyric = result;
          } else if (typeof result === 'object') {
            lyric = result.rawLrc || result.lrc || result.lyric || result.lrclist || result.text || '';
            tlyric = result.translation || result.tlyric || result.translated || '';
          }

          // 成功获取歌词
          if (lyric.length > 0) {
            if (!source.isOriginal) {
              console.error(`[MusicFreeAdapter] Fallback to alternative source: ${source.pluginId} (lyric length: ${lyric.length})`);
            }
            this.updateSourceStats(source.pluginId, true);
            return { lyric, tlyric };
          }

          // 歌词为空，重试
          if (attempt < maxRetries) {
            await new Promise(resolve => setTimeout(resolve, 300 * attempt));
          }
        } catch (err) {
          if (attempt === maxRetries) {
            // 这个音源所有重试都失败了
            this.updateSourceStats(source.pluginId, false);
            if (sourceIdx < maxSources - 1) {
              console.error(`[MusicFreeAdapter] Source ${source.pluginId} failed: ${err.message}, trying next source...`);
            }
          } else {
            await new Promise(resolve => setTimeout(resolve, 300 * attempt));
          }
        }
      }
    }

    console.error(`[MusicFreeAdapter] All ${maxSources} sources failed for song: ${musicItem.name || musicItem.title}`);
    return { lyric: '', tlyric: '' };
  }

  /**
   * 获取封面图片
   */
  async getPic(pluginId, musicItem) {
    // 直接返回音乐项中的 artwork
    if (musicItem.artwork) {
      return { url: musicItem.artwork };
    }

    const plugin = this.plugins.get(pluginId);
    if (!plugin || !plugin.getPic) {
      return { url: '' };
    }

    const result = await plugin.getPic(musicItem);
    return { url: result || '' };
  }

  /**
   * 获取所有已加载的插件
   */
  getPlugins() {
    return Array.from(this.plugins.keys());
  }
}

module.exports = MusicFreeAdapter;
