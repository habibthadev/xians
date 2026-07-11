<script lang="ts">
  import { page } from '$app/state';
  import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
  import '../app.css';

  let { children }: { children: import('svelte').Snippet } = $props();

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: 2, staleTime: 30_000 },
    },
  });

  const siteTitle = 'xians — File sharing, instant';
  const siteDesc = 'Share files instantly with no sign-up, no accounts, no fuss. Just drag, drop, and share.';
  const siteUrl = 'https://xians.vercel.app';
</script>

<svelte:head>
  <title>{siteTitle}</title>
  <meta name="description" content={siteDesc} />
  <meta name="robots" content="index, follow" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta charset="utf-8" />
  <link rel="canonical" href={siteUrl} />
  <meta property="og:title" content={siteTitle} />
  <meta property="og:description" content={siteDesc} />
  <meta property="og:type" content="website" />
  <meta property="og:url" content={siteUrl} />
  <meta property="og:image" content={`${siteUrl}/og?title=${encodeURIComponent('xians')}&description=${encodeURIComponent(siteDesc)}`} />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content={siteTitle} />
  <meta name="twitter:description" content={siteDesc} />
</svelte:head>

<QueryClientProvider client={queryClient}>
  {@render children()}
</QueryClientProvider>
