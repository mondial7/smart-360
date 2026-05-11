<script setup lang="ts">
import { computed } from 'vue'

interface Slice {
  label: string
  value: number
  color: string
}

const props = withDefaults(defineProps<{
  slices: Slice[]
  size?: number
  thickness?: number
}>(), {
  size: 180,
  thickness: 24
})

const total = computed(() => props.slices.reduce((sum, s) => sum + s.value, 0))
const radius = computed(() => (props.size - props.thickness) / 2)
const circumference = computed(() => 2 * Math.PI * radius.value)
const center = computed(() => props.size / 2)

interface Segment {
  label: string
  value: number
  color: string
  dasharray: string
  dashoffset: number
}

const segments = computed<Segment[]>(() => {
  if (total.value === 0) return []
  let consumed = 0
  return props.slices
    .filter((s) => s.value > 0)
    .map((s) => {
      const portion = s.value / total.value
      const length = portion * circumference.value
      const seg: Segment = {
        label: s.label,
        value: s.value,
        color: s.color,
        dasharray: `${length} ${circumference.value - length}`,
        dashoffset: -consumed
      }
      consumed += length
      return seg
    })
})
</script>

<template>
  <div class="donut">
    <svg :viewBox="`0 0 ${size} ${size}`" :width="size" :height="size" class="donut__svg" role="img" aria-label="Status breakdown">
      <circle
        :cx="center"
        :cy="center"
        :r="radius"
        fill="none"
        stroke="var(--bg-secondary)"
        :stroke-width="thickness"
      />
      <circle
        v-for="(seg, i) in segments"
        :key="`seg-${i}`"
        :cx="center"
        :cy="center"
        :r="radius"
        fill="none"
        :stroke="seg.color"
        :stroke-width="thickness"
        :stroke-dasharray="seg.dasharray"
        :stroke-dashoffset="seg.dashoffset"
        :transform="`rotate(-90 ${center} ${center})`"
        class="donut__segment"
      />
      <text :x="center" :y="center - 4" text-anchor="middle" dominant-baseline="middle" class="donut__total-value">
        {{ total }}
      </text>
      <text :x="center" :y="center + 14" text-anchor="middle" dominant-baseline="middle" class="donut__total-label">
        rounds
      </text>
    </svg>

    <ul class="donut__legend">
      <li v-for="(s, i) in slices" :key="`leg-${i}`" class="donut__legend-item">
        <span class="donut__chip" :style="{ background: s.color }"></span>
        <span class="donut__label">{{ s.label }}</span>
        <span class="donut__value">{{ s.value }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped lang="scss">
.donut {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  flex-wrap: wrap;

  &__svg {
    flex-shrink: 0;
  }

  &__segment {
    transition: stroke-dashoffset 0.4s ease, stroke-dasharray 0.4s ease;
  }

  &__total-value {
    font-size: 1.4rem;
    font-weight: 700;
    fill: var(--text-primary);
  }

  &__total-label {
    font-size: 0.7rem;
    fill: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  &__legend {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: max-content 1fr max-content;
    column-gap: 0.5rem;
    row-gap: 0.4rem;
    align-items: center;
  }

  &__legend-item {
    display: contents;
  }

  &__chip {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 2px;
  }

  &__label {
    color: var(--text-primary);
    font-size: 0.85rem;
    text-transform: capitalize;
  }

  &__value {
    color: var(--text-secondary);
    font-size: 0.85rem;
    font-variant-numeric: tabular-nums;
    text-align: right;
  }
}
</style>
