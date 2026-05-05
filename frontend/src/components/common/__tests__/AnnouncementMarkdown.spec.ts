import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import AnnouncementMarkdown from '@/components/common/AnnouncementMarkdown.vue'

function mountMarkdown(content: string) {
  return mount(AnnouncementMarkdown, {
    props: { content },
  })
}

describe('AnnouncementMarkdown', () => {
  it('renders ordinary non-image Markdown formatting', () => {
    const wrapper = mountMarkdown([
      '# Release notes',
      '',
      'Read the **important** [details](https://example.com/docs).',
      '',
      '- First change',
      '- Second change',
    ].join('\n'))

    expect(wrapper.get('h1').text()).toBe('Release notes')
    expect(wrapper.get('strong').text()).toBe('important')
    expect(wrapper.get('a').attributes('href')).toBe('https://example.com/docs')
    expect(wrapper.findAll('li').map((item) => item.text())).toEqual([
      'First change',
      'Second change',
    ])
    expect(wrapper.find('img').exists()).toBe(false)
  })

  it('renders HTTPS markdown images with src and alt text', () => {
    const wrapper = mountMarkdown('![Release notes](https://example.com/release.png)')
    const image = wrapper.get('img')

    expect(image.attributes('src')).toBe('https://example.com/release.png')
    expect(image.attributes('alt')).toBe('Release notes')
    expect(image.attributes('loading')).toBe('lazy')
    expect(image.attributes('decoding')).toBe('async')
    expect(image.attributes('referrerpolicy')).toBe('no-referrer')
  })

  it('renders same-origin relative markdown images', () => {
    const wrapper = mountMarkdown('![Dashboard](/images/dashboard.png)')
    const image = wrapper.get('img')

    expect(image.attributes('src')).toBe('/images/dashboard.png')
    expect(image.attributes('alt')).toBe('Dashboard')
  })

  it('removes usable image src values for protocol-relative image URLs', () => {
    const wrapper = mountMarkdown('![Protocol relative](//example.com/image.png)')
    const image = wrapper.get('img')

    expect(image.attributes('src')).toBeUndefined()
    expect(image.attributes('alt')).toBe('Protocol relative')
    expect(image.attributes('loading')).toBeUndefined()
    expect(image.attributes('decoding')).toBeUndefined()
    expect(image.attributes('referrerpolicy')).toBeUndefined()
  })

  it('removes usable image src values for non-HTTPS image protocols', () => {
    const wrapper = mountMarkdown('![Ftp](ftp://example.com/image.png)')
    const image = wrapper.get('img')

    expect(image.attributes('src')).toBeUndefined()
    expect(image.attributes('alt')).toBe('Ftp')
    expect(image.attributes('loading')).toBeUndefined()
    expect(image.attributes('decoding')).toBeUndefined()
    expect(image.attributes('referrerpolicy')).toBeUndefined()
  })

  it('removes usable image src values for disallowed protocols', () => {
    const wrapper = mountMarkdown([
      '![Http](http://example.com/insecure.png)',
      '![Javascript](javascript:alert(1))',
      '![Data](data:image/png;base64,AAAA)',
    ].join('\n'))
    const images = wrapper.findAll('img')

    expect(images).toHaveLength(3)
    images.forEach((image) => {
      expect(image.attributes('src')).toBeUndefined()
      expect(image.attributes('loading')).toBeUndefined()
      expect(image.attributes('decoding')).toBeUndefined()
      expect(image.attributes('referrerpolicy')).toBeUndefined()
    })
  })

  it('removes script content', () => {
    const wrapper = mountMarkdown('Safe text<script>alert("xss")</script>')

    expect(wrapper.html()).toContain('Safe text')
    expect(wrapper.find('script').exists()).toBe(false)
    expect(wrapper.html()).not.toContain('alert')
  })

  it('removes raw HTML picture/source/srcset while preserving safe fallback image', () => {
    const wrapper = mountMarkdown('<picture><source srcset="http://example.com/a.png"><img src="https://example.com/ok.png"></picture>')

    expect(wrapper.find('picture').exists()).toBe(false)
    expect(wrapper.find('source').exists()).toBe(false)
    expect(wrapper.html()).not.toContain('srcset')

    const image = wrapper.get('img')
    expect(image.attributes('src')).toBe('https://example.com/ok.png')
    expect(image.attributes('loading')).toBe('lazy')
    expect(image.attributes('decoding')).toBe('async')
    expect(image.attributes('referrerpolicy')).toBe('no-referrer')
  })

  it('removes raw HTML elements and attributes that can load external media outside img src', () => {
    const wrapper = mountMarkdown([
      '<div style="background-image:url(http://example.com/pixel.png)">x</div>',
      '<input type="image" src="http://example.com/button.png">',
      '<table background="http://example.com/bg.png"><tr><td>x</td></tr></table>',
    ].join(''))

    expect(wrapper.text()).toContain('x')
    expect(wrapper.find('input').exists()).toBe(true)
    expect(wrapper.find('input').attributes('src')).toBeUndefined()
    expect(wrapper.find('table').exists()).toBe(true)
    expect(wrapper.html()).not.toContain('style=')
    expect(wrapper.html()).not.toContain('background=')
    expect(wrapper.html()).not.toContain('src=')
    expect(wrapper.html()).not.toContain('http://example.com')
  })
})
