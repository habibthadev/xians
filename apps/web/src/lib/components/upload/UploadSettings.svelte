<script lang="ts">
  import gsap from 'gsap';
  import Button from '$components/ui/Button.svelte';
  import Input from '$components/ui/Input.svelte';
  import { API_BASE } from '$lib/api/env';

  let {
    batchId,
    onComplete,
    onBack,
  }: {
    batchId: string;
    onComplete: (slug: string, ownerToken: string) => void;
    onBack: () => void;
  } = $props();

  let ttl: string = $state('7d');
  let password: string = $state('');
  let maxDownloads: string = $state('');
  let isSubmitting = $state(false);
  let error = $state('');

  let container: HTMLDivElement;

  const ttlOptions = ['24h', '7d', '30d', 'never'] as const;

  function ttlLabel(opt: string): string {
    if (opt === '24h') return '24 hours';
    if (opt === '7d') return '7 days';
    if (opt === '30d') return '30 days';
    return 'Never';
  }

  $effect(() => {
    const ctx = gsap.context(() => {
      gsap.fromTo(container, { y: 20, opacity: 0 }, { y: 0, opacity: 1, duration: 0.4, ease: 'power3.out' });
    }, container);
    return () => ctx.revert();
  });

  async function complete() {
    isSubmitting = true;
    try {
      const resp = await fetch(
        `${API_BASE}/api/batches/${batchId}/complete`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            ttl,
            password: password || undefined,
            maxDownloads: maxDownloads ? parseInt(maxDownloads) : undefined,
          }),
        }
      );
      if (!resp.ok) throw new Error((await resp.json()).error || 'Complete failed');
      const data = await resp.json();
      onComplete(data.slug, data.ownerToken);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Something went wrong';
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div bind:this={container} class="w-full">
  <p class="text-lg font-semibold text-ink">Link settings</p>

  <div class="mt-6">
      <span class="mb-3 block text-sm font-medium text-ink">Link expires after</span>
    <div class="flex flex-wrap gap-2">
      {#each ttlOptions as option}
        <button
          onclick={() => { ttl = option; }}
          class="rounded-full border-2 px-4 py-2 text-sm font-semibold transition-all duration-200 {ttl === option ? 'border-ink bg-ink text-white' : 'border-surface-alt bg-transparent text-body hover:border-ink/30 hover:text-ink'}"
        >
          {ttlLabel(option)}
        </button>
      {/each}
    </div>
  </div>

  <div class="mt-6 flex flex-col gap-4">
    <Input bind:value={password} type="password" label="Password (optional)" placeholder="Leave empty for no password" />
    <Input bind:value={maxDownloads} type="number" label="Max downloads (optional)" placeholder="No limit" />
  </div>

  <div class="mt-6 flex gap-3">
    {#if error}
      <p class="mb-2 text-sm text-red-500">{error}</p>
    {/if}
    <Button variant="secondary" onclick={onBack}>Back</Button>
    <Button variant="accent" onclick={complete} disabled={isSubmitting}>
      {isSubmitting ? 'Creating...' : 'Get link'}
    </Button>
  </div>
</div>
