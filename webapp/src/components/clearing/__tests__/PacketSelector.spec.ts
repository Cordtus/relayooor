import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PacketSelector from '@/components/clearing/PacketSelector.vue'
import type { StuckPacket } from '@/services/packets'

describe('PacketSelector', () => {
  const mockPackets: StuckPacket[] = [
    {
      id: '1',
      channelId: 'channel-0',
      sequence: 123,
      sourceChain: 'osmosis-1',
      destinationChain: 'cosmoshub-4',
      stuckDuration: '30m',
      amount: '1000000',
      denom: 'uosmo',
      sender: 'osmo1sender',
      receiver: 'cosmos1receiver',
      timestamp: new Date().toISOString(),
      chain: 'osmosis-1',
      channel: 'channel-0',
      age: '30m',
      attempts: 2,
    },
    {
      id: '2',
      channelId: 'channel-0',
      sequence: 124,
      sourceChain: 'osmosis-1',
      destinationChain: 'cosmoshub-4',
      stuckDuration: '1h',
      amount: '500000',
      denom: 'uosmo',
      sender: 'osmo1sender',
      receiver: 'cosmos1receiver',
      timestamp: new Date().toISOString(),
      chain: 'osmosis-1',
      channel: 'channel-0',
      age: '1h',
      attempts: 5,
    },
    {
      id: '3',
      channelId: 'channel-141',
      sequence: 456,
      sourceChain: 'cosmoshub-4',
      destinationChain: 'osmosis-1',
      stuckDuration: '2h',
      amount: '2000000',
      denom: 'uatom',
      sender: 'cosmos1sender',
      receiver: 'osmo1receiver',
      timestamp: new Date().toISOString(),
      chain: 'cosmoshub-4',
      channel: 'channel-141',
      age: '2h',
      attempts: 10,
    },
  ]

  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders packet list correctly', () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
      },
    })

    // Check packet count
    expect(wrapper.findAll('[data-testid="packet-item"]')).toHaveLength(3)
    
    // Check first packet details
    const firstPacket = wrapper.find('[data-testid="packet-item-1"]')
    expect(firstPacket.text()).toContain('channel-0')
    expect(firstPacket.text()).toContain('#123')
    expect(firstPacket.text()).toContain('1 OSMO')
    expect(firstPacket.text()).toContain('30m')
  })

  it('displays empty state when no packets', () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: [],
        selected: [],
      },
    })

    expect(wrapper.find('[data-testid="empty-state"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('No stuck packets found')
  })

  it('handles packet selection', async () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
        'onUpdate:selected': (e: StuckPacket[]) => wrapper.setProps({ selected: e }),
      },
    })

    // Click first packet checkbox
    await wrapper.find('[data-testid="packet-checkbox-1"]').trigger('change')

    expect(wrapper.emitted('update:selected')).toBeTruthy()
    expect(wrapper.emitted('update:selected')?.[0]).toEqual([[mockPackets[0]]])
  })

  it('handles select all functionality', async () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
        'onUpdate:selected': (e: StuckPacket[]) => wrapper.setProps({ selected: e }),
      },
    })

    // Click select all
    await wrapper.find('[data-testid="select-all-checkbox"]').trigger('change')

    expect(wrapper.emitted('update:selected')?.[0]).toEqual([mockPackets])
  })

  it('handles deselect all when all selected', async () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: mockPackets,
        'onUpdate:selected': (e: StuckPacket[]) => wrapper.setProps({ selected: e }),
      },
    })

    // Select all checkbox should be checked
    const selectAll = wrapper.find('[data-testid="select-all-checkbox"]')
    expect((selectAll.element as HTMLInputElement).checked).toBe(true)

    // Click to deselect all
    await selectAll.trigger('change')

    expect(wrapper.emitted('update:selected')?.[0]).toEqual([[]])
  })

  it('shows indeterminate state for partial selection', async () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [mockPackets[0]], // Only first packet selected
      },
    })

    const selectAll = wrapper.find('[data-testid="select-all-checkbox"]').element as HTMLInputElement
    expect(selectAll.indeterminate).toBe(true)
  })

  it('filters packets by chain', async () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
      },
    })

    // Select osmosis filter
    await wrapper.find('[data-testid="chain-filter"]').setValue('osmosis-1')

    // Should only show osmosis packets
    expect(wrapper.findAll('[data-testid^="packet-item-"]')).toHaveLength(2)
    expect(wrapper.find('[data-testid="packet-item-3"]').exists()).toBe(false)
  })

  it('sorts packets by age', async () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
      },
    })

    // Sort by age descending (oldest first)
    await wrapper.find('[data-testid="sort-select"]').setValue('age-desc')

    const items = wrapper.findAll('[data-testid^="packet-item-"]')
    expect(items[0].text()).toContain('2h') // Oldest
    expect(items[2].text()).toContain('30m') // Newest
  })

  it('sorts packets by amount', async () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
      },
    })

    // Sort by amount descending
    await wrapper.find('[data-testid="sort-select"]').setValue('amount-desc')

    const items = wrapper.findAll('[data-testid^="packet-item-"]')
    expect(items[0].text()).toContain('2 ATOM') // Highest value
    expect(items[2].text()).toContain('0.5 OSMO') // Lowest value
  })

  it('calculates total value correctly', () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: mockPackets,
      },
    })

    const summary = wrapper.find('[data-testid="selection-summary"]')
    expect(summary.text()).toContain('3 packets selected')
    expect(summary.text()).toContain('Total: 1.5 OSMO, 2 ATOM')
  })

  it('shows warning for high attempt count', () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
      },
    })

    // Third packet has 10 attempts
    const highAttemptPacket = wrapper.find('[data-testid="packet-item-3"]')
    expect(highAttemptPacket.find('[data-testid="high-attempts-warning"]').exists()).toBe(true)
    expect(highAttemptPacket.text()).toContain('10 attempts')
  })

  it('formats time correctly', () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: [
          { ...mockPackets[0], age: '1m', stuckDuration: '1m' }, // 1 minute
          { ...mockPackets[1], age: '1h', stuckDuration: '1h' }, // 1 hour
          { ...mockPackets[2], age: '1d', stuckDuration: '1d' }, // 1 day
        ],
        selected: [],
      },
    })

    const items = wrapper.findAll('[data-testid^="packet-item-"]')
    expect(items[0].text()).toContain('1m')
    expect(items[1].text()).toContain('1h')
    expect(items[2].text()).toContain('1d')
  })

  it('emits clear event for individual packet', async () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
        showClearButtons: true,
      },
    })

    await wrapper.find('[data-testid="clear-packet-1"]').trigger('click')

    expect(wrapper.emitted('clear-packet')).toBeTruthy()
    expect(wrapper.emitted('clear-packet')?.[0]).toEqual([mockPackets[0]])
  })

  it('disables selection for packets being cleared', () => {
    const wrapper = mount(PacketSelector, {
      props: {
        stuckPackets: mockPackets,
        selected: [],
        clearing: ['1'], // First packet is being cleared
      },
    })

    const firstCheckbox = wrapper.find('[data-testid="packet-checkbox-1"]').element as HTMLInputElement
    expect(firstCheckbox.disabled).toBe(true)
    
    // Shows clearing status
    expect(wrapper.find('[data-testid="packet-item-1"]').text()).toContain('Clearing...')
  })
})