import { Router } from 'express';
import { Agent, ProxyAgent, fetch as undiciFetch } from 'undici';
import { requireAuth } from '../auth.js';
import { getConfig } from '../config.js';

/**
 * GitHub 状态：服务端拉取 GitHub API（避免浏览器 CORS、隐藏票证）。
 * 配置 github.mode = 'manual'（手填数字）| 'api'（自动拉取）；
 * token 可选（只读 PAT，提升配额并可查私有统计）；refreshMinutes 控制缓存。
 */
let cache = { at: 0, key: '', data: null };

/** 同步日志（进程内，最近 20 条） */
const syncLog = [];

function logSync(ok, message) {
  syncLog.unshift({ at: new Date().toISOString(), ok, message });
  if (syncLog.length > 20) syncLog.pop();
}

/**
 * 出站策略：可配代理（github.proxy）与跳过 TLS 校验（github.insecureTls）。
 * insecureTls 仅作用于本只读拉取通道——用于本机存在 TLS 注入
 * （安全软件/TUN 代理自签证书）导致 Node 证书验证失败的环境。
 */
function dispatcher() {
  const cfg = getConfig().github;
  const connect = cfg.insecureTls ? { rejectUnauthorized: false } : undefined;
  const proxy = String(cfg.proxy ?? '').trim() || process.env.HTTPS_PROXY || process.env.HTTP_PROXY;
  if (proxy) return new ProxyAgent({ uri: proxy, connect });
  return connect ? new Agent({ connect }) : undefined;
}

async function gh(path, token) {
  const headers = {
    Accept: 'application/vnd.github+json',
    'User-Agent': 'myself-blog',
  };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await undiciFetch(`https://api.github.com${path}`, {
    headers,
    dispatcher: dispatcher(),
  });
  if (!res.ok) throw new Error(`github_${res.status}`);
  return res.json();
}

/**
 * 贡献热力图：
 * 有票证 → GraphQL contributionCalendar（含精确次数）；
 * 无票证 → 抓取公开贡献页 HTML 解析 data-date/data-level。
 */
async function fetchCalendar(username, token) {
  if (token) {
    const res = await undiciFetch('https://api.github.com/graphql', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
        'User-Agent': 'myself-blog',
      },
      dispatcher: dispatcher(),
      body: JSON.stringify({
        query: `query($login:String!){user(login:$login){contributionsCollection{contributionCalendar{weeks{contributionDays{date contributionCount contributionLevel}}}}}}`,
        variables: { login: username },
      }),
    });
    if (!res.ok) throw new Error(`github_graphql_${res.status}`);
    const json = await res.json();
    const weeks = json?.data?.user?.contributionsCollection?.contributionCalendar?.weeks ?? [];
    const LEVELS = { NONE: 0, FIRST_QUARTILE: 1, SECOND_QUARTILE: 2, THIRD_QUARTILE: 3, FOURTH_QUARTILE: 4 };
    return weeks.flatMap((w) => w.contributionDays.map((d) => ({
      date: d.date,
      count: d.contributionCount,
      level: LEVELS[d.contributionLevel] ?? 0,
    })));
  }

  const res = await undiciFetch(
    `https://github.com/users/${encodeURIComponent(username)}/contributions`,
    { headers: { 'User-Agent': 'myself-blog' }, dispatcher: dispatcher() },
  );
  if (!res.ok) throw new Error(`github_contrib_${res.status}`);
  const html = await res.text();
  const days = [];
  const re = /data-date="(\d{4}-\d{2}-\d{2})"[^>]*data-level="(\d)"/g;
  let match;
  while ((match = re.exec(html)) !== null) {
    days.push({ date: match[1], count: 0, level: Number(match[2]) });
  }
  days.sort((a, b) => (a.date < b.date ? -1 : 1));
  return days;
}

async function fetchStatus(username, token) {
  const user = await gh(`/users/${encodeURIComponent(username)}`, token);
  const repos = await gh(
    `/users/${encodeURIComponent(username)}/repos?per_page=100&sort=pushed`,
    token,
  );
  const stars = repos.reduce((sum, r) => sum + (r.stargazers_count ?? 0), 0);

  // 年度提交总数：Commit Search API（匿名可用，配额低；失败回退事件统计）
  let commitsThisYear = 0;
  const year = new Date().getFullYear();
  try {
    const search = await gh(
      `/search/commits?q=${encodeURIComponent(`author:${username} author-date:>=${year}-01-01`)}&per_page=1`,
      token,
    );
    commitsThisYear = search.total_count ?? 0;
  } catch { /* 回退下方事件粗算 */ }

  // 最近公开动态（PushEvent 提交）
  let activities = [];
  try {
    const events = await gh(`/users/${encodeURIComponent(username)}/events/public?per_page=30`, token);
    let eventCommits = 0;
    for (const ev of events) {
      if (ev.type !== 'PushEvent') continue;
      if (new Date(ev.created_at).getFullYear() === year) {
        eventCommits += ev.payload?.commits?.length ?? 0;
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
    if (!commitsThisYear) commitsThisYear = eventCommits;
  } catch { /* 动态失败不致命 */ }

  // 贡献热力图（失败不致命）
  let heatmap = [];
  try {
    heatmap = await fetchCalendar(username, token);
  } catch { /* 留空，前端隐藏 */ }

  return {
    username,
    stats: {
      repos: user.public_repos ?? repos.length,
      stars,
      followers: user.followers ?? 0,
      commits: commitsThisYear,
    },
    activities,
    heatmap,
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
    logSync(true, `拉取成功：${data.stats.repos} 仓库 / ${data.stats.stars} stars`);
    res.json({ mode: 'api', cached: false, ...data });
  } catch (err) {
    logSync(false, String(err?.message ?? err));
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

/** POST /api/v1/admin/github/sync 立即同步（强刷缓存），返回最新数据 */
githubRouter.post('/admin/github/sync', requireAuth, async (_req, res) => {
  const cfg = getConfig().github;
  try {
    const data = await fetchStatus(cfg.username, cfg.token || undefined);
    cache = { at: Date.now(), key: `${cfg.username}:${cfg.token ? 'tok' : 'anon'}`, data };
    logSync(true, `手动同步成功：${data.stats.repos} 仓库 / ${data.stats.stars} stars`);
    res.json({ ok: true, data });
  } catch (err) {
    logSync(false, `手动同步失败：${String(err?.message ?? err)}`);
    res.status(502).json({ ok: false, error: String(err?.message ?? err) });
  }
});

/** GET /api/v1/admin/github/log 同步日志 + 当前缓存预览 */
githubRouter.get('/admin/github/log', requireAuth, (_req, res) => {
  res.json({
    log: syncLog,
    preview: cache.data,
    cachedAt: cache.at ? new Date(cache.at).toISOString() : null,
  });
});
