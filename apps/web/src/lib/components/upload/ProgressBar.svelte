<script lang="ts">
  import gsap from 'gsap';

  let {
    value = 0,
  }: {
    value?: number;
  } = $props();

  let fill: HTMLDivElement;

  $effect(() => {
    if (!fill) return;
    const ctx = gsap.context(() => {
      gsap.to(fill, { width: `${value}%`, duration: 0.3, ease: 'power2.out' });
    }, fill);
    return () => ctx.revert();
  });
</script>

<div class="h-2 w-24 overflow-hidden rounded-full bg-surface-alt">
  <div bind:this={fill} class="h-full rounded-full bg-accent" style="width: 0%"></div>
</div>
