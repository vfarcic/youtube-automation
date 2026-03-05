import { describe, it, expect, afterEach } from 'vitest';
import { http, HttpResponse } from 'msw';
import { get, ApiError } from '../api/client';
import { server } from './server';

afterEach(() => {
  localStorage.clear();
});

describe('API client', () => {
  it('adds Bearer token from localStorage', async () => {
    localStorage.setItem('api_token', 'my-secret');

    let capturedAuth = '';
    server.use(
      http.get('/api/test', ({ request }) => {
        capturedAuth = request.headers.get('Authorization') ?? '';
        return HttpResponse.json({ ok: true });
      }),
    );

    await get('/api/test');
    expect(capturedAuth).toBe('Bearer my-secret');
  });

  it('works without token', async () => {
    let capturedAuth: string | null = '';
    server.use(
      http.get('/api/test', ({ request }) => {
        capturedAuth = request.headers.get('Authorization');
        return HttpResponse.json({ ok: true });
      }),
    );

    await get('/api/test');
    expect(capturedAuth).toBeNull();
  });

  it('throws ApiError on 401', async () => {
    server.use(
      http.get('/api/test', () => new HttpResponse('Unauthorized', { status: 401 })),
    );

    await expect(get('/api/test')).rejects.toThrow(ApiError);
  });

  it('throws ApiError on 404', async () => {
    server.use(
      http.get('/api/test', () => new HttpResponse('Not Found', { status: 404 })),
    );

    await expect(get('/api/test')).rejects.toThrow(ApiError);
  });

  it('throws ApiError on 500', async () => {
    server.use(
      http.get('/api/test', () => new HttpResponse('Server Error', { status: 500 })),
    );

    await expect(get('/api/test')).rejects.toThrow(ApiError);
  });
});
