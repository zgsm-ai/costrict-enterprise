<template>
    <div class="flex items-center justify-center h-full w-full">
        <div class="relative max-w-93.75 aspect-375/667 w-full">
            <img
                class="block absolute max-w-93.75 z-1"
                src="../../../assets/summary/step6_cover.webp"
                alt="summary_cover"
            />

            <Transition name="dissolve">
                <img
                    v-if="showStar"
                    :class="['block absolute z-2 star-icon', twinkleAnimation]"
                    src="../../../assets/summary/step6_icon_1.webp"
                    alt="star"
                />
            </Transition>

            <Transition name="dissolve">
                <img
                    v-if="showButton"
                    class="z-2 absolute left-1/2 -translate-x-1/2 top-80 w-19.75 cursor-pointer"
                    src="../../../assets/summary/step6_icon_2.webp"
                    alt="check_button"
                    @click="handleNext"
                />
            </Transition>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 年度总结模块step6
 */
import { ref, onMounted } from 'vue';
import { useSafeWait } from '../hooks/useSafeWait';

const { safeWait } = useSafeWait();

defineProps<{
    stepId: number;
}>();

const emit = defineEmits<{
    next: [];
}>();

const showStar = ref(false);
const showButton = ref(false);
const twinkleAnimation = ref('');

const handleNext = () => {
    emit('next');
};

onMounted(async () => {
    await safeWait(300);
    showStar.value = true;

    twinkleAnimation.value = 'twinkle';
    await safeWait(2000);

    twinkleAnimation.value = '';
    showButton.value = true;
});
</script>

<style scoped lang="less">
.star-icon {
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
}

@keyframes twinkle {
    0%,
    100% {
        opacity: 1;
    }
    50% {
        opacity: 0;
    }
}

.twinkle {
    animation: twinkle 2s linear infinite;
}

.dissolve-enter-active,
.dissolve-leave-active {
    transition: opacity 300ms ease-in-out;
}

.dissolve-enter-from,
.dissolve-leave-to {
    opacity: 0;
}
</style>
