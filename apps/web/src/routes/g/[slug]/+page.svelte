<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import type { PageProps } from './$types';
  import type { FileInfo } from '@xians/shared-types';
  import PasswordGate from '$components/gallery/PasswordGate.svelte';
  import GalleryGrid from '$components/gallery/GalleryGrid.svelte';
  import { API_BASE } from '$lib/api/env';

  let { data }: PageProps = $props();
  let unlocked = $state(false);
  let loading = $state(true);
  let galleryData = $state<{
    files: FileInfo[];
    createdAt: string;
    expiresAt: string | null;
    downloadCount: number;
    maxDownloads: number | null;
  } | null>(null);
  let loadError = $state<string | null>(null);

  const siteUrl = 'https://xians.vercel.app';
  const ogUrl = $derived(
    `${siteUrl}/og?title=${encodeURIComponent(data.title || 'Gallery - xians')}&description=${encodeURIComponent(data.description || 'Shared media gallery')}`
  );

  async function fetchGallery() {
    loading = true;
    loadError = null;
    try {
      const res = await fetch(`${API_BASE}/api/g/${data.slug}`, { credentials: 'include' });
      if (!res.ok) throw new Error(res.status === 401 ? 'Unauthorized' : 'Failed to load gallery');
      const json = await res.json();
      galleryData = {
        files: json.files || [],
        createdAt: json.createdAt,
        expiresAt: json.expiresAt,
        downloadCount: json.downloadCount ?? 0,
        maxDownloads: json.maxDownloads,
      };
    } catch (e) {
      loadError = (e as Error).message;
    } finally {
      loading = false;
    }
  }

  async function handleUnlocked() {
    unlocked = true;
    await fetchGallery();
  }

  onMount(async () => {
    if (!data.requiresPassword) {
      galleryData = data.files
        ? {
            files: data.files,
            createdAt: data.createdAt || '',
            expiresAt: data.expiresAt,
            downloadCount: data.downloadCount ?? 0,
            maxDownloads: data.maxDownloads,
          }
        : null;
      unlocked = true;
      loading = false;
    } else {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>{data.title || 'Gallery - xians'}</title>
  <meta name="description" content={data.description || 'Shared media gallery'} />
  <meta name="robots" content="index, follow" />
  <meta property="og:title" content={data.title || 'Gallery - xians'} />
  <meta property="og:description" content={data.description || 'Shared media gallery'} />
  <meta property="og:type" content="website" />
  <meta property="og:url" content={`${page.url.origin}/g/${data.slug}`} />
  <meta property="og:image" content={ogUrl} />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content={data.title || 'Gallery - xians'} />
  <meta name="twitter:description" content={data.description || 'Shared media gallery'} />
</svelte:head>

<div class="mx-auto min-h-dvh max-w-6xl px-5 py-8 md:px-8">
  {#if loading}
    <div class="flex min-h-[60dvh] items-center justify-center">
      <div class="flex items-center gap-3 text-body">
        <svg class="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10" stroke-dasharray="31.4 31.4" stroke-linecap="round"/></svg>
        <span class="text-sm">Loading gallery...</span>
      </div>
    </div>
  {:else if loadError}
    <div class="flex min-h-[60dvh] flex-col items-center justify-center gap-6 text-center">
      <div class="flex h-20 w-20 items-center justify-center rounded-full bg-surface-alt">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-body"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
      </div>
      <p class="text-lg font-semibold text-ink">Something went wrong</p>
      <p class="max-w-sm text-sm text-body">{loadError === 'Unauthorized' ? 'Invalid or expired access. Please refresh and try again.' : loadError}</p>
      <button onclick={() => fetchGallery()} class="btn-primary mt-2">Try again</button>
    </div>
  {:else if !unlocked}
    <div class="flex min-h-[60dvh] items-center justify-center">
      <PasswordGate slug={data.slug} onUnlocked={handleUnlocked} />
    </div>
  {:else if galleryData?.files && galleryData.files.length > 0}
    <GalleryGrid
      files={galleryData.files}
      slug={data.slug}
      createdAt={galleryData.createdAt || ''}
      expiresAt={galleryData.expiresAt}
      downloadCount={galleryData.downloadCount ?? 0}
      maxDownloads={galleryData.maxDownloads}
    />
  {:else if galleryData?.files && galleryData.files.length === 0}
    <div class="flex min-h-[60dvh] flex-col items-center justify-center gap-6 text-center">
      <div class="flex h-20 w-20 items-center justify-center rounded-full bg-surface-alt">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-body"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
      </div>
      <p class="text-lg font-semibold text-ink">No files</p>
      <p class="max-w-sm text-sm text-body">This gallery has no files yet.</p>
      <a href="/" class="btn-primary mt-2">Go home</a>
    </div>
  {:else}
    <div class="flex min-h-[60dvh] flex-col items-center justify-center gap-6 text-center">
      <div class="flex h-20 w-20 items-center justify-center rounded-full bg-surface-alt">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-body"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
      </div>
      <p class="text-lg font-semibold text-ink">Gallery not found</p>
      <p class="max-w-sm text-sm text-body">
        This link may have expired or been removed. Ask the sender for a new one.
      </p>
      <a href="/" class="btn-primary mt-2">Go home</a>
    </div>
  {/if}
</div>
