<script lang="ts">
  import gsap from 'gsap';
  import Iconify from '@iconify/svelte';
  import Button from '$components/ui/Button.svelte';
  import Input from '$components/ui/Input.svelte';
  import { API_BASE } from '$lib/api/env';

  let {
    slug,
    onUnlocked,
  }: {
    slug: string;
    onUnlocked: () => void;
  } = $props();

  let password = $state('');
  let error = $state('');
  let isSubmitting = $state(false);
  let card: HTMLDivElement;

  async function submit() {
    isSubmitting = true;
    error = '';
    try {
      const resp = await fetch(
        `${API_BASE}/api/g/${slug}/unlock`,
        { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ password }) }
      );
      if (!resp.ok) {
        const body = await resp.json();
        throw new Error(body.error || 'Wrong password');
      }
      onUnlocked();
    } catch (err) {
      error = (err as Error).message;
      if (card) {
        gsap.fromTo(card, { x: -10 }, { x: 10, duration: 0.1, repeat: 5, yoyo: true, ease: 'power2.inOut' });
      }
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="mx-auto max-w-sm text-center">
  <div class="mb-6 flex justify-center">
    <div class="flex h-16 w-16 items-center justify-center rounded-full bg-accent/10">
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-accent"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
    </div>
  </div>

  <p class="text-lg font-semibold text-ink">Password protected</p>
  <p class="mt-2 text-sm text-body">Enter the password to view the files.</p>

  <div bind:this={card} class="mt-6 flex flex-col gap-3">
    <Input bind:value={password} type="password" label="Password" {error} placeholder="Enter password" />
    <Button variant="accent" onclick={submit} disabled={isSubmitting}>
      {isSubmitting ? 'Checking...' : 'Unlock'}
    </Button>
  </div>
</div>
