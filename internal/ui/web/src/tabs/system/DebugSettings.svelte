<script lang="ts">
  import Dropdown from '$components/Dropdown.svelte';
  import { status as dumpsStatusValue, refreshStatus, togglePassthrough, setDumpsBuffer } from '$stores/dumps';
  import { m } from '../../paraglide/messages.js';

  // The Debug window's own settings: where dumps go and how many events
  // lerd-ui keeps for the lenses.

  // Buffer sizes offered, from the floor to the ceiling lerd-ui allows, each
  // with what it costs at about 4.5 KB an event.
  const BUFFER_SIZES = [3000, 5000, 10000, 15000, 20000];
  // A size set from the CLI or MCP is listed too, so the menu shows it.
  const bufferOptions = $derived(
    [...new Set([...BUFFER_SIZES, $dumpsStatusValue?.capacity ?? 0])]
      .filter((n) => n > 0)
      .sort((a, b) => a - b)
      .map((n) => ({ value: String(n), label: m.dumps_bridge_bufferOption({ count: n.toLocaleString(), mb: Math.floor((n * 4.5) / 1000) }) }))
  );
  let bufferError = $state('');
  async function resizeBuffer(v: string) {
    bufferError = '';
    try {
      await setDumpsBuffer(Number(v));
    } catch (e) {
      bufferError = e instanceof Error ? e.message : String(e);
    }
  }

  let switchingPassthrough = $state(false);
  async function flipPassthrough() {
    if (switchingPassthrough) return;
    switchingPassthrough = true;
    try {
      await togglePassthrough(!$dumpsStatusValue?.passthrough);
      await refreshStatus();
    } finally {
      switchingPassthrough = false;
    }
  }
</script>

<div class="px-3 sm:px-5 py-4 space-y-5 text-xs text-gray-600 dark:text-gray-300 max-w-2xl">
  <div class="space-y-1.5">
    <label class="inline-flex items-center gap-2 cursor-pointer select-none font-medium text-gray-800 dark:text-gray-100">
      <input
        type="checkbox"
        class="rounded-sm border-gray-300 dark:border-lerd-border bg-white dark:bg-lerd-card text-lerd-red focus:ring-lerd-red"
        checked={Boolean($dumpsStatusValue?.passthrough)}
        disabled={switchingPassthrough}
        onchange={flipPassthrough}
      />
      {m.dumps_bridge_passthrough()}
    </label>
    <p class="text-[11px] {switchingPassthrough ? 'text-amber-600 dark:text-amber-400' : 'text-gray-500 dark:text-gray-400'}">
      {switchingPassthrough ? m.dumps_bridge_passthroughRestarting() : m.dumps_bridge_passthroughHint()}
    </p>
  </div>

  {#if $dumpsStatusValue?.capacity}
    <div class="space-y-1.5">
      <div class="flex items-center gap-2 flex-wrap font-medium text-gray-800 dark:text-gray-100">
        <span>{m.dumps_bridge_buffer()}</span>
        <Dropdown value={String($dumpsStatusValue.capacity)} options={bufferOptions} onchange={resizeBuffer} minMenuWidth={240} />
      </div>
      <p class="text-[11px] {bufferError ? 'text-amber-600 dark:text-amber-400' : 'text-gray-500 dark:text-gray-400'}">{bufferError || m.dumps_bridge_bufferHint()}</p>
    </div>
  {/if}
</div>
