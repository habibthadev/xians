<script lang="ts">
  import Iconify from '@iconify/svelte';
  import type { FileInfo } from '@xians/shared-types';
  import { API_BASE } from '$lib/api/env';

  let {
    file,
    slug,
    onclick,
  }: {
    file: FileInfo;
    slug: string;
    onclick?: () => void;
  } = $props();

  const isVideo = $derived(file.contentType.startsWith('video/'));
  const isGif = $derived(file.contentType === 'image/gif');
  const viewUrl = $derived(`${API_BASE}/api/g/${slug}/files/${file.id}/view`);
  const downloadUrl = $derived(`${API_BASE}/api/g/${slug}/files/${file.id}`);
</script>

<div
  role="button"
  tabindex="0"
  class="group relative cursor-pointer overflow-hidden rounded-2xl bg-surface-alt transition-all duration-300 hover:bg-surface-alt/80"
  {onclick}
  onkeydown={(e) => e.key === 'Enter' && onclick?.()}
>
  {#if isVideo}
    <div class="flex aspect-video items-center justify-center bg-surface-alt">
      <div class="flex flex-col items-center gap-2 text-accent">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
        <span class="text-xs font-medium text-body">{file.contentType.split('/')[1].toUpperCase()}</span>
      </div>
    </div>
  {:else}
    <img
      src={viewUrl}
      alt={file.originalName}
      loading="lazy"
      class="aspect-square w-full object-cover"
    />
  {/if}

  <div class="absolute inset-0 flex items-end bg-gradient-to-t from-ink/80 via-ink/20 to-transparent p-4 opacity-0 transition-opacity duration-300 group-hover:opacity-100">
    <div class="flex w-full items-center justify-between">
      <p class="truncate text-sm font-medium text-white">{file.originalName}</p>
      <a href={downloadUrl} download class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-white/20 text-white hover:bg-white/30" onclick={(e) => e.stopPropagation()}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
      </a>
    </div>
  </div>

  <div class="absolute right-3 top-3">
    <span class="rounded-full bg-ink/70 px-2.5 py-1 text-xs font-medium text-white">
      {(file.sizeBytes / 1024 / 1024).toFixed(1)} MB
    </span>
  </div>
</div>
