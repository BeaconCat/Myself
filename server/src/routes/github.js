import { Router } from 'express';
import { getConfig } from '../config.js';

/**
 * GitHub 状态：服务端拉取 GitHub API（避免浏览器 CORS、隐藏票证）。
 * 配置 github.mode = 'manual'（手填数字）| 'api'（自动拉取）；
 * token 可选（只读 PAT，提升配额并可查私有统计）；refreshMinutes 控制缓存。
 */
let cache = { at: 0, key: '', data: null };

async function gh(path, token) {
  const headers = {
    Accept: 'application/vnd.github+json',
    'User-Agent': 'myself-blog',
  };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`https://api.github.com${path}`, { headers });
  if (!res.ok) throw new Error(`github_${res.status}`);
  return res.json();
}

async function fetchStatus(username, token) {
  const user = await gh(`/users/${encodeURIComponent(username)}`, token);
  const repos = await gh(
    `/users/${encodeURIComponent(username)}/repos?per_page=100&sort=pushed`,
    token,
  );
  const stars = repos.reduce((sum, r) => sum + (r.stargazers_count ?? 0), 0);

  // 最近公开动态（PushEvent 提交）
  let activities = [];
  let commitsThisYear = 0;
  try {
    const events = await gh(`/users/${encodeURIComponent(username)}/events/public?per_page=30`, token);
    const year = new Date().getFullYear();
    for (const ev of events) {
      if (ev.type !== 'PushEvent') continue;
      if (new Date(ev.created_at).getFullYear() === year) {
        commitsThisYear += ev.payload?.commits?.length ?? 0;
      }
      for (const commit of ev.payload?.commits ?? []) {
        if (activities.length >= 5) break;
        activities.push({
          type: 'commit',
          repo: ev.repo?.name?.split('/')[1] ?? '',
          text: String(commit.message ?? '').split('\n')[0].slice(0, 80),
          time: ev.created_at,
        });
      }
    }
  } catch { /* 动态失败不致命 */ }

  return {
    username,
    stats: {
      repos: user.public_repos ?? repos.length,
      stars,
      followers: user.followers ?? 0,
      commits: commitsThisYear,
    },
    activities,
    fetchedAt: new Date().toISOString(),
  };
}

export const githubRouter = Router();

/** GET /api/v1/github-status 公开：api 模式拉取（带缓存），manual 模式回配置数字 */
githubRouter.get('/github-status', async (_req, res) => {
  const cfg = getConfig().github;
  if (cfg.mode !== 'api') {
    res.json({
      mode: 'manual',
      username: cfg.username,
      stats: cfg.stats,
      activities: [],
    });
    return;
  }

  const ttl = Math.max(1, Number(cfg.refreshMinutes) || 30) * 60_000;
  const key = `${cfg.username}:${cfg.token ? 'tok' : 'anon'}`;
  if (cache.data && cache.key === key && Date.now() - cache.at < ttl) {
    res.json({ mode: 'api', cached: true, ...cache.data });
    return;
  }

  try {
    const data = await fetchStatus(cfg.username, cfg.token || undefined);
    cache = { at: Date.now(), key, data };
    res.json({ mode: 'api', cached: false, ...data });
  } catch (err) {
    // 拉取失败回退：旧缓存 → 手填数字
    if (cache.data && cache.key === key) {
      res.json({ mode: 'api', cached: true, stale: true, ...cache.data });
      return;
    }
    res.json({
      mode: 'manual',
      fallback: String(err?.message ?? err),
      username: cfg.username,
      stats: cfg.stats,
      activities: [],
    });
  }
});
