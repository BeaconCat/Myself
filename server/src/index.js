import express from 'express';
import { initAuth } from './auth.js';
import { seedIfEmpty } from './db.js';
import { adminRouter } from './routes/admin.js';
import { postsRouter } from './routes/posts.js';
import { settingsRouter } from './routes/settings.js';

seedIfEmpty();
initAuth();

const app = express();
app.use(express.json({ limit: '2mb' }));

app.use('/api/v1', postsRouter);
app.use('/api/v1', settingsRouter);
app.use('/api/v1', adminRouter);

app.use((err, _req, res, _next) => {
  console.error(err);
  res.status(500).json({ error: 'internal_error' });
});

const PORT = Number(process.env.PORT) || 3100;
app.listen(PORT, () => {
  console.log(`[myself-server] listening on http://localhost:${PORT}`);
});
