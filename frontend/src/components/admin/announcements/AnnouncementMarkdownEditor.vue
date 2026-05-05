<template>
  <div class="rounded-lg border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
    <div class="flex flex-wrap items-center justify-between gap-2 border-b border-gray-100 px-3 py-2 dark:border-dark-700">
      <div class="flex items-center gap-2">
        <button type="button" class="btn btn-secondary btn-sm" :disabled="uploading" @click="fileInput?.click()">
          <Icon name="upload" size="sm" class="mr-1.5" />
          {{ uploading ? t('admin.announcements.editor.uploading') : t('admin.announcements.editor.uploadImage') }}
        </button>
        <input ref="fileInput" type="file" accept="image/png,image/jpeg,image/webp,image/gif" class="hidden" @change="handleFileSelect" />
        <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.announcements.editor.pasteHint') }}</span>
      </div>
      <button type="button" class="btn btn-secondary btn-sm" @click="previewOpen = !previewOpen">
        <Icon name="eye" size="sm" class="mr-1.5" />
        {{ previewOpen ? t('admin.announcements.editor.hidePreview') : t('admin.announcements.editor.preview') }}
      </button>
    </div>

    <textarea
      ref="textareaRef"
      :value="modelValue"
      rows="8"
      class="min-h-[12rem] w-full resize-y border-0 bg-transparent px-3 py-3 text-sm leading-6 text-gray-900 outline-none placeholder:text-gray-400 focus:ring-0 dark:text-white"
      :placeholder="t('admin.announcements.editor.placeholder')"
      required
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
      @paste="handlePaste"
    ></textarea>

    <p v-if="error" class="border-t border-red-100 bg-red-50 px-3 py-2 text-xs text-red-600 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300">
      {{ error }}
    </p>

    <div v-if="previewOpen" class="border-t border-gray-100 bg-gray-50 px-4 py-4 dark:border-dark-700 dark:bg-dark-900/30">
      <div class="markdown-body prose prose-sm max-w-none dark:prose-invert" v-html="previewHtml"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { uploadImage } from '@/api/admin/announcements'
import { renderSafeMarkdown } from '@/utils/markdown'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const error = ref('')
const previewOpen = ref(false)

const previewHtml = computed(() => renderSafeMarkdown(props.modelValue))

async function handlePaste(event: ClipboardEvent) {
  const item = Array.from(event.clipboardData?.items || []).find((candidate) =>
    candidate.kind === 'file' && candidate.type.startsWith('image/')
  )
  const file = item?.getAsFile()
  if (!file) return

  event.preventDefault()
  await uploadAndInsert(file)
}

async function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  await uploadAndInsert(file)
}

async function uploadAndInsert(file: File) {
  if (!isSupportedImage(file)) {
    error.value = t('admin.announcements.editor.unsupportedType')
    return
  }
  uploading.value = true
  error.value = ''
  try {
    const result = await uploadImage(file)
    insertAtCursor(`![${t('admin.announcements.editor.imageAlt')}](${result.url})`)
  } catch (err: any) {
    error.value = err?.message || t('admin.announcements.editor.uploadFailed')
  } finally {
    uploading.value = false
  }
}

function isSupportedImage(file: File): boolean {
  return ['image/png', 'image/jpeg', 'image/jpg', 'image/webp', 'image/gif'].includes(file.type)
}

function insertAtCursor(markdown: string) {
  const textarea = textareaRef.value
  const value = props.modelValue || ''
  const start = textarea?.selectionStart ?? value.length
  const end = textarea?.selectionEnd ?? start
  const prefix = value.slice(0, start)
  const suffix = value.slice(end)
  const needsLeadingSpace = prefix.length > 0 && !/\s$/.test(prefix)
  const needsTrailingSpace = suffix.length > 0 && !/^\s/.test(suffix)
  const insertion = `${needsLeadingSpace ? ' ' : ''}${markdown}${needsTrailingSpace ? ' ' : ''}`
  emit('update:modelValue', `${prefix}${insertion}${suffix}`)

  nextTick(() => {
    const pos = start + insertion.length
    textareaRef.value?.setSelectionRange(pos, pos)
    textareaRef.value?.focus()
  })
}
</script>
