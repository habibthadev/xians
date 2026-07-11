<script lang="ts">
  import gsap from 'gsap';
  import Iconify from '@iconify/svelte';
  import Button from '$components/ui/Button.svelte';
  import ProgressBar from './ProgressBar.svelte';
  import { API_BASE } from '$lib/api/env';

  let {
    files,
    onBatchInit,
    onBack,
    onProceed,
  }: {
    files: File[];
    onBatchInit: (id: string) => void;
    onBack: () => void;
    onProceed: () => void;
  } = $props();

  type FileState = {
    file: File;
    status: 'pending' | 'uploading' | 'done' | 'error';
    progress: number;
    error?: string;
    uploadUrl?: string;
  };

  let fileStates: FileState[] = $state([]);
  let isUploading = $state(false);

  let container: HTMLDivElement;

  const MAX_CONCURRENT = 3;

  $effect(() => {
    fileStates = files.map((f) => ({ file: f, status: 'pending' as const, progress: 0 }));
    const ctx = gsap.context(() => {
      gsap.fromTo(
        container.children,
        { y: 20, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.3, stagger: 0.05, ease: 'power3.out' }
      );
    }, container);
    return () => ctx.revert();
  });

  const allowedTypes = new Set([
    'image/jpeg', 'image/png', 'image/webp', 'image/gif',
    'video/mp4', 'video/quicktime', 'video/webm',
  ]);

  let validationErrors: string[] = $state([]);

  function validate(): boolean {
    validationErrors = [];
    if (files.length > 20) {
      validationErrors.push('Maximum 20 files per batch');
      return false;
    }
    for (const f of files) {
      if (f.size === 0) validationErrors.push(`${f.name}: file is empty`);
      else if (f.size > 100 * 1024 * 1024) validationErrors.push(`${f.name}: exceeds 100MB limit`);
      else if (!allowedTypes.has(f.type)) validationErrors.push(`${f.name}: unsupported file type`);
    }
    return validationErrors.length === 0;
  }

  async function startUpload() {
    if (!validate() || isUploading) return;
    isUploading = true;

    const remaining = fileStates.filter((fs) => fs.status === 'pending');
    if (remaining.length === 0) {
      isUploading = false;
      return;
    }

    if (!fileStates[0]?.uploadUrl) {
      const manifest = fileStates.map((fs) => ({
        name: fs.file.name,
        sizeBytes: fs.file.size,
        contentType: fs.file.type,
      }));

      try {
        const resp = await fetch(
          `${API_BASE}/api/batches/init`,
          { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ files: manifest }) }
        );
        if (!resp.ok) throw new Error((await resp.json()).error || 'Init failed');
        const data = await resp.json();
        onBatchInit(data.batchId);

        for (let i = 0; i < data.files.length; i++) {
          fileStates[i].uploadUrl = data.files[i].url;
        }
      } catch (err) {
        console.error('Init failed:', err);
        isUploading = false;
        return;
      }
    }

    const pending = fileStates.filter((fs) => fs.status === 'pending');
    const queue = pending.slice();
    let inFlight = 0;

    function processNext() {
      while (inFlight < MAX_CONCURRENT && queue.length > 0) {
        const next = queue.shift()!;
        inFlight++;
        uploadFile(next).finally(() => {
          inFlight--;
          processNext();
        });
      }
      if (inFlight === 0) {
        isUploading = false;
      }
    }

    processNext();
  }

  function uploadFile(state: FileState): Promise<void> {
    return new Promise((resolve) => {
      state.status = 'uploading';
      const xhr = new XMLHttpRequest();
      xhr.timeout = 600_000;

      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
          state.progress = Math.round((e.loaded / e.total) * 100);
        }
      };

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          state.status = 'done';
          state.progress = 100;
        } else {
          state.status = 'error';
          state.error = `Upload failed (${xhr.status})`;
        }
        resolve();
      };

      xhr.onerror = () => {
        state.status = 'error';
        state.error = 'Network error';
        resolve();
      };

      xhr.ontimeout = () => {
        state.status = 'error';
        state.error = 'Upload timed out';
        resolve();
      };

      xhr.open('PUT', state.uploadUrl!, true);
      xhr.setRequestHeader('Content-Type', state.file.type);
      xhr.send(state.file);
    });
  }

  async function retryFile(state: FileState) {
    state.status = 'pending';
    state.progress = 0;
    state.error = undefined;
    await startUpload();
  }

  const allDone = $derived(fileStates.every((fs) => fs.status === 'done'));
  const uploadedCount = $derived(fileStates.filter((fs) => fs.status === 'done').length);
  const pendingCount = $derived(fileStates.filter((fs) => fs.status === 'pending').length);
  const totalProgress = $derived(
    Math.round(fileStates.reduce((sum, fs) => sum + fs.progress, 0) / fileStates.length)
  );
</script>

<div bind:this={container} class="w-full">
  {#if validationErrors.length > 0}
    <div class="mb-4 rounded-2xl bg-red-50 p-4 text-sm text-red-700">
      {#each validationErrors as err}
        <p>{err}</p>
      {/each}
    </div>
  {/if}

  <div class="flex flex-col gap-2">
    {#if isUploading || fileStates.some((fs) => fs.status === 'done')}
      <div class="mb-1 flex items-center gap-3 rounded-2xl bg-surface-alt px-4 py-2.5">
        <div class="h-2 flex-1 overflow-hidden rounded-full bg-ink/10">
          {#key totalProgress}
            <div class="h-full rounded-full bg-accent transition-all duration-300 ease-out" style="width: {totalProgress}%"></div>
          {/key}
        </div>
        <span class="shrink-0 text-xs font-medium text-body">{uploadedCount}/{files.length} ({totalProgress}%)</span>
      </div>
    {/if}
    {#each fileStates as state}
      <div class="flex items-center gap-3 rounded-2xl bg-surface-alt p-3">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-accent/10 text-accent">
          <Iconify icon={state.file.type.startsWith('video') ? 'tabler:video' : state.file.type === 'image/gif' ? 'tabler:gif' : 'tabler:photo'} class="h-4 w-4" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-medium text-ink">{state.file.name}</p>
          <p class="text-xs text-body">{(state.file.size / 1024 / 1024).toFixed(1)} MB</p>
          {#if state.error}
            <p class="text-xs text-red-500">{state.error}</p>
          {/if}
        </div>
        <div class="shrink-0">
          {#if state.status === 'uploading'}
            <ProgressBar value={state.progress} />
          {:else if state.status === 'done'}
            <div class="flex h-7 w-7 items-center justify-center rounded-full bg-ill-mint">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#0F1729" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
            </div>
          {:else if state.status === 'error'}
            <button onclick={() => retryFile(state)} class="text-xs font-medium text-red-500 hover:opacity-80">Retry</button>
          {:else}
            <span class="text-xs text-body">Pending</span>
          {/if}
        </div>
      </div>
    {/each}
  </div>

  <div class="mt-6 flex gap-3">
    <Button variant="secondary" onclick={onBack}>Back</Button>
    {#if allDone}
      <Button variant="accent" onclick={onProceed}>Continue</Button>
    {:else}
      <Button variant="accent" onclick={startUpload} disabled={isUploading}>
        {isUploading ? 'Uploading...' : `Upload ${pendingCount} file${pendingCount !== 1 ? 's' : ''}`}
      </Button>
    {/if}
  </div>
</div>
