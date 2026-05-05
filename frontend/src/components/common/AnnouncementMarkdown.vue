<template>
  <div
    class="markdown-body prose prose-sm max-w-none dark:prose-invert"
    v-html="renderedContent"
  ></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps<{
  content: string
}>()

const renderedContent = computed(() => renderMarkdown(props.content))

function renderMarkdown(content: string): string {
  if (!content) return ''

  const html = marked.parse(content, {
    breaks: true,
    gfm: true,
  }) as string
  const sanitizedHtml = DOMPurify.sanitize(html, {
    FORBID_TAGS: [
      'audio',
      'embed',
      'iframe',
      'math',
      'object',
      'picture',
      'source',
      'svg',
      'video',
    ],
    FORBID_ATTR: ['background', 'poster', 'srcset', 'style'],
  })

  return sanitizeUrlBearingAttributes(sanitizedHtml)
}

function sanitizeUrlBearingAttributes(html: string): string {
  const template = document.createElement('template')
  template.innerHTML = html

  template.content.querySelectorAll('[src]').forEach((element) => {
    if (element.tagName.toLowerCase() !== 'img') {
      element.removeAttribute('src')
    }
  })

  sanitizeImages(template.content)

  return template.innerHTML
}

function sanitizeImages(root: DocumentFragment): void {
  root.querySelectorAll('img').forEach((image) => {
    const src = image.getAttribute('src')
    image.removeAttribute('srcset')

    if (src && isAllowedImageSrc(src)) {
      image.setAttribute('src', src.trim())
      image.setAttribute('loading', 'lazy')
      image.setAttribute('decoding', 'async')
      image.setAttribute('referrerpolicy', 'no-referrer')
    } else {
      image.removeAttribute('src')
    }
  })
}

function isAllowedImageSrc(src: string): boolean {
  const trimmedSrc = src.trim()

  if (!trimmedSrc || trimmedSrc.startsWith('//')) {
    return false
  }

  try {
    const url = new URL(trimmedSrc)
    return url.protocol === 'https:'
  } catch {
    return !/^[a-zA-Z][a-zA-Z\d+.-]*:/.test(trimmedSrc)
  }
}
</script>

<style>
/* Enhanced Markdown Styles */
.markdown-body {
  @apply text-[15px] leading-[1.75];
  @apply text-gray-700 dark:text-gray-300;
}

.markdown-body h1 {
  @apply mb-6 mt-8 border-b border-gray-200 pb-3 text-3xl font-bold text-gray-900 dark:border-dark-600 dark:text-white;
}

.markdown-body h2 {
  @apply mb-4 mt-7 border-b border-gray-100 pb-2 text-2xl font-bold text-gray-900 dark:border-dark-700 dark:text-white;
}

.markdown-body h3 {
  @apply mb-3 mt-6 text-xl font-semibold text-gray-900 dark:text-white;
}

.markdown-body h4 {
  @apply mb-2 mt-5 text-lg font-semibold text-gray-900 dark:text-white;
}

.markdown-body p {
  @apply mb-4 leading-relaxed;
}

.markdown-body a {
  @apply font-medium text-blue-600 underline decoration-blue-600/30 decoration-2 underline-offset-2 transition-all hover:decoration-blue-600 dark:text-blue-400 dark:decoration-blue-400/30 dark:hover:decoration-blue-400;
}

.markdown-body ul,
.markdown-body ol {
  @apply mb-4 ml-6 space-y-2;
}

.markdown-body ul {
  @apply list-disc;
}

.markdown-body ol {
  @apply list-decimal;
}

.markdown-body li {
  @apply leading-relaxed;
  @apply pl-2;
}

.markdown-body li::marker {
  @apply text-blue-600 dark:text-blue-400;
}

.markdown-body blockquote {
  @apply relative my-5 border-l-4 border-blue-500 bg-blue-50/50 py-3 pl-5 pr-4 italic text-gray-700 dark:border-blue-400 dark:bg-blue-900/10 dark:text-gray-300;
}

.markdown-body blockquote::before {
  content: '"';
  @apply absolute -left-1 top-0 text-5xl font-serif text-blue-500/20 dark:text-blue-400/20;
}

.markdown-body code {
  @apply rounded-lg bg-gray-100 px-2 py-1 text-[13px] font-mono text-pink-600 dark:bg-dark-700 dark:text-pink-400;
}

.markdown-body pre {
  @apply my-5 overflow-x-auto rounded-xl border border-gray-200 bg-gray-50 p-5 dark:border-dark-600 dark:bg-dark-900/50;
}

.markdown-body pre code {
  @apply bg-transparent p-0 text-[13px] text-gray-800 dark:text-gray-200;
}

.markdown-body hr {
  @apply my-8 border-0 border-t-2 border-gray-200 dark:border-dark-700;
}

.markdown-body table {
  @apply mb-5 w-full overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600;
}

.markdown-body th,
.markdown-body td {
  @apply border-r border-b border-gray-200 px-4 py-3 text-left dark:border-dark-600;
}

.markdown-body th:last-child,
.markdown-body td:last-child {
  @apply border-r-0;
}

.markdown-body tr:last-child td {
  @apply border-b-0;
}

.markdown-body th {
  @apply bg-gradient-to-br from-blue-50 to-indigo-50 font-semibold text-gray-900 dark:from-blue-900/20 dark:to-indigo-900/10 dark:text-white;
}

.markdown-body tbody tr {
  @apply transition-colors hover:bg-gray-50 dark:hover:bg-dark-700/30;
}

.markdown-body img {
  @apply my-5 max-w-full rounded-xl border border-gray-200 shadow-md dark:border-dark-600;
}

.markdown-body strong {
  @apply font-semibold text-gray-900 dark:text-white;
}

.markdown-body em {
  @apply italic text-gray-600 dark:text-gray-400;
}
</style>
