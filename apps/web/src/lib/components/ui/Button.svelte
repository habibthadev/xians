<script lang="ts">
  import type { Snippet } from 'svelte';
  import gsap from 'gsap';

  let {
    onclick,
    variant = 'primary',
    disabled = false,
    type = 'button' as const,
    children,
  }: {
    onclick?: () => void;
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'accent';
    disabled?: boolean;
    type?: 'button' | 'submit' | 'reset';
    children: Snippet;
  } = $props();

  let btn: HTMLButtonElement;

  $effect(() => {
    const ctx = gsap.context(() => {
      gsap.set(btn, { transformStyle: 'preserve-3d', willChange: 'transform' });
    }, btn);
    return () => ctx.revert();
  });

  function handleClick(e: MouseEvent) {
    if (disabled) return;
    gsap.fromTo(btn, { scale: 0.97 }, { scale: 1, duration: 0.15, ease: 'power2.out' });
    onclick?.();
  }
</script>

<button
  bind:this={btn}
  {type}
  {disabled}
  {onclick}
  class={[
    'inline-flex items-center justify-center gap-2 rounded-full text-sm font-semibold transition-all duration-200',
    variant === 'primary' && 'btn-primary',
    variant === 'secondary' && 'btn-secondary',
    variant === 'accent' && 'btn-accent',
    variant === 'ghost' && 'rounded-full px-4 py-2 text-body hover:bg-surface-alt hover:text-ink',
    variant === 'danger' && 'rounded-full bg-red-600 px-6 py-3 text-white hover:opacity-90',
    disabled && 'opacity-50 cursor-not-allowed',
    !disabled && 'active:scale-[0.97]',
    !variant?.match(/^(primary|secondary|accent)$/) && 'px-6 py-3',
  ]}
>
  {@render children()}
</button>
