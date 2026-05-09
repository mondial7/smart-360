<script setup lang="ts">
import { computed } from 'vue'

interface Axis {
  label: string
  value: number
}

const props = withDefaults(defineProps<{
  axes: Axis[]
  size?: number
  max?: number
}>(), {
  size: 240,
  max: 0
})

const center = computed(() => props.size / 2)
const radius = computed(() => props.size / 2 - 28)

const effectiveMax = computed(() => {
  if (props.max > 0) return props.max
  const dataMax = Math.max(1, ...props.axes.map((a) => a.value))
  return Math.ceil(dataMax * 1.1)
})

function pointFor(index: number, value: number): [number, number] {
  const count = props.axes.length
  const angle = (Math.PI * 2 * index) / count - Math.PI / 2
  const ratio = effectiveMax.value === 0 ? 0 : value / effectiveMax.value
  const r = radius.value * ratio
  return [center.value + r * Math.cos(angle), center.value + r * Math.sin(angle)]
}

function axisLine(index: number) {
  const [x, y] = pointFor(index, effectiveMax.value)
  return { x, y }
}

function labelPosition(index: number) {
  const count = props.axes.length
  const angle = (Math.PI * 2 * index) / count - Math.PI / 2
  const r = radius.value + 18
  return {
    x: center.value + r * Math.cos(angle),
    y: center.value + r * Math.sin(angle)
  }
}

const polygonPoints = computed(() =>
  props.axes
    .map((axis, i) => pointFor(i, axis.value).join(','))
    .join(' ')
)

const ringRatios = [0.25, 0.5, 0.75, 1]
</script>

<template>
  <svg
    :viewBox="`0 0 ${size} ${size}`"
    :width="size"
    :height="size"
    role="img"
    aria-label="Feedback radar chart"
    class="radar"
  >
    <g class="radar__rings">
      <polygon
        v-for="(ratio, ringIndex) in ringRatios"
        :key="ringIndex"
        :points="axes
          .map((_, i) => {
            const angle = (Math.PI * 2 * i) / axes.length - Math.PI / 2
            const r = radius * ratio
            return `${center + r * Math.cos(angle)},${center + r * Math.sin(angle)}`
          })
          .join(' ')"
        class="radar__ring"
      />
    </g>

    <g class="radar__axes">
      <line
        v-for="(_, i) in axes"
        :key="`axis-${i}`"
        :x1="center"
        :y1="center"
        :x2="axisLine(i).x"
        :y2="axisLine(i).y"
        class="radar__axis"
      />
    </g>

    <polygon :points="polygonPoints" class="radar__shape" />

    <g class="radar__points">
      <circle
        v-for="(axis, i) in axes"
        :key="`pt-${i}`"
        :cx="pointFor(i, axis.value)[0]"
        :cy="pointFor(i, axis.value)[1]"
        r="3.5"
        class="radar__point"
      />
    </g>

    <g class="radar__labels">
      <text
        v-for="(axis, i) in axes"
        :key="`lb-${i}`"
        :x="labelPosition(i).x"
        :y="labelPosition(i).y"
        text-anchor="middle"
        dominant-baseline="middle"
        class="radar__label"
      >
        {{ axis.label }} ({{ axis.value }})
      </text>
    </g>
  </svg>
</template>

<style scoped lang="scss">
.radar {
  display: block;
  margin: 0 auto;

  &__ring {
    fill: none;
    stroke: var(--border-color);
    stroke-width: 1;
    opacity: 0.5;
  }

  &__axis {
    stroke: var(--border-color);
    stroke-width: 1;
    opacity: 0.6;
  }

  &__shape {
    fill: var(--color-primary);
    fill-opacity: 0.18;
    stroke: var(--color-primary);
    stroke-width: 2;
  }

  &__point {
    fill: var(--color-primary);
  }

  &__label {
    font-size: 0.72rem;
    fill: var(--text-secondary);
    font-weight: 500;
  }
}
</style>
