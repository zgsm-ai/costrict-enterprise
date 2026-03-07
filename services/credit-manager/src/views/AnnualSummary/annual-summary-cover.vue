<template>
    <div class="flex items-center justify-center h-full w-full">
        <div class="relative max-w-93.75 aspect-375/667 w-full">
            <img
                class="block absolute max-w-93.75 z-1"
                src="../../assets/summary/cover.webp"
                alt="summary_cover"
            />
            <Transition name="fade">
                <img
                    v-if="showCoverNext"
                    class="block absolute max-w-93.75 z-2"
                    src="../../assets/summary/cover_next.webp"
                    alt="summary_cover_icon"
                />
            </Transition>
            <img
                class="absolute w-26.75 right-12 top-91.25 cursor-pointer z-10"
                src="../../assets/summary/enter.webp"
                alt="enter"
                @click="handleEnter"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 年度总结封面静态页
 */
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';

const router = useRouter();
const route = useRoute();

const showCoverNext = ref(false);

const handleEnter = async () => {
    // 未登录，跳转到登录页面，携带 URL 参数
    router.push({
        path: '/login',
        query: route.query,
    });
};

onMounted(async () => {
    await new Promise((resolve) => setTimeout(resolve, 800));
    showCoverNext.value = true;
});
</script>

<style scoped lang="less">
.fade-enter-active {
    transition: opacity 1000ms ease-in-out;
}

.fade-enter-from {
    opacity: 0;
}
</style>
