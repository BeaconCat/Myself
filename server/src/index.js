import express from 'express';
import { initAuth } from './auth.js';
import { seedIfEmpty } from './db.js';
import { adminRouter } from './routes/admin.js';
import { postsRouter } from './routes/posts.js';
import { settingsRouter } from './routes/settings.js';
import { apiKeysRouter } from './routes/apikeys.js';
import { externalRouter } from './routes/external.js';
import { mediaRouter, UPLOAD_DIR } from './routes/media.js';
import { dataRouter, startAutoBackup } from './routes/datacenter.js';
import { qualityRouter } from './routes/quality.js';

seedIfEmpty();
initAuth();

const app = express();
app.use(express.json({ limit: '2mb' }));

app.use('/uploads', express.static(UPLOAD_DIR, { dotfiles: 'ignore', maxAge: '1d' }));
app.use('/api/v1', postsRouter);
app.use('/api/v1', mediaRouter);
app.use('/api/v1', dataRouter);
app.use('/api/v1', qualityRouter);
startAutoBackup();
app.use('/api/v1', settingsRouter);
app.use('/api/v1', apiKeysRouter);
app.use('/api/v1', externalRouter);
app.use('/api/v1', adminRouter);

app.use((err, _req, res, _next) => {
  console.error(err);
  res.status(500).json({ error: 'internal_error' });
});

const PORT = Number(process.env.PORT) || 3100;
app.listen(PORT, () => {
  console.log(`[myself-server] listening on http://localhost:${PORT}`);
});
