<script lang="ts">
  import gsap from 'gsap';

  let {
    onFilesSelected,
  }: {
    onFilesSelected: (files: File[]) => void;
  } = $props();

  let isDragging = $state(false);
  let zone: HTMLDivElement;
  let input: HTMLInputElement;

  $effect(() => {
    const ctx = gsap.context(() => {
      gsap.set(zone, { transformStyle: 'preserve-3d', willChange: 'transform, border-color' });
    }, zone);
    return () => ctx.revert();
  });

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    if (!isDragging) {
      isDragging = true;
      gsap.to(zone, { scale: 1.01, borderColor: '#2F6FEB', duration: 0.2, ease: 'power2.out' });
    }
  }

  function handleDragLeave() {
    isDragging = false;
    gsap.to(zone, { scale: 1, borderColor: '#e2e8f0', duration: 0.2, ease: 'power2.out' });
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
    gsap.to(zone, { scale: 1, duration: 0.2 });
    const files = Array.from(e.dataTransfer?.files || []);
    if (files.length > 0) onFilesSelected(files);
  }

  function handleFilePick(e: Event) {
    const files = Array.from((e.target as HTMLInputElement).files || []);
    if (files.length > 0) onFilesSelected(files);
  }

  function handlePaste(e: ClipboardEvent) {
    const items = Array.from(e.clipboardData?.items || []);
    const files: File[] = [];
    for (const item of items) {
      if (item.kind === 'file') {
        const file = item.getAsFile();
        if (file) files.push(file);
      }
    }
    if (files.length > 0) onFilesSelected(files);
  }
</script>

<svelte:window onpaste={handlePaste} />

<div
  bind:this={zone}
  role="button"
  tabindex="0"
  class={[
    'flex w-full cursor-pointer flex-col items-center gap-5 rounded-3xl border-2 border-dashed p-14 text-center transition-all duration-300',
    isDragging
      ? '!border-accent !bg-accent/5'
      : 'border-surface-alt hover:border-accent/40 hover:bg-accent/5',
  ]}
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
  onclick={() => input.click()}
  onkeydown={(e) => e.key === 'Enter' && input.click()}
>
  <svg width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-accent">
    <rect x="2" y="2" width="20" height="20" rx="4" stroke="currentColor"/>
    <path d="M12 8v8M8 12h8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
  </svg>
  <div>
    <p class="text-lg font-semibold text-ink">Drop files here</p>
    <p class="mt-2 text-sm text-body">or click to browse - paste from clipboard</p>
  </div>
  <div class="mt-2 flex flex-wrap items-center justify-center gap-2 text-xs text-body">
    <span class="rounded-full bg-surface px-3 py-1.5">JPG</span>
    <span class="rounded-full bg-surface px-3 py-1.5">PNG</span>
    <span class="rounded-full bg-surface px-3 py-1.5">WebP</span>
    <span class="rounded-full bg-surface px-3 py-1.5">GIF</span>
    <span class="rounded-full bg-surface px-3 py-1.5">MP4</span>
    <span class="rounded-full bg-surface px-3 py-1.5">MOV</span>
    <span class="rounded-full bg-surface px-3 py-1.5">WebM</span>
    <span>up to 100MB each</span>
  </div>
  <input bind:this={input} type="file" class="hidden" multiple accept="image/jpeg,image/png,image/webp,image/gif,video/mp4,video/quicktime,video/webm" onchange={handleFilePick} />
</div>
