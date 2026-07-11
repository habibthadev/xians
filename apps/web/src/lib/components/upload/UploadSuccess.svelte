<script lang="ts">
  import gsap from 'gsap';
  import Iconify from '@iconify/svelte';
  import Button from '$components/ui/Button.svelte';
  import Toast from '$components/ui/Toast.svelte';

  let {
    slug,
    ownerToken,
    onNewUpload,
  }: {
    slug: string;
    ownerToken: string;
    onNewUpload: () => void;
  } = $props();

  let link = $derived(`${window.location.origin}/g/${slug}`);
  let copied = $state(false);
  let toastVisible = $state(false);
  let toastMessage = $state('');
  let copyBtn: HTMLButtonElement;
  let checkIcon: HTMLDivElement;
  let root: HTMLDivElement;

  $effect(() => {
    const ctx = gsap.context(() => {
      gsap.fromTo('#success-root > *', { y: 20, opacity: 0 }, { y: 0, opacity: 1, duration: 0.4, stagger: 0.1, ease: 'power3.out' });
    }, root);
    return () => ctx.revert();
  });

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(link);
      copied = true;
      gsap.fromTo(checkIcon, { scale: 1.5, opacity: 0 }, { scale: 1, opacity: 1, duration: 0.3, ease: 'back.out(2)' });
      toastMessage = 'Link copied!';
      toastVisible = true;
    } catch {
      toastMessage = 'Copy failed';
      toastVisible = true;
    }
  }
</script>

<div bind:this={root} id="success-root" class="w-full text-center">
  <div class="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-ill-mint">
    <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="#0F1729" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
  </div>

  <p class="text-lg font-semibold text-ink">Your link is ready</p>
  <p class="mt-3 text-sm text-body">Share this link anywhere. The files will be delivered at full quality.</p>

  <div class="mx-auto mt-8 flex max-w-lg items-center gap-3 rounded-2xl bg-surface-alt p-4">
    <code class="flex-1 truncate text-left text-sm font-medium text-ink">{link}</code>
    <button
      bind:this={copyBtn}
      onclick={copyLink}
      class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-ink text-white hover:opacity-90"
    >
      <div bind:this={checkIcon}>
        {#if !copied}
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
        {:else}
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
        {/if}
      </div>
    </button>
  </div>

  <div class="mx-auto mt-6 max-w-lg rounded-2xl bg-ill-orange/10 p-5 text-left">
    <p class="mb-1 font-semibold text-sm text-ink">Save this to delete later</p>
    <p class="font-mono text-xs break-all text-body">{ownerToken}</p>
    <p class="mt-2 text-xs text-body">We do not store this anywhere you can retrieve it again. Save it now or lose it.</p>
  </div>

  <div class="mt-8">
    <Button onclick={onNewUpload}>Upload another batch</Button>
  </div>
</div>

<Toast bind:visible={toastVisible} message={toastMessage} type="success" />
