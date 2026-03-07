<template>
    <div class="flex items-center justify-center h-full w-full">
        <div class="relative max-w-93.75 aspect-375/667 w-full">
            <img
                class="block absolute max-w-93.75 z-1"
                src="../../../assets/summary/step5_cover.webp"
                alt="summary_cover"
            />

            <Transition name="dissolve">
                <img
                    v-if="showLight"
                    class="block absolute z-2 mx-auto"
                    src="../../../assets/summary/light.webp"
                    alt="light"
                />
            </Transition>

            <Transition name="dissolve">
                <div
                    v-if="showText1"
                    class="absolute z-3 font-zcool right-3.75 top-17.75 text-lg text-white"
                >
                    <p>{{ t('annualSummary.yearlyCreditsConsumption') }}</p>
                </div>
            </Transition>

            <Transition name="text2-dissolve">
                <div
                    v-if="showText2"
                    class="absolute z-3 font-zcool top-116.5 text-white w-full"
                >
                    <p class="text-5xl ml-7.25 truncate max-w-full">
                        {{ userData?.creditUsage || '---' }} <span class="text-3xl">⚡</span>
                    </p>
                </div>
            </Transition>

            <Transition name="text3-dissolve">
                <div
                    v-if="showText3"
                    class="relative z-3"
                >
                    <p class="font-zcool z-3 absolute top-133.5 left-7.25 text-base">
                        {{ t('annualSummary.ifOnlyCaloriesConsumed') }}
                    </p>
                    <img
                        class="block absolute z-3 mx-auto w-77.75 left-7.25 top-126.5"
                        src="../../../assets/summary/step5_icon_1.webp"
                        alt="step5_icon_1"
                    />
                </div>
            </Transition>

            <Transition name="next-dissolve">
                <img
                    v-if="showNextIcon"
                    class="z-4 w-20.5 absolute right-3.5 top-159 cursor-pointer"
                    src="../../../assets/summary/step2_icon_3.webp"
                    alt="summary_cover_icon_3"
                    @click="handleNext"
                />
            </Transition>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 年度总结模块step5
 */
import { ref, onMounted } from 'vue';
import { useSafeWait } from '../hooks/useSafeWait';
import { getT } from '@/utils/i18n';
import type { UserMeData } from '@/api/bos/activity.bo';

const { safeWait } = useSafeWait();
const t = getT();

defineProps<{
    stepId: number;
    userData: UserMeData | null;
}>();

const emit = defineEmits<{
    next: [];
}>();

const showLight = ref(false);
const showText1 = ref(false);
const showText2 = ref(false);
const showText3 = ref(false);
const showNextIcon = ref(false);

onMounted(async () => {
    await safeWait(800);
    showLight.value = true;

    await safeWait(300);
    showText1.value = true;

    await safeWait(500);
    showText2.value = true;

    await safeWait(1300);
    showText3.value = true;

    await safeWait(1300);
    showNextIcon.value = true;
});

const handleNext = () => {
    emit('next');
};
</script>

<style scoped lang="less">
.dissolve-enter-active {
    transition: opacity 500ms ease-in-out;
}

.dissolve-leave-active {
    transition: opacity 500ms ease-in-out;
}

.dissolve-enter-from,
.dissolve-leave-to {
    opacity: 0;
}

.step4_title {
    background: linear-gradient(180deg, #ffffff 42%, rgba(255, 248, 188, 0) 100%);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
}

.text2-dissolve-enter-active {
    transition: opacity 300ms ease-in-out;
}

.text2-dissolve-leave-active {
    transition: opacity 300ms ease-in-out;
}

.text2-dissolve-enter-from,
.text2-dissolve-leave-to {
    opacity: 0;
}

.text3-dissolve-enter-active {
    transition: opacity 500ms ease-in-out;
}

.text3-dissolve-leave-active {
    transition: opacity 500ms ease-in-out;
}

.text3-dissolve-enter-from,
.text3-dissolve-leave-to {
    opacity: 0;
}

.next-dissolve-enter-active {
    transition: opacity 300ms ease-in-out;
}

.next-dissolve-leave-active {
    transition: opacity 300ms ease-in-out;
}

.next-dissolve-enter-from,
.next-dissolve-leave-to {
    opacity: 0;
}
</style>
