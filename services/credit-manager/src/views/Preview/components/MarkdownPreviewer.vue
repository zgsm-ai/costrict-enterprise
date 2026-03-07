<template>
    <div
        class="markdown-body"
        v-html="sanitizedHtml"
    />
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { gfmHeadingId } from 'marked-gfm-heading-id';

const props = defineProps({
    markdown: {
        type: String,
        required: true,
    },
});

marked.setOptions({
    gfm: true,
    pedantic: false,
});

// 锚点处理id
marked.use(gfmHeadingId());

const sanitizedHtml = computed(() => {
    const rawHtml = marked(props.markdown);
    return DOMPurify.sanitize(rawHtml as string);
});
</script>

<style scoped lang="less">
.markdown-body {
    box-sizing: border-box;
    min-width: 200px;
    max-width: 980px;
    margin: 0 auto;
    padding: 45px;
    background-color: #ffffff;
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    color: #000;

    :deep(ul),
    :deep(ol) {
        list-style-type: revert;
        list-style-position: revert;
        padding-left: 2em;
    }

    :deep(ul) {
        list-style-type: disc;

        ul {
            list-style-type: circle;

            ul {
                list-style-type: square;
            }
        }
    }

    :deep(ol) {
        list-style-type: decimal;

        ul {
            list-style-type: circle;

            ul {
                list-style-type: square;
            }
        }
    }
}

@media (prefers-color-scheme: dark) {
    .markdown-body {
        background-color: #0d1117;
        border-color: #30363d;
        color: #c9d1d9;
    }
}
</style>
