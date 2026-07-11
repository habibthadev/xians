import { z } from "zod";

// ── File manifest ──

export const FileManifestEntrySchema = z.object({
  name: z.string().min(1).max(255),
  sizeBytes: z.number().int().positive(),
  contentType: z.enum(["image/jpeg", "image/png", "image/webp", "image/gif", "video/mp4", "video/quicktime", "video/webm"]),
});

export const InitBatchRequestSchema = z.object({
  files: z.array(FileManifestEntrySchema).min(1).max(20),
});

export type FileManifestEntry = z.infer<typeof FileManifestEntrySchema>;
export type InitBatchRequest = z.infer<typeof InitBatchRequestSchema>;

// ── Init response ──

export const PresignedUrlSchema = z.object({
  fileId: z.string(),
  url: z.string().url(),
  method: z.literal("PUT"),
  parts: z
    .array(
      z.object({
        partNumber: z.number().int().positive(),
        url: z.string().url(),
      })
    )
    .optional(),
});

export const InitBatchResponseSchema = z.object({
  batchId: z.string(),
  files: z.array(PresignedUrlSchema),
});

export type PresignedUrl = z.infer<typeof PresignedUrlSchema>;
export type InitBatchResponse = z.infer<typeof InitBatchResponseSchema>;

// ── Complete ──

export const CompleteBatchRequestSchema = z.object({
  ttl: z.enum(["24h", "7d", "30d", "never"]).default("7d"),
  password: z.string().max(128).optional(),
  maxDownloads: z.number().int().positive().max(1000).optional(),
});

export const CompleteBatchResponseSchema = z.object({
  slug: z.string(),
  ownerToken: z.string(),
});

export type CompleteBatchRequest = z.infer<typeof CompleteBatchRequestSchema>;
export type CompleteBatchResponse = z.infer<typeof CompleteBatchResponseSchema>;

// ── Gallery ──

export const FileInfoSchema = z.object({
  id: z.string(),
  originalName: z.string(),
  contentType: z.string(),
  sizeBytes: z.number().int().positive(),
  width: z.number().int().positive().nullable(),
  height: z.number().int().positive().nullable(),
  durationMs: z.number().int().positive().nullable(),
});

export const GalleryResponseSchema = z.object({
  slug: z.string(),
  files: z.array(FileInfoSchema),
  createdAt: z.string(),
  expiresAt: z.string().nullable(),
  requiresPassword: z.boolean(),
  downloadCount: z.number().int().nonnegative(),
  maxDownloads: z.number().int().positive().nullable(),
  status: z.enum(["active", "expired", "deleted", "reported"]),
});

export type FileInfo = z.infer<typeof FileInfoSchema>;
export type GalleryResponse = z.infer<typeof GalleryResponseSchema>;

// ── Unlock ──

export const UnlockRequestSchema = z.object({
  password: z.string().min(1).max(128),
});

export const UnlockResponseSchema = z.object({
  success: z.literal(true),
});

export type UnlockRequest = z.infer<typeof UnlockRequestSchema>;
export type UnlockResponse = z.infer<typeof UnlockResponseSchema>;

// ── Report ──

export const ReportRequestSchema = z.object({
  reason: z.string().min(1).max(500),
});

export type ReportRequest = z.infer<typeof ReportRequestSchema>;

// ── Error ──

export const ApiErrorSchema = z.object({
  error: z.string(),
  details: z.any().optional(),
});

export type ApiError = z.infer<typeof ApiErrorSchema>;
