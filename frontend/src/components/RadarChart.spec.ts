import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import RadarChart from './RadarChart.vue'

const axes = [
  { label: 'Communication', value: 5 },
  { label: 'Ownership', value: 3 },
  { label: 'Craft', value: 4 },
]

describe('RadarChart', () => {
  it('renders an accessible svg', () => {
    const wrapper = mount(RadarChart, { props: { axes } })
    const svg = wrapper.find('svg.radar')
    expect(svg.exists()).toBe(true)
    expect(svg.attributes('role')).toBe('img')
    expect(svg.attributes('aria-label')).toBe('Feedback radar chart')
  })

  it('draws one point and one label per axis', () => {
    const wrapper = mount(RadarChart, { props: { axes } })
    expect(wrapper.findAll('circle.radar__point')).toHaveLength(axes.length)
    expect(wrapper.findAll('text.radar__label')).toHaveLength(axes.length)
  })

  it('labels each axis with its name and value', () => {
    const wrapper = mount(RadarChart, { props: { axes } })
    const labels = wrapper.findAll('text.radar__label').map((n) => n.text())
    expect(labels).toContain('Communication (5)')
    expect(labels).toContain('Ownership (3)')
  })

  it('renders the data polygon', () => {
    const wrapper = mount(RadarChart, { props: { axes } })
    const shape = wrapper.find('polygon.radar__shape')
    expect(shape.exists()).toBe(true)
    // points attr should carry one coordinate pair per axis
    expect(shape.attributes('points')?.trim().split(/\s+/)).toHaveLength(axes.length)
  })
})
