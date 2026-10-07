<script lang="ts">
  import DetailTabs from '$components/DetailTabs.svelte';
  import { onMount } from 'svelte';
  import DumpsTab from '$tabs/DumpsTab.svelte';
  import QueriesLens from '$components/QueriesLens.svelte';
  import KindLens from '$components/KindLens.svelte';
  import DebugDisabled from '$components/DebugDisabled.svelte';
  import BrowserLens from '$components/BrowserLens.svelte';
  import RequestTimeline from '$components/RequestTimeline.svelte';
  import Dropdown from '$components/Dropdown.svelte';
  import { writable } from 'svelte/store';
  import { apiJson } from '$lib/api';
  import type { DumpEvent } from '$lib/dumpsStream';
  import { buildWaterfall, type ServedRequest } from '$lib/requestWaterfall';
  import { debugLens, debugLensTabs, type DebugLens } from '$stores/debugLens';
  import { refreshStatus, startDumpsStream, stopDumpsStream } from '$stores/dumps';
  import { refreshDevtoolsStatus, debugCaptureEnabled } from '$stores/queries';
  import { countKinds, debugEvents, requestChoices, scopeLensEvents } from '$stores/debugEvents';
  import Icon from '$components/Icon.svelte';
  import { modal } from '$stores/modals';
  import { tooltip } from '$lib/tooltip';
  import { m } from '../../paraglide/messages.js';

  // The lenses hold the event stream open themselves; the timeline reads it
  // too, so the tab keeps it open while it is shown.
  onMount(() => {
    void refreshStatus();
    void refreshDevtoolsStatus();
    startDumpsStream();
    return stopDumpsStream;
  });

  interface Props {
    siteName?: string;
    framework?: string;
    domain?: string;
    branch?: string;
    // phpLenses is false for a site without PHP, which only has browser events.
    phpLenses?: boolean;
    // rid pins the lenses to one request, as a recent request's inspector
    // does; served is that request as nginx timed it, for the timeline.
    rid?: string;
    served?: ServedRequest;
  }
  let { siteName = '', framework = '', domain = '', branch = '', phpLenses = true, rid = '', served }: Props = $props();

  // Without a pinned request, a filter over the lenses picks one, all by
  // default. A picked request gains a timeline of what it did.
  let picked = $state('');
  const scope = writable('');
  $effect(() => scope.set(rid || picked));
  // A pinned or picked request may predate what the stream replayed; the
  // server's ring still holds it.
  const fetched = writable<DumpEvent[]>([]);
  $effect(() => {
    const want = rid || picked;
    fetched.set([]);
    if (want) apiJson<DumpEvent[]>(`/api/dumps?${new URLSearchParams({ rid: want })}`).then((evs) => (want === (rid || picked) ? fetched.set(evs) : undefined), () => {});
  });
  const events = scopeLensEvents(scope, fetched);
  const choices = $derived(rid ? [] : requestChoices($debugEvents, siteName));
  $effect(() => {
    if (picked && !choices.some((c) => c.rid === picked)) picked = '';
  });
  let timeline = $state(Boolean(rid));
  const showTimeline = $derived(timeline && Boolean(rid || picked));
  const waterfall = $derived(showTimeline ? buildWaterfall($events, rid ? served : undefined) : null);

  // Cache comes solely from the Laravel adapter, so it only applies to Laravel
  // sites; everything else is framework-agnostic (PDO and the Symfony
  // Mailer/Twig/EventDispatcher/Messenger/HttpClient seams cover every PHP app).
  const isLaravel = $derived(framework.toLowerCase() === 'laravel');
  const laravelOnly: DebugLens[] = ['cache'];
  const counts = $derived(countKinds($events, siteName));

  const tabs = $derived([
    ...(rid || picked ? [{ id: 'timeline', label: m.debug_tab_timeline(), group: 'timeline' }] : []),
    ...debugLensTabs(counts, isLaravel)
  ]);
  const active = $derived(showTimeline ? 'timeline' : $debugLens);
  function pick(id: string) {
    timeline = id === 'timeline';
    if (!timeline) debugLens.set(id as DebugLens);
  }

  // A site without PHP keeps the lens bar, with Browser as its only lens.
  const browserOnly = $derived(tabs.filter((t) => t.id === 'browser'));

  // Full screen works like Tinker's: transient, and Escape leaves it only once
  // no modal is stacked on top, since the modal owns Escape first.
  let fullscreen = $state(false);
  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && fullscreen && $modal.kind === null && !e.defaultPrevented) {
      e.preventDefault();
      fullscreen = false;
    }
  }

  // If the remembered lens isn't available for this framework, fall back.
  $effect(() => {
    if (!isLaravel && laravelOnly.includes($debugLens)) debugLens.set('queries');
  });
</script>

<svelte:window onkeydown={onKeydown} />

{#snippet fullscreenAction()}
  {#if choices.length > 0}
    <Dropdown
      value={picked}
      options={[{ value: '', label: m.debug_filter_allRequests() }, ...choices.map((c) => ({ value: c.rid, label: c.label, description: new Date(c.ts).toLocaleTimeString([], { hour12: false }) }))]}
      onchange={(v) => (picked = v)}
      title={m.debug_filter_request()}
      align="right"
      minMenuWidth={260}
    />
  {/if}
  <!-- Full screen hides the site header, so name the site here. -->
  {#if !rid && fullscreen && domain}<span class="text-xs font-mono text-gray-600 dark:text-gray-300">{domain}</span>{/if}
  {#if !rid}<button
    type="button"
    onclick={() => (fullscreen = !fullscreen)}
    class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
    use:tooltip={fullscreen ? m.debug_exitFullscreenTitle() : m.debug_fullscreenTitle()}
    aria-label={fullscreen ? m.debug_exitFullscreenTitle() : m.debug_fullscreenTitle()}
  >
    <Icon name={fullscreen ? 'minimize' : 'maximize'} class="w-4 h-4" />
  </button>{/if}
{/snippet}

<div class="flex flex-col overflow-hidden {fullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-lerd-bg' : 'h-full'}">
  {#if !$debugCaptureEnabled && !rid}
    <DebugDisabled />
  {:else if !phpLenses}
    <DetailTabs tabs={browserOnly} active="browser" onchange={() => {}} keepSingle actions={fullscreenAction} />
    <div class="flex-1 min-h-0 overflow-hidden">
      <BrowserLens siteScope={siteName} />
    </div>
  {:else}
    <DetailTabs {tabs} {active} onchange={pick} actions={fullscreenAction} />
    <div class="flex-1 min-h-0 overflow-hidden">
      {#if waterfall}
        <RequestTimeline {waterfall} />
      {:else if $debugLens === 'browser'}
        <BrowserLens siteScope={siteName} />
      {:else if $debugLens === 'dumps'}
        <DumpsTab siteScope={siteName} />
      {:else if $debugLens === 'queries'}
        <QueriesLens siteScope={siteName} />
      {:else}
        <KindLens kind={$debugLens as 'jobs' | 'views' | 'mail' | 'cache' | 'events' | 'http' | 'logs' | 'exceptions' | 'messages'} siteScope={siteName} />
      {/if}
    </div>
  {/if}
</div>
