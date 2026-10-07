<script lang="ts">
  import { lensEvents } from '$stores/debugEvents';
  import { m } from '../paraglide/messages.js';

  // What one request carried and answered with, as PHP reported it ending:
  // its headers, query, body and cookies, and the headers it sent back.

  const debugEvents = lensEvents();
  type Values = Record<string, unknown>;
  interface Reported {
    method?: string;
    uri?: string;
    status?: number;
    time_ms?: number;
    memory_peak?: number;
    headers?: Values;
    query?: Values;
    body?: Values;
    cookies?: Values;
    response_headers?: Values;
  }
  const req = $derived(($debugEvents.find((ev) => ev.kind === 'request')?.data ?? null) as Reported | null);

  const show = (v: unknown) => (typeof v === 'string' ? v : JSON.stringify(v));
  const mb = (bytes: number) => `${(bytes / 1048576).toFixed(1)} MB`;
</script>

{#snippet table(title: string, values: Values | undefined)}
  {#if values && Object.keys(values).length}
    <section class="space-y-1">
      <h4 class="text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">{title}</h4>
      <div class="rounded-md border border-gray-200 dark:border-lerd-border divide-y divide-gray-100 dark:divide-lerd-border/60">
        {#each Object.entries(values) as [k, v] (k)}
          <div class="grid grid-cols-[minmax(8rem,14rem)_1fr] gap-3 px-2.5 py-1.5 text-xs">
            <span class="font-mono text-gray-500 dark:text-gray-400 break-all">{k}</span>
            <span class="font-mono text-gray-800 dark:text-gray-200 break-all">{show(v)}</span>
          </div>
        {/each}
      </div>
    </section>
  {/if}
{/snippet}

<div class="h-full overflow-y-auto p-3 space-y-4">
  {#if !req}
    <p class="p-6 text-center text-xs text-gray-500 dark:text-gray-400">{m.requestLens_empty()}</p>
  {:else}
    <div class="flex items-center gap-3 flex-wrap text-xs">
      <span class="font-mono font-semibold text-gray-800 dark:text-gray-100">{req.method} {req.uri}</span>
      {#if req.status}<span class="font-mono font-semibold {req.status >= 500 ? 'text-red-600 dark:text-red-400' : req.status >= 400 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'}">{req.status}</span>{/if}
      {#if req.time_ms}<span class="text-gray-500 dark:text-gray-400">{m.requestLens_php({ time: Math.round(req.time_ms) })}</span>{/if}
      {#if req.memory_peak}<span class="text-gray-500 dark:text-gray-400">{m.requestLens_memory({ size: mb(req.memory_peak) })}</span>{/if}
    </div>
    {@render table(m.requestLens_headers(), req.headers)}
    {@render table(m.requestLens_query(), req.query)}
    {#if req.body && Object.keys(req.body).length}
      <section class="space-y-1">
        <h4 class="text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">{m.requestLens_body()}</h4>
        <pre class="rounded-md border border-gray-200 dark:border-lerd-border px-2.5 py-2 text-xs font-mono text-gray-800 dark:text-gray-200 whitespace-pre-wrap break-all">{JSON.stringify(req.body, null, 2)}</pre>
      </section>
    {/if}
    {@render table(m.requestLens_cookies(), req.cookies)}
    {@render table(m.requestLens_responseHeaders(), req.response_headers)}
  {/if}
</div>
