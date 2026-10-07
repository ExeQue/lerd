<script lang="ts">
  import Icon from './Icon.svelte';
  import { lensEvents, sentBy, sentRequests, type LinkedRequest } from '$stores/debugEvents';
  import { m } from '../paraglide/messages.js';

  // The page a request came from and the requests it sent, each opening in
  // place of this one.

  interface Props {
    rid: string;
    onopen?: (r: LinkedRequest) => void;
  }
  let { rid, onopen }: Props = $props();

  const debugEvents = lensEvents();
  const parent = $derived(sentBy($debugEvents, rid));
  const sent = $derived(sentRequests($debugEvents, rid));

  function tone(status?: number) {
    if (!status) return 'text-red-600 dark:text-red-400';
    if (status >= 500) return 'text-red-600 dark:text-red-400';
    if (status >= 400) return 'text-amber-600 dark:text-amber-400';
    return 'text-emerald-600 dark:text-emerald-400';
  }
</script>

{#snippet row(r: LinkedRequest)}
  <button
    type="button"
    disabled={!onopen}
    onclick={() => onopen?.(r)}
    class="w-full flex items-center gap-3 px-2.5 py-1.5 text-left text-xs hover:bg-gray-50 dark:hover:bg-white/5 disabled:cursor-default disabled:hover:bg-transparent"
  >
    <span class="font-mono text-gray-800 dark:text-gray-200 truncate flex-1 min-w-0">{r.label}</span>
    {#if r.status !== undefined}<span class="shrink-0 font-mono font-semibold {tone(r.status)}">{r.status || '·'}</span>{/if}
    {#if r.millis}<span class="shrink-0 w-16 text-right tabular-nums text-gray-500 dark:text-gray-400">{Math.round(r.millis)} ms</span>{/if}
    {#if onopen}<Icon name="chevron" class="w-3.5 h-3.5 shrink-0 -rotate-90 text-gray-400" />{/if}
  </button>
{/snippet}

<div class="h-full overflow-y-auto p-3 space-y-4">
  {#if parent}
    <section class="space-y-1">
      <h4 class="text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">{m.linked_sentBy()}</h4>
      <div class="rounded-md border border-gray-200 dark:border-lerd-border">{@render row(parent)}</div>
    </section>
  {/if}
  {#if sent.length}
    <section class="space-y-1">
      <h4 class="text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">{m.linked_sent({ count: sent.length })}</h4>
      <div class="rounded-md border border-gray-200 dark:border-lerd-border divide-y divide-gray-100 dark:divide-lerd-border/60">
        {#each sent as r, i (r.rid + i)}{@render row(r)}{/each}
      </div>
    </section>
  {/if}
  {#if !parent && !sent.length}
    <p class="p-6 text-center text-xs text-gray-500 dark:text-gray-400">{m.linked_empty()}</p>
  {/if}
</div>
