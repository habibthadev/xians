<script lang="ts">
  import { page } from '$app/state';

  let message = $derived(
    page.status === 404 ? "We couldn't find that page." :
    page.status === 410 ? 'This link has expired or been removed.' :
    'Something went wrong.'
  );
</script>

<div class="flex min-h-dvh flex-col items-center justify-center gap-8 px-5 text-center">
  <div class="flex h-28 w-28 items-center justify-center rounded-full bg-surface-alt">
    <svg width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-body"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
  </div>
  <h1 class="font-display text-8xl font-normal text-ink">{page.status}</h1>
  <p class="max-w-md text-base text-body">{message}</p>
  <a href="/" class="btn-primary mt-4 gap-2">
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/></svg>
    Go home
  </a>
</div>
