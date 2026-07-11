<script lang="ts">
  let {
    value = $bindable(''),
    placeholder = '',
    type = 'text',
    label = '',
    error = '',
    oninput,
  }: {
    value?: string;
    placeholder?: string;
    type?: string;
    label?: string;
    error?: string;
    oninput?: (e: Event) => void;
  } = $props();

  let inputId = `input-${Math.random().toString(36).slice(2, 8)}`;
</script>

<div class="flex flex-col gap-2">
  {#if label}
    <label for={inputId} class="text-sm font-medium text-ink">{label}</label>
  {/if}
  <input
    {type}
    {placeholder}
    bind:value
    {oninput}
    id={inputId}
    class="input-base {error ? '!border-red-500 focus:!ring-red-500/20' : ''}"
  />
  {#if error}
    <p class="text-xs font-medium text-red-500">{error}</p>
  {/if}
</div>
