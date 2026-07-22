<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  // [{ date: 'YYYY-MM-DD', total: Number }]
  points: { type: Array, default: () => [] },
})

// Measure the container so text stays crisp instead of being scaled by viewBox.
const wrap = ref(null)
const width = ref(680)
let ro
onMounted(() => {
  if (!wrap.value) return
  ro = new ResizeObserver(([e]) => (width.value = e.contentRect.width))
  ro.observe(wrap.value)
})
onBeforeUnmount(() => ro?.disconnect())

const H = 232
const padL = 56
const padR = 18
const padT = 26
const padB = 30
const plotW = computed(() => Math.max(10, width.value - padL - padR))
const plotH = H - padT - padB

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')
// Compact axis ticks: Indian scale (k / L / Cr).
function short(n) {
  if (n >= 1e7) return '৳' + +(n / 1e7).toFixed(1) + 'Cr'
  if (n >= 1e5) return '৳' + +(n / 1e5).toFixed(1) + 'L'
  if (n >= 1e3) return '৳' + +(n / 1e3).toFixed(0) + 'k'
  return '৳' + n
}

const hasData = computed(() => props.points.some((p) => Number(p.total) > 0))

// Round the axis top to a clean number so ticks read 0 / 5k / 10k.
function niceStep(range) {
  const exp = Math.floor(Math.log10(range || 1))
  const f = range / Math.pow(10, exp)
  const nf = f < 1.5 ? 1 : f < 3 ? 2 : f < 7 ? 5 : 10
  return nf * Math.pow(10, exp)
}
const TICKS = 4
const scale = computed(() => {
  const max = Math.max(...props.points.map((p) => Number(p.total) || 0), 0)
  if (max <= 0) return { top: 1, step: 1 }
  const step = niceStep(max / TICKS)
  return { top: Math.ceil(max / step) * step, step }
})
const ticks = computed(() => {
  const out = []
  for (let v = 0; v <= scale.value.top + 1e-6; v += scale.value.step) out.push(v)
  return out
})

const n = computed(() => props.points.length)
const x = (i) => (n.value <= 1 ? padL + plotW.value / 2 : padL + (plotW.value * i) / (n.value - 1))
const y = (v) => padT + plotH - (Number(v) || 0) / scale.value.top * plotH

const coords = computed(() => props.points.map((p, i) => ({ ...p, cx: x(i), cy: y(p.total), i })))
const linePath = computed(() => coords.value.map((p, i) => `${i ? 'L' : 'M'}${p.cx},${p.cy}`).join(' '))
const areaPath = computed(() => {
  if (!coords.value.length) return ''
  const base = padT + plotH
  return `${linePath.value} L${coords.value.at(-1).cx},${base} L${coords.value[0].cx},${base} Z`
})

// Label only the extreme — a number on every point is noise.
const peak = computed(() => {
  if (!hasData.value) return null
  return coords.value.reduce((a, b) => (Number(b.total) > Number(a.total) ? b : a))
})

const hover = ref(null)
const active = computed(() => (hover.value == null ? null : coords.value[hover.value]))

const dayLabel = (d) =>
  new Date(d + 'T00:00:00').toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
const fullLabel = (d) =>
  new Date(d + 'T00:00:00').toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })

// Keep the tooltip inside the card.
const tipLeft = computed(() => {
  if (!active.value) return 0
  return Math.min(Math.max(active.value.cx, 68), width.value - 68)
})
</script>

<template>
  <div ref="wrap" class="viz relative">
    <svg :width="width" :height="H" class="block select-none" role="img" aria-label="Sales for the last 7 days">
      <!-- Gridlines + y ticks: solid hairlines, one step off the surface.
           With no data there are no meaningful values, so only the baseline shows. -->
      <g class="text-slate-200 dark:text-slate-700">
        <line
          v-for="t in (hasData ? ticks : [0])"
          :key="'g' + t"
          :x1="padL" :x2="padL + plotW" :y1="y(t)" :y2="y(t)"
          stroke="currentColor" stroke-width="1" shape-rendering="crispEdges"
        />
      </g>
      <g v-if="hasData" class="text-slate-400 dark:text-slate-500" style="font-variant-numeric: tabular-nums">
        <text
          v-for="t in ticks"
          :key="'t' + t"
          :x="padL - 10" :y="y(t) + 4"
          text-anchor="end" font-size="11" fill="currentColor"
        >{{ short(t) }}</text>
      </g>

      <template v-if="hasData">
        <!-- Area wash + 2px line, both in the single series hue -->
        <path :d="areaPath" fill="currentColor" fill-opacity="0.10" class="text-brand-600 dark:text-brand-500" />
        <path
          :d="linePath" fill="none" stroke="currentColor" stroke-width="2"
          stroke-linejoin="round" stroke-linecap="round"
          class="text-brand-600 dark:text-brand-500"
        />

        <!-- Crosshair for the hovered column -->
        <line
          v-if="active"
          :x1="active.cx" :x2="active.cx" :y1="padT" :y2="padT + plotH"
          stroke="currentColor" stroke-width="1" class="text-slate-300 dark:text-slate-600"
        />

        <!-- Markers: 2px surface ring keeps them legible over the line -->
        <g v-for="p in coords" :key="'m' + p.date">
          <circle
            v-if="Number(p.total) > 0 || hover === p.i"
            :cx="p.cx" :cy="p.cy"
            :r="hover === p.i ? 6 : 4.5"
            fill="currentColor" class="text-brand-600 dark:text-brand-500"
            stroke="var(--viz-surface)" stroke-width="2"
          />
        </g>

        <!-- Direct label on the extreme only -->
        <text
          v-if="peak && hover === null"
          :x="Math.min(Math.max(peak.cx, padL + 22), padL + plotW - 22)"
          :y="peak.cy - 12"
          text-anchor="middle" font-size="11" font-weight="600"
          fill="currentColor" class="text-slate-600 dark:text-slate-300"
        >{{ short(peak.total) }}</text>
      </template>

      <!-- X labels -->
      <g class="text-slate-400 dark:text-slate-500">
        <text
          v-for="p in coords" :key="'x' + p.date"
          :x="p.cx" :y="H - 10" text-anchor="middle" font-size="11" fill="currentColor"
        >{{ dayLabel(p.date) }}</text>
      </g>

      <!-- Hit areas: full-column targets, far bigger than the marks -->
      <rect
        v-for="p in (hasData ? coords : [])" :key="'h' + p.date"
        :x="p.cx - (plotW / Math.max(1, n - 1)) / 2" :y="padT"
        :width="plotW / Math.max(1, n - 1)" :height="plotH"
        fill="transparent"
        @mouseenter="hover = p.i" @mouseleave="hover = null"
      />
    </svg>

    <!-- Empty state -->
    <div v-if="!hasData" class="pointer-events-none absolute inset-0 grid place-items-center">
      <span class="rounded-lg bg-white/80 px-3 py-1.5 text-sm text-slate-400 dark:bg-slate-800/80">
        No sales in the last 7 days
      </span>
    </div>

    <!-- Tooltip (enhances; values are also on the axis + peak label) -->
    <div
      v-if="active"
      class="pointer-events-none absolute z-10 -translate-x-1/2 -translate-y-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs shadow-lg dark:border-slate-600 dark:bg-slate-700"
      :style="{ left: tipLeft + 'px', top: active.cy - 14 + 'px' }"
    >
      <div class="text-slate-400">{{ fullLabel(active.date) }}</div>
      <div class="mt-0.5 font-semibold text-slate-800 dark:text-slate-100">{{ money(active.total) }}</div>
    </div>
  </div>
</template>

<style scoped>
/* Marker rings are drawn in the card surface color so they read in both themes. */
.viz {
  --viz-surface: #ffffff;
}
:global(.dark) .viz {
  --viz-surface: #1e293b; /* slate-800 = card bg in dark */
}
</style>
