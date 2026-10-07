<script lang="ts">
  import Modal from './Modal.svelte';
  import Icon from './Icon.svelte';
  import SiteDebugTab from '$tabs/sites/SiteDebugTab.svelte';
  import { profileKeyFor } from '$stores/profiler';
  import { openProfilerReport } from '$stores/dashboard';
  import type { RecentRequest } from '$stores/analytics';
  import type { Site } from '$stores/sites';
  import { m } from '../paraglide/messages.js';

  // One recent request, through the same lenses as the Debug tab, pinned to
  // what it ran and opening on its timeline.

  interface Props {
    site: Site;
    request: RecentRequest | null;
    onclose: () => void;
  }
  let { site, request, onclose }: Props = $props();

  // The access log records a request as it ends, so it started that long before.
  const served = $derived(request ? { label: `${request.method} ${request.uri}`, start: request.at_millis - request.millis, millis: request.millis } : undefined);

  let profileKey = $state('');
  $effect(() => {
    const rid = request?.rid ?? '';
    profileKey = '';
    if (rid) void profileKeyFor(rid).then((k) => (rid === request?.rid ? (profileKey = k) : undefined));
  });
</script>

<Modal open={request !== null} title={served?.label ?? ''} size="2xl" {onclose}>
  {#if request?.rid}
    <div class="h-[75vh] flex flex-col">
      {#if profileKey}
        <div class="flex justify-end px-5 pt-2 -mb-1">
          <button type="button" onclick={() => { onclose(); openProfilerReport(profileKey); }} class="inline-flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400 hover:text-lerd-red">
            <Icon name="clock" class="w-3.5 h-3.5" />{m.timeline_flameGraph()}
          </button>
        </div>
      {/if}
      <div class="flex-1 min-h-0">
        {#key request.rid}
          <SiteDebugTab siteName={site.name} framework={site.framework} domain={site.domain} phpLenses={Boolean(site.uses_php)} rid={request.rid} {served} />
        {/key}
      </div>
    </div>
  {/if}
</Modal>
