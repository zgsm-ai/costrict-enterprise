<template>
    <div class="flex items-center justify-center h-full w-full">
        <div class="relative max-w-93.75 aspect-375/667 w-full bg-black overflow-hidden">
            <!-- 视频播放器 -->
            <video
                v-if="videoSrc"
                :src="videoSrc"
                class="absolute inset-0 w-full h-full object-cover"
                autoplay
                muted
                playsinline
                @ended="handleVideoEnded"
            ></video>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 年度总结模块step7
 */
import { computed } from 'vue';
import { TYPE_VIDEO_MAP, DEFAULT_RESULT_TYPE } from '../const';

const props = withDefaults(
    defineProps<{
        stepId: number;
        type?: string;
    }>(),
    {
        type: DEFAULT_RESULT_TYPE,
    },
);

const emit = defineEmits<{
    next: [];
}>();

const videoSrc = computed(() => TYPE_VIDEO_MAP[props.type] || TYPE_VIDEO_MAP[DEFAULT_RESULT_TYPE]);

/**
 * 视频播放完毕事件处理
 */
const handleVideoEnded = () => {
    emit('next');
};
</script>

<style scoped lang="less"></style>
