<script lang="ts">
  import gsap from 'gsap';

  let openFaq: number | null = $state(null);
  let hero: HTMLDivElement;

  const faqs = [
    { q: 'What is xians?', a: 'Upload files, get a shareable link. No accounts, no compression, no hassle. That is the whole thing.' },
    { q: 'How long do links last?', a: 'You choose: 24 hours, 7 days, 30 days, or keep them forever. You can also limit by download count.' },
    { q: 'Is there a file size limit?', a: 'Each file can be up to 100MB. You can upload up to 20 files per batch.' },
    { q: 'Can I password-protect my links?', a: 'Yes. Add a password during upload. Only people with the password can access the files.' },
    { q: 'How do I delete my files?', a: 'Save the owner token shown after upload. Use it to delete files anytime before they expire.' },
  ];

  $effect(() => {
    const ctx = gsap.context(() => {
      gsap.fromTo('.animate-up', { y: 24, opacity: 0 }, { y: 0, opacity: 1, duration: 0.6, stagger: 0.12, ease: 'power3.out' });
    }, hero);
    return () => ctx.revert();
  });
</script>

<svelte:head>
  <title>xians — Share files. Get a link. Done.</title>
</svelte:head>

<nav class="sticky top-0 z-50 border-b border-surface-alt bg-surface/90 backdrop-blur-md">
  <div class="container-narrow flex items-center justify-between py-4">
    <a href="/" class="font-ui text-lg font-bold text-ink">xians</a>
    <div class="flex items-center gap-6">
      <a href="#features" class="hidden text-sm font-medium text-body hover:text-ink md:block">Features</a>
      <a href="#faq" class="hidden text-sm font-medium text-body hover:text-ink md:block">FAQ</a>
      <a href="/upload" class="btn-primary !px-4 !py-2">Get started</a>
    </div>
  </div>
</nav>

<section bind:this={hero} class="overflow-hidden bg-surface pt-16 pb-20 md:pt-24 md:pb-28">
  <div class="container-narrow flex flex-col items-center text-center">
    <div class="animate-up mb-8 flex items-center gap-3">
      <svg width="56" height="56" viewBox="0 0 56 56" fill="none">
        <rect x="2" y="2" width="52" height="52" rx="26" fill="#AEDFF7"/>
        <path d="M20 28h16M28 20v16" stroke="#0F1729" stroke-width="2.5" stroke-linecap="round"/>
        <circle cx="28" cy="28" r="22" stroke="#0F1729" stroke-width="1.5"/>
      </svg>
      <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
        <rect x="2" y="2" width="60" height="60" rx="30" fill="#2F6FEB"/>
        <path d="M22 42l7-7 7 7M36 22l-7 7-7-7" stroke="white" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="M29 29v14M35 29v14" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
        <circle cx="32" cy="32" r="26" stroke="#0F1729" stroke-width="1.5"/>
      </svg>
      <svg width="56" height="56" viewBox="0 0 56 56" fill="none">
        <rect x="2" y="2" width="52" height="52" rx="26" fill="#F5A742"/>
        <path d="M18 36c0-6 4-10 10-10s10 4 10 10" stroke="#0F1729" stroke-width="2.5" stroke-linecap="round"/>
        <circle cx="28" cy="22" r="6" stroke="#0F1729" stroke-width="2.5"/>
      </svg>
    </div>

    <div class="animate-up mb-8 inline-flex items-center gap-2 rounded-full bg-accent-tint px-5 py-2.5 text-sm font-medium text-accent">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="2" x2="12" y2="22"/><polyline points="19 15 12 22 5 15"/></svg>
      Say hello to xians — your file companion
    </div>

    <h1 class="animate-up heading-xl max-w-5xl text-ink">
      Share files.<br/>Get a link.<br/><span class="text-accent">Done.</span>
    </h1>

    <p class="animate-up mt-8 max-w-lg text-base leading-relaxed text-body">
      Upload files and get a shareable link instantly. No accounts, no compression, no quality loss.
    </p>

    <div class="animate-up mt-10 flex w-full max-w-xs flex-col gap-4">
      <a href="/upload" class="btn-primary w-full !py-3.5">
        Start uploading
      </a>
      <a href="#how-it-works" class="btn-secondary w-full !py-3.5">
        How it works
      </a>
    </div>
  </div>
</section>

<section class="relative z-10 -mt-4 pb-24">
  <div class="container-narrow">
    <a href="/upload" class="block rounded-3xl bg-surface-alt p-8 md:p-12 transition-all hover:bg-surface-alt/80">
      <div class="mx-auto max-w-md text-center">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="mx-auto mb-4 text-accent">
          <rect x="2" y="2" width="20" height="20" rx="4" stroke="currentColor"/>
          <path d="M12 8v8M8 12h8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <p class="font-display text-xl text-ink uppercase">Drop files here to start</p>
        <p class="mt-2 text-sm text-body">or click to browse</p>
      </div>
    </a>
  </div>
</section>

<section class="bg-surface-alt py-24 md:py-32" id="features">
  <div class="container-narrow">
    <div class="mb-16">
      <h2 class="heading-lg text-ink">More than file sharing</h2>
      <p class="mt-6 max-w-lg text-base leading-relaxed text-body">
        Upload files, share links, manage access, and track downloads. Everything you need, nothing you do not.
      </p>
    </div>

    <div class="space-y-8">
      <div class="overflow-hidden rounded-3xl bg-surface">
        <div class="flex aspect-[16/9] items-center justify-center bg-surface-alt md:aspect-[21/9]">
          <svg width="200" height="120" viewBox="0 0 200 120" fill="none" class="md:scale-150">
            <rect x="10" y="40" width="70" height="50" rx="8" fill="#AEDFF7"/>
            <rect x="14" y="46" width="62" height="6" rx="3" fill="white" opacity="0.5"/>
            <rect x="14" y="58" width="40" height="4" rx="2" fill="white" opacity="0.7"/>
            <rect x="14" y="68" width="50" height="4" rx="2" fill="white" opacity="0.5"/>
            <path d="M120 65l20-20 20 20" stroke="#2F6FEB" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M140 45v36" stroke="#2F6FEB" stroke-width="3" stroke-linecap="round"/>
            <circle cx="160" cy="40" r="20" fill="#7FD8C4"/>
            <path d="M152 44l4-4 4 4" stroke="#0F1729" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M156 40v10" stroke="#0F1729" stroke-width="2.5" stroke-linecap="round"/>
            <circle cx="40" cy="30" r="16" fill="#F5A742"/>
            <path d="M34 34l4-4 4 4" stroke="#0F1729" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M38 30v8" stroke="#0F1729" stroke-width="2" stroke-linecap="round"/>
            <circle cx="100" cy="20" r="8" fill="white" stroke="#0F1729" stroke-width="1.5"/>
            <circle cx="70" cy="18" r="5" fill="white" stroke="#0F1729" stroke-width="1.5"/>
            <circle cx="130" cy="22" r="6" fill="white" stroke="#0F1729" stroke-width="1.5"/>
          </svg>
        </div>
        <div class="p-8 md:p-12">
          <h3 class="font-display text-2xl font-normal text-ink uppercase">Built for speed</h3>
          <p class="mt-3 max-w-lg text-sm text-body">Direct-to-cloud uploads mean no waiting. Files are served at full quality, every time.</p>
        </div>
      </div>

      <div class="overflow-hidden rounded-3xl bg-surface">
        <div class="flex aspect-[16/9] items-center justify-center bg-surface-alt md:aspect-[21/9]">
          <svg width="200" height="120" viewBox="0 0 200 120" fill="none" class="md:scale-150">
            <rect x="20" y="20" width="100" height="80" rx="12" fill="#AEDFF7"/>
            <rect x="16" y="16" width="108" height="12" rx="6" fill="#7FD8C4"/>
            <rect x="28" y="40" width="36" height="8" rx="4" fill="white" opacity="0.6"/>
            <rect x="28" y="54" width="60" height="6" rx="3" fill="white" opacity="0.5"/>
            <rect x="28" y="66" width="48" height="6" rx="3" fill="white" opacity="0.4"/>
            <circle cx="150" cy="60" r="28" fill="#F5A742"/>
            <path d="M142 60l6-6 6 6M148 54v12" stroke="#0F1729" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <circle cx="148" cy="60" r="24" stroke="#0F1729" stroke-width="1.5"/>
            <circle cx="82" cy="76" r="14" fill="white" stroke="#0F1729" stroke-width="1.5"/>
            <path d="M78 80l4-4 4 4" stroke="#0F1729" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M82 76v6" stroke="#0F1729" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <div class="p-8 md:p-12">
          <h3 class="font-display text-2xl font-normal text-ink uppercase">Your files, your rules</h3>
          <p class="mt-3 max-w-lg text-sm text-body">Password protection, expiring links, download limits. Full control over how your files are shared.</p>
        </div>
      </div>

      <div class="overflow-hidden rounded-3xl bg-surface">
        <div class="flex aspect-[16/9] items-center justify-center bg-surface-alt md:aspect-[21/9]">
          <svg width="200" height="120" viewBox="0 0 200 120" fill="none" class="md:scale-150">
            <circle cx="60" cy="50" r="30" fill="#AEDFF7"/>
            <path d="M48 56l8-8 8 8" stroke="#2F6FEB" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M56 48v16" stroke="#2F6FEB" stroke-width="3" stroke-linecap="round"/>
            <rect x="100" y="30" width="60" height="50" rx="10" fill="#F5A742"/>
            <path d="M110 50h40M110 62h30" stroke="white" stroke-width="3" stroke-linecap="round"/>
            <rect x="110" y="38" width="20" height="6" rx="3" fill="white" opacity="0.5"/>
            <circle cx="70" cy="30" r="8" fill="white" stroke="#0F1729" stroke-width="1.5"/>
            <circle cx="140" cy="26" r="6" fill="white" stroke="#0F1729" stroke-width="1.5"/>
            <path d="M160 80l10-10 10 10" stroke="#7FD8C4" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M170 70v16" stroke="#7FD8C4" stroke-width="2.5" stroke-linecap="round"/>
            <circle cx="185" cy="60" r="6" fill="#7FD8C4"/>
          </svg>
        </div>
        <div class="p-8 md:p-12">
          <h3 class="font-display text-2xl font-normal text-ink uppercase">Simple as it gets</h3>
          <p class="mt-3 max-w-lg text-sm text-body">No accounts, no sign-ups, no tracking. Just upload and share. That is the whole promise.</p>
        </div>
      </div>
    </div>
  </div>
</section>

<section class="bg-surface py-24 md:py-32" id="how-it-works">
  <div class="container-narrow">
    <div class="mb-16">
      <h2 class="heading-lg text-ink">How it works</h2>
      <p class="mt-6 max-w-lg text-base leading-relaxed text-body">Three steps. That is it.</p>
    </div>

    <div class="space-y-16">
      <div class="flex flex-col gap-6 md:flex-row md:items-start md:gap-12">
        <div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-accent text-base font-bold text-white">1</div>
        <div>
          <h3 class="font-display text-xl font-normal text-ink uppercase">Drop your files</h3>
          <p class="mt-2 max-w-md text-sm text-body">Drag, click, or paste. Works on any device. Up to 20 files, 100MB each.</p>
        </div>
      </div>
      <div class="flex flex-col gap-6 md:flex-row md:items-start md:gap-12">
        <div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-ink text-base font-bold text-white">2</div>
        <div>
          <h3 class="font-display text-xl font-normal text-ink uppercase">Get your link</h3>
          <p class="mt-2 max-w-md text-sm text-body">Set options like password or expiry. Copy the shareable link.</p>
        </div>
      </div>
      <div class="flex flex-col gap-6 md:flex-row md:items-start md:gap-12">
        <div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-ill-mint text-base font-bold text-ink">3</div>
        <div>
          <h3 class="font-display text-xl font-normal text-ink uppercase">Share it</h3>
          <p class="mt-2 max-w-md text-sm text-body">Send the link anywhere. Recipients download at full quality, no sign-up required.</p>
        </div>
      </div>
    </div>
  </div>
</section>

<section class="bg-surface-alt py-24 md:py-32" id="faq">
  <div class="container-narrow max-w-3xl">
    <h2 class="heading-lg text-ink mb-16">Frequently asked questions</h2>

    <div class="space-y-0">
      {#each faqs as faq, i}
        <div class="border-b border-surface">
          <button
            onclick={() => openFaq = openFaq === i ? null : i}
            class="flex w-full items-center justify-between py-6 text-left"
          >
            <span class="pr-4 text-base font-medium text-ink">{faq.q}</span>
            <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-accent text-white transition-transform duration-200 {openFaq === i ? 'rotate-45' : ''}">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </span>
          </button>
          {#if openFaq === i}
            <div class="pb-6 text-sm leading-relaxed text-body">{faq.a}</div>
          {/if}
        </div>
      {/each}
    </div>

    <a href="#top" class="mt-8 inline-flex items-center gap-1.5 text-sm font-medium text-accent hover:opacity-80">
      See more FAQs
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
    </a>
  </div>
</section>

<section class="bg-surface-alt py-24 md:py-32">
  <div class="container-narrow">
    <div class="max-w-2xl">
      <h2 class="heading-lg text-ink">Ready to share?</h2>
      <p class="mt-6 max-w-lg text-base leading-relaxed text-body">
        Join people who are already using xians to share files simply and securely.
      </p>
      <div class="mt-8">
        <a href="/upload" class="btn-primary !px-8 !py-3.5">
          Get started
        </a>
      </div>
    </div>
  </div>
</section>

<footer class="border-t border-surface-alt bg-surface py-10">
  <div class="container-narrow flex flex-col items-start gap-5 md:flex-row md:items-center md:justify-between">
    <span class="font-ui text-base font-bold text-ink">xians</span>
    <div class="flex gap-6 text-sm text-body">
      <a href="#features" class="hover:text-ink">Features</a>
      <a href="#how-it-works" class="hover:text-ink">How it works</a>
      <a href="#faq" class="hover:text-ink">FAQ</a>
    </div>
    <div class="flex flex-col items-start gap-1 md:items-end">
      <p class="text-xs text-body">Files served at full quality, always.</p>
      <p class="text-xs text-body">
        Built by <a href="https://habibthadev.tech" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline">habibthadev</a>
      </p>
    </div>
  </div>
</footer>
