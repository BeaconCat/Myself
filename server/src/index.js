import express from 'express';
import { seedIfEmpty } from './db.js';
import { postsRouter } from './routes/posts.js';

seedIfEmpty();

const app = express();
app.use(express.json({ limit: '2mb' }));

app.use('/api/v1', postsRouter);

app.use((err, _req, res, _next) => {
  console.error(err);
  res.status(500).json({ error: 'internal_error' });
});

const PORT = Number(process.env.PORT) || 3100;
app.listen(PORT, () => {
  console.log(`[myself-server] listening on http://localhost:${PORT}`);
});
