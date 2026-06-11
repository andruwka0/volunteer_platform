import { execFile } from 'node:child_process';
import { cp, mkdir, rm } from 'node:fs/promises';
import { resolve } from 'node:path';
import { promisify } from 'node:util';
import { fileURLToPath } from 'node:url';

const execFileAsync = promisify(execFile);
const root = resolve(fileURLToPath(new URL('..', import.meta.url)));
const dist = resolve(root, 'dist');

await execFileAsync(process.execPath, ['--check', resolve(root, 'server.js')]);
await execFileAsync(process.execPath, ['--check', resolve(root, 'static/js/main.js')]);

await rm(dist, { recursive: true, force: true });
await mkdir(dist, { recursive: true });
await cp(resolve(root, 'index.html'), resolve(dist, 'index.html'));
await cp(resolve(root, 'static'), resolve(dist, 'static'), { recursive: true });

console.log('Frontend build completed: app/dist');
