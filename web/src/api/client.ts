export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('api_token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(path, { ...options, headers });

  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new ApiError(res.status, body || res.statusText);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export function get<T>(path: string): Promise<T> {
  return request<T>(path);
}

export async function getBlob(path: string): Promise<Blob> {
  const token = localStorage.getItem('api_token');
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const res = await fetch(path, { headers });
  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new ApiError(res.status, body || res.statusText);
  }
  return res.blob();
}

export function post<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'POST', body: JSON.stringify(body) });
}

export function put<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'PUT', body: JSON.stringify(body) });
}

export function patch<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'PATCH', body: JSON.stringify(body) });
}

export function del(path: string): Promise<void> {
  return request<void>(path, { method: 'DELETE' });
}

export async function uploadFile<T>(path: string, file: File, fieldName: string = 'thumbnail'): Promise<T> {
  const token = localStorage.getItem('api_token');
  const formData = new FormData();
  formData.append(fieldName, file);

  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  // Do NOT set Content-Type — browser sets it with multipart boundary

  const res = await fetch(path, { method: 'POST', headers, body: formData });
  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new ApiError(res.status, body || res.statusText);
  }
  return res.json();
}

export function uploadFileWithProgress<T>(
  path: string,
  file: File,
  fieldName: string,
  onProgress?: (percent: number) => void,
): Promise<T> {
  return new Promise((resolve, reject) => {
    const token = localStorage.getItem('api_token');
    const formData = new FormData();
    formData.append(fieldName, file);

    const xhr = new XMLHttpRequest();
    xhr.open('POST', path);

    if (token) {
      xhr.setRequestHeader('Authorization', `Bearer ${token}`);
    }

    // Phase 1 (0–50%): XHR send — tracks bytes leaving the browser buffer.
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) {
        onProgress(Math.round((e.loaded / e.total) * 50));
      }
    };

    // Phase 2 (50→95%): Server-side processing creep — starts after send completes.
    let creepTimer: ReturnType<typeof setInterval> | undefined;
    let currentProgress = 0;

    xhr.upload.onloadend = () => {
      currentProgress = 50;
      creepTimer = setInterval(() => {
        const remaining = 95 - currentProgress;
        const step = Math.max(0.5, remaining * 0.05);
        currentProgress = Math.min(95, currentProgress + step);
        onProgress?.(Math.round(currentProgress));
      }, 500);
    };

    xhr.onload = () => {
      clearInterval(creepTimer);
      onProgress?.(100);
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText));
        } catch {
          reject(new ApiError(xhr.status, 'Invalid JSON response'));
        }
      } else {
        reject(new ApiError(xhr.status, xhr.responseText || xhr.statusText));
      }
    };

    xhr.onerror = () => {
      clearInterval(creepTimer);
      reject(new ApiError(0, 'Network error'));
    };

    xhr.send(formData);
  });
}
