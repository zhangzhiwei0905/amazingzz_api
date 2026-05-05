import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar backend mode regular user navigation', () => {
  it('renders the regular user menu even when backend mode is enabled', () => {
    expect(componentSource).toContain('<!-- Regular User View -->')
    expect(componentSource).toContain('<template v-else>')
    expect(componentSource).not.toContain('<template v-else-if="!appStore.backendModeEnabled">')
  })

  it('keeps package-sales self-service menus visible without the legacy payment flag', () => {
    expect(componentSource).toContain("path: '/subscriptions'")
    expect(componentSource).toContain("path: '/purchase', label: t('nav.buySubscription'), icon: RechargeSubscriptionIcon, hideInSimpleMode: true }")
    expect(componentSource).toContain("path: '/orders', label: t('nav.myOrders'), icon: OrderListIcon, hideInSimpleMode: true }")
    expect(componentSource).not.toContain("path: '/purchase', label: t('nav.buySubscription'), icon: RechargeSubscriptionIcon, hideInSimpleMode: true, featureFlag: flagPayment")
    expect(componentSource).not.toContain("path: '/orders', label: t('nav.myOrders'), icon: OrderListIcon, hideInSimpleMode: true, featureFlag: flagPayment")
  })
})

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar subscription plan shelf navigation', () => {
  it('exposes subscription plan shelf as a dedicated admin menu item', () => {
    expect(componentSource).toContain("path: '/admin/subscription-plans'")
    expect(componentSource).toContain("label: t('nav.subscriptionPlanShelf')")
    expect(componentSource).not.toContain("path: '/admin/subscription-plans', label: t('nav.subscriptionPlanShelf'), icon: CreditCardIcon, hideInSimpleMode: true, featureFlag: flagAdminPayment")
    expect(componentSource).not.toContain("path: '/admin/orders/plans', label: t('nav.paymentPlans')")
  })
})
