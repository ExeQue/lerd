<script lang="ts">
  import Modal from './Modal.svelte';
  import Icon from './Icon.svelte';
  import SiteDebugTab from '$tabs/sites/SiteDebugTab.svelte';
  import { profileKeyFor } from '$stores/profiler';
  import { openProfilerReport } from '$stores/dashboard';
  import { sites as allSites, type Site } from '$stores/sites';
  import type { RecentRequest } from '$stores/analytics';
  import type { LinkedRequest } from '$stores/debugEvents';
  import type { ServedRequest } from '$lib/requestWaterfall';
  import { m } from '../paraglide/messages.js';

  // One recent request, through the same lenses as the Debug tab, pinned to
  // what it ran and opening on its timeline. A linked request, the page that
  // sent it or one it sent, opens in its place with a way back.

  interface Props {
    site: Site;
    request: RecentRequest | null;
    onclose: () => void;
  }
  let { site, request, onclose }: Props = $props();

  interface Shown {
    rid: string;
    site: Site;
    served?: ServedRequest;
  }
  // The access log records a request as it ends, so it started that long before.
  const first = $derived<Shown | null>(
    request?.rid ? { rid: request.rid, site, served: { label: `${request.method} ${request.uri}`, start: request.at_millis - request.millis, millis: request.millis } } : null
  );
  let trail = $state<Shown[]>([]);
  $effect(() => {
    void request;
    trail = [];
  });
  const shown = $derived(trail.at(-1) ?? first);

  // A request a page sent may have gone to another site, which its URL names.
  function siteOf(url: string): Site {
    try {
      const host = new URL(url, `http://${site.domain}`).hostname;
      return $allSites.find((s) => s.domain === host || s.domains?.includes(host)) ?? site;
    } catch {
      return site;
    }
  }
  function open(r: LinkedRequest) {
    const url = r.label.replace(/^[A-Z]+\s+/, '');
    const served = r.millis ? { label: r.label, start: r.start ?? 0, millis: r.millis } : undefined;
    trail = [...trail, { rid: r.rid, site: siteOf(url), served }];
  }

  let profileKey = $state('');
  $effect(() => {
    const rid = shown?.rid ?? '';
    profileKey = '';
    if (rid) void profileKeyFor(rid).then((k) => (rid === shown?.rid ? (profileKey = k) : undefined));
  });
</script>

<Modal open={request !== null} title={shown?.served?.label ?? shown?.rid ?? ''} size="2xl" {onclose}>
  {#if shown}
    <div class="h-[75vh] flex flex-col">
      {#if trail.length || profileKey}
        <div class="flex items-center gap-3 px-5 pt-2 -mb-1 text-xs">
          {#if trail.length}
            <button type="button" onclick={() => (trail = trail.slice(0, -1))} class="inline-flex items-center gap-1 text-gray-500 dark:text-gray-400 hover:text-lerd-red">
              <Icon name="back" class="w-3.5 h-3.5" />{m.linked_back()}
            </button>
          {/if}
          {#if profileKey}
            <button type="button" onclick={() => { onclose(); openProfilerReport(profileKey); }} class="ml-auto inline-flex items-center gap-1.5 text-gray-500 dark:text-gray-400 hover:text-lerd-red">
              <Icon name="clock" class="w-3.5 h-3.5" />{m.timeline_flameGraph()}
            </button>
          {/if}
        </div>
      {/if}
      <div class="flex-1 min-h-0">
        {#key shown.rid}
          <SiteDebugTab siteName={shown.site.name} framework={shown.site.framework} domain={shown.site.domain} phpLenses={Boolean(shown.site.uses_php)} rid={shown.rid} served={shown.served} onopen={open} />
        {/key}
      </div>
    </div>
  {/if}
</Modal>
