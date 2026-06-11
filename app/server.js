import { createServer } from 'node:http';
import { createReadStream, existsSync } from 'node:fs';
import { stat } from 'node:fs/promises';
import { extname, join, normalize, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = fileURLToPath(new URL('.', import.meta.url));
const isProduction = process.env.NODE_ENV === 'production';
const publicRoot = resolve(__dirname, isProduction ? 'dist' : '.');
const backendUrl = new URL(process.env.BACKEND_URL || 'http://localhost:8080');
const port = Number(process.env.FRONTEND_PORT || process.env.PORT || 5173);
const host = process.env.FRONTEND_HOST || '127.0.0.1';
const apiPrefixes = ['/auth', '/events', '/admin'];

const mimeTypes = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.webp': 'image/webp',
};

function isApiRequest(pathname) {
  return apiPrefixes.some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`));
}

function send(res, statusCode, body, headers = {}) {
  res.writeHead(statusCode, headers);
  res.end(body);
}

function proxyToBackend(req, res) {
  const target = new URL(req.url || '/', backendUrl);
  const proxyReq = globalThis.fetch(target, {
    method: req.method,
    headers: req.headers,
    body: ['GET', 'HEAD'].includes(req.method || 'GET') ? undefined : req,
    duplex: 'half',
  });

  proxyReq
    .then(async (backendRes) => {
      const headers = Object.fromEntries(backendRes.headers.entries());
      headers['access-control-allow-origin'] = '*';
      res.writeHead(backendRes.status, headers);
      if (backendRes.body) {
        for await (const chunk of backendRes.body) {
          res.write(chunk);
        }
      }
      res.end();
    })
    .catch((error) => {
      send(
        res,
        502,
        JSON.stringify({ message: `Backend недоступен: ${error.message}` }),
        { 'Content-Type': 'application/json; charset=utf-8' },
      );
    });
}

async function serveStatic(req, res) {
  const requestUrl = new URL(req.url || '/', `http://${req.headers.host || 'localhost'}`);
  const pathname = decodeURIComponent(requestUrl.pathname);

  if (isApiRequest(pathname)) {
    proxyToBackend(req, res);
    return;
  }

  const relativePath = pathname === '/' ? 'index.html' : pathname.slice(1);
  const safePath = normalize(relativePath).replace(/^(\.\.(\/|\\|$))+/, '');
  let filePath = resolve(join(publicRoot, safePath));

  if (!filePath.startsWith(publicRoot)) {
    send(res, 403, 'Forbidden');
    return;
  }

  if (!existsSync(filePath)) {
    filePath = resolve(join(publicRoot, 'index.html'));
  }

  try {
    const info = await stat(filePath);
    if (!info.isFile()) {
      send(res, 404, 'Not found');
      return;
    }
    res.writeHead(200, {
      'Content-Type': mimeTypes[extname(filePath)] || 'application/octet-stream',
      'Cache-Control': isProduction ? 'public, max-age=300' : 'no-store',
    });
    createReadStream(filePath).pipe(res);
  } catch {
    send(res, 404, 'Not found');
  }
}

createServer(serveStatic).listen(port, host, () => {
  console.log(`Frontend: http://${host}:${port}`);
  console.log(`Backend proxy: ${backendUrl.origin}`);
});
