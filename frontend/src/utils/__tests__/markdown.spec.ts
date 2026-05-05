import { describe, expect, it } from 'vitest'
import { renderSafeMarkdown } from '@/utils/markdown'

describe('renderSafeMarkdown', () => {
  it('keeps safe markdown images', () => {
    const html = renderSafeMarkdown('![logo](https://cdn.example.com/announcements/logo.png)')

    expect(html).toContain('<img')
    expect(html).toContain('src="https://cdn.example.com/announcements/logo.png"')
    expect(html).toContain('alt="logo"')
  })

  it('removes dangerous image attributes', () => {
    const html = renderSafeMarkdown('<img src="https://cdn.example.com/a.png" onerror="alert(1)">')

    expect(html).toContain('<img')
    expect(html).toContain('src="https://cdn.example.com/a.png"')
    expect(html).not.toContain('onerror')
  })
})
