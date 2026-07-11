import {
  InitBatchResponseSchema,
  type InitBatchRequest,
  type InitBatchResponse,
  type CompleteBatchRequest,
  type CompleteBatchResponse,
  type GalleryResponse,
  type UnlockRequest,
} from '@xians/shared-types';
import { API_BASE } from './env';

class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public details?: unknown
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error || 'Request failed', body);
  }

  return res.json();
}

export function initBatch(data: InitBatchRequest): Promise<InitBatchResponse> {
  return request('/api/batches/init', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function completeBatch(
  batchId: string,
  data: CompleteBatchRequest
): Promise<CompleteBatchResponse> {
  return request(`/api/batches/${batchId}/complete`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function getGallery(slug: string): Promise<GalleryResponse> {
  return request(`/api/g/${slug}`);
}

export function unlockGallery(slug: string, data: UnlockRequest): Promise<{ success: true }> {
  return request(`/api/g/${slug}/unlock`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function downloadFileUrl(slug: string, fileId: string): string {
  return `${API_BASE}/api/g/${slug}/files/${fileId}`;
}

export function downloadZipUrl(slug: string): string {
  return `${API_BASE}/api/g/${slug}/zip`;
}

export function reportBatch(slug: string, reason: string): Promise<void> {
  return request(`/api/g/${slug}/report`, {
    method: 'POST',
    body: JSON.stringify({ reason }),
  });
}

export function deleteBatch(
  slug: string,
  ownerToken: string
): Promise<void> {
  return fetch(`${API_BASE}/api/g/${slug}`, {
    method: 'DELETE',
    headers: { 'X-Owner-Token': ownerToken },
  }).then((r) => {
    if (!r.ok) throw new ApiError(r.status, 'Delete failed');
  });
}
