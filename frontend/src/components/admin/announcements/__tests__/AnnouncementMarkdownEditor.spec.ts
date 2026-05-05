import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AnnouncementMarkdownEditor from '../AnnouncementMarkdownEditor.vue'

const uploadImage = vi.fn()

vi.mock('@/api/admin/announcements', () => ({
  uploadImage: (...args: unknown[]) => uploadImage(...args),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const messages: Record<string, string> = {
        'admin.announcements.editor.imageAlt': '\u56fe\u7247',
        'admin.announcements.editor.unsupportedType': 'unsupported image',
        'admin.announcements.editor.uploadFailed': 'upload failed',
      }
      return messages[key] || key
    },
  }),
}))

describe('AnnouncementMarkdownEditor', () => {
  beforeEach(() => {
    uploadImage.mockReset()
  })

  it('uploads a pasted image and inserts markdown at the cursor', async () => {
    uploadImage.mockResolvedValueOnce({ url: 'https://cdn.example.com/announcements/pasted.png' })
    const wrapper = mount(AnnouncementMarkdownEditor, {
      props: { modelValue: 'before ' },
    })
    const textarea = wrapper.get('textarea')
    const element = textarea.element as HTMLTextAreaElement
    element.selectionStart = 7
    element.selectionEnd = 7

    await textarea.trigger('paste', {
      clipboardData: {
        items: [
          {
            kind: 'file',
            type: 'image/png',
            getAsFile: () => new File(['image-bytes'], 'pasted.png', { type: 'image/png' }),
          },
        ],
      },
      preventDefault: vi.fn(),
    })
    await Promise.resolve()
    await Promise.resolve()

    expect(uploadImage).toHaveBeenCalled()
    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toBe('before ![\u56fe\u7247](https://cdn.example.com/announcements/pasted.png)')
  })

  it('keeps the original content and shows an error when upload fails', async () => {
    uploadImage.mockRejectedValueOnce(new Error('OSS unavailable'))
    const wrapper = mount(AnnouncementMarkdownEditor, {
      props: { modelValue: 'unchanged' },
    })

    await wrapper.get('textarea').trigger('paste', {
      clipboardData: {
        items: [
          {
            kind: 'file',
            type: 'image/png',
            getAsFile: () => new File(['image-bytes'], 'pasted.png', { type: 'image/png' }),
          },
        ],
      },
      preventDefault: vi.fn(),
    })
    await Promise.resolve()
    await Promise.resolve()

    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    expect(wrapper.text()).toContain('OSS unavailable')
  })
})
