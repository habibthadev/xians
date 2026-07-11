<script lang="ts">
  import gsap from 'gsap';
  import type { FileInfo } from '@xians/shared-types';
  import { API_BASE } from '$lib/api/env';

  let {
    file,
    slug,
    onclose,
  }: {
    file: FileInfo;
    slug: string;
    onclose: () => void;
  } = $props();

  let overlay: HTMLDivElement;
  let content: HTMLDivElement;

  const viewUrl = $derived(`${API_BASE}/api/g/${slug}/files/${file.id}/view`);
  const downloadUrl = $derived(`${API_BASE}/api/g/${slug}/files/${file.id}`);

  $effect(() => {
    const ctx = gsap.context(() => {
      gsap.fromTo(overlay, { opacity: 0 }, { opacity: 1, duration: 0.2 });
      gsap.fromTo(content, { scale: 0.95, opacity: 0 }, { scale: 1, opacity: 1, duration: 0.3, ease: 'power3.out' });
    });

    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onclose();
    }
    window.addEventListener('keydown', handleKey);

    return () => {
      ctx.revert();
      window.removeEventListener('keydown', handleKey);
    };
  });

  function close() {
    gsap.to(overlay, { opacity: 0, duration: 0.2 });
    gsap.to(content, { scale: 0.95, opacity: 0, duration: 0.2, onComplete: onclose });
  }
</script>

<div
  bind:this={overlay}
  class="fixed inset-0 z-50 flex items-center justify-center bg-ink/80 p-4 backdrop-blur-sm"
  onclick={close}
  onkeydown={(e) => e.key === 'Escape' && close()}
  role="dialog"
  tabindex="-1"
>
  <div
    bind:this={content}
    class="relative max-h-[90vh] max-w-[90vw] overflow-hidden rounded-3xl bg-surface"
    onclick={(e) => e.stopPropagation()}
  >
    <div class="flex items-center justify-between border-b border-surface-alt px-6 py-4">
      <p class="truncate text-sm font-medium text-ink">{file.originalName}</p>
      <div class="flex items-center gap-3">
        <a
          href={downloadUrl}
          download
          class="btn-primary !px-4 !py-2 !text-xs"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          Download
        </a>
        <button onclick={close} class="flex h-9 w-9 items-center justify-center rounded-full text-body hover:bg-surface-alt hover:text-ink" aria-label="Close">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>
    </div>
    <div class="flex items-center justify-center p-4">
      {#if file.contentType.startsWith('video/')}
        <video src={viewUrl} controls class="max-h-[70vh] max-w-full rounded-xl"></video>
      {:else}
        <img src={viewUrl} alt={file.originalName} class="max-h-[70vh] max-w-full rounded-xl object-contain" />
      {/if}
    </div>
  </div>
</div>
