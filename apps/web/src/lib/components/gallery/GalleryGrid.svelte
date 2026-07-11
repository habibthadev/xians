<script lang="ts">
  import { tick } from 'svelte';
  import gsap from 'gsap';
  import type { FileInfo } from '@xians/shared-types';
  import MediaCard from './MediaCard.svelte';
  import Lightbox from './Lightbox.svelte';
  import { API_BASE } from '$lib/api/env';

  let {
    files,
    slug,
    createdAt,
    expiresAt,
    downloadCount,
    maxDownloads,
  }: {
    files: FileInfo[];
    slug: string;
    createdAt: string;
    expiresAt: string | null;
    downloadCount: number;
    maxDownloads: number | null;
  } = $props();

  let grid: HTMLDivElement;
  let lightboxFile: FileInfo | null = $state(null);

  const reducedMotion = $derived(
    typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
  );

  $effect(() => {
    if (reducedMotion || !grid) return;
    const ctx = gsap.context(() => {
      gsap.fromTo(
        grid.children,
        { y: 30, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.4, stagger: 0.06, ease: 'power3.out' }
      );
    }, grid);
    return () => ctx.revert();
  });

  const zipUrl = $derived(`${API_BASE}/api/g/${slug}/zip`);
  const totalSize = $derived(files.reduce((sum, f) => sum + f.sizeBytes, 0));
</script>

<div>
  <div class="mb-8">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <p class="text-lg font-semibold text-ink">Gallery</p>
        <p class="mt-2 text-sm text-body">
          {files.length} file{files.length !== 1 ? 's' : ''} - {(totalSize / 1024 / 1024).toFixed(1)} MB total
        </p>
        {#if expiresAt}
          <p class="mt-1 text-sm text-body">Expires {new Date(expiresAt).toLocaleDateString()}</p>
        {/if}
        {#if maxDownloads}
          <p class="mt-1 text-sm text-body">{downloadCount}/{maxDownloads} downloads</p>
        {/if}
      </div>
      <a href={zipUrl} class="btn-primary">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
        Download all
      </a>
    </div>
  </div>

  <div bind:this={grid} class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
    {#each files as file}
      <MediaCard {file} {slug} onclick={() => (lightboxFile = file)} />
    {/each}
  </div>
</div>

{#if lightboxFile}
  <Lightbox file={lightboxFile} {slug} onclose={() => (lightboxFile = null)} />
{/if}
