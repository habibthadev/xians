<script lang="ts">
  import gsap from 'gsap';

  let {
    message = '',
    type = 'info',
    visible = $bindable(false),
  }: {
    message?: string;
    type?: 'info' | 'success' | 'error';
    visible?: boolean;
  } = $props();

  let el: HTMLDivElement = $state(null!);

  $effect(() => {
    if (!visible || !el) return;
    const ctx = gsap.context(() => {
      gsap.fromTo(el, { y: 20, opacity: 0 }, { y: 0, opacity: 1, duration: 0.3, ease: 'power3.out' });
    }, el);
    const timer = setTimeout(() => {
      visible = false;
    }, 3000);
    return () => {
      ctx.revert();
      clearTimeout(timer);
    };
  });
</script>

{#if visible}
  <div
    bind:this={el}
    class="fixed bottom-6 left-1/2 z-50 -translate-x-1/2 rounded-full px-6 py-3.5 text-sm font-semibold text-white"
    class:bg-ink={type === 'info'}
    class:bg-green-600={type === 'success'}
    class:bg-red-600={type === 'error'}
  >
    {message}
  </div>
{/if}
