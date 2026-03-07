<template>
    <n-spin
        :show="isLoading"
        stroke="#1876F2"
    >
        <div class="app-container">
            <main class="preview-pane">
                <MarkdownPreviewer
                    v-if="markdownContent"
                    :markdown="markdownContent"
                />
            </main>
        </div>
    </n-spin>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import MarkdownPreviewer from './components/MarkdownPreviewer.vue';
import { NSpin } from 'naive-ui';
import { storeToRefs } from 'pinia';
import { useUserStore } from '@/store/user';

const markdownContent = ref('');

const isLoading = ref(false);

const fetchMarkdownFile = async (filename: string) => {
    try {
        isLoading.value = true;
        const response = await fetch(filename);
        if (!response.ok) {
            throw new Error(`error: ${response.status}`);
        }
        return await response.text();
    } catch (e) {
        throw e;
    } finally {
        isLoading.value = false;
    }
};

const userStore = useUserStore();

const { isPrivate, userName } = storeToRefs(userStore);

watch(
    userName,
    async (val) => {
        if (val) {
            const fileUrl = `md/${isPrivate.value ? 'private' : 'public'}.md`;

            const content = await fetchMarkdownFile(fileUrl);

            markdownContent.value = content;
        }
    },
    {
        immediate: true,
    },
);
</script>

<style scoped lang="less">
body {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
    background-color: #f6f8fa;
    color: #24292e;
}

.app-container {
    padding: 20px;
    min-height: calc(100vh - 68px);
}

header {
    text-align: center;
    margin-bottom: 2rem;
    border-bottom: 1px solid #e1e4e8;
    padding-bottom: 1rem;
}

.editor-layout {
    display: flex;
    gap: 2rem;
    height: calc(100vh - 150px);
}

.editor-pane,
.preview-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
}

h2 {
    margin-top: 0;
    font-size: 1.2rem;
    color: #586069;
}

.editor-input {
    width: 100%;
    height: 100%;
    border: 1px solid #d1d5da;
    border-radius: 6px;
    padding: 16px;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
    font-size: 14px;
    line-height: 1.5;
    resize: none;
    box-sizing: border-box;
}

.editor-input:focus {
    outline: none;
    border-color: #0366d6;
    box-shadow: 0 0 0 3px rgba(3, 102, 214, 0.3);
}

.preview-pane .markdown-body {
    overflow-y: auto;
    height: 100%;
}
</style>
