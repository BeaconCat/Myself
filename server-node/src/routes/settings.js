import { Router } from 'express';
import { requireAuth } from '../auth.js';
import { getConfig, saveConfig } from '../config.js';

export const settingsRouter = Router();

/** GET /api/v1/site-config 公开站点配置（前台启动读取） */
settingsRouter.get('/site-config', (_req, res) => {
  const cfg = getConfig();
  res.json({
    site: cfg.site,
    loading: cfg.loading,
    theme: cfg.theme,
    hero: cfg.hero,
    thoughts: cfg.thoughts,
    covers: cfg.covers,
    timezone: cfg.timezone,
    // 公开配置不下发票证
    github: { ...cfg.github, token: undefined },
    about: cfg.about,
  });
});

/** GET /api/v1/admin/settings 完整配置 */
settingsRouter.get('/admin/settings', requireAuth, (_req, res) => {
  res.json(getConfig());
});

/** PUT /api/v1/admin/settings 部分更新（深合并） */
settingsRouter.put('/admin/settings', requireAuth, (req, res) => {
  if (!req.body || typeof req.body !== 'object') {
    res.status(400).json({ error: 'invalid_config' });
    return;
  }
  res.json(saveConfig(req.body));
});
