<template>
    <div class="flex items-center justify-center h-full w-full">
        <div class="relative max-w-93.75 aspect-375/667 w-full">
            <Transition name="fade">
                <img
                    v-if="!showSecondCover"
                    class="block absolute max-w-93.75 z-1"
                    src="../../../assets/summary/step3_cover.webp"
                    alt="summary_cover"
                />
                <img
                    v-else
                    class="block absolute max-w-93.75 z-1"
                    src="../../../assets/summary/step3_cover_2.webp"
                    alt="summary_cover_2"
                />
            </Transition>
            <Transition name="fade">
                <img
                    v-if="showIcon1"
                    class="block absolute z-2 w-37.5 left-38.75 top-54"
                    src="../../../assets/summary/step3_icon_1.webp"
                    alt="icon_1"
                />
            </Transition>
            <Transition name="fade">
                <div
                    v-if="showIcon2"
                    class="relative z-2"
                >
                    <img
                        class="block absolute z-2 w-19 left-50 top-8.5 cursor-pointer"
                        src="../../../assets/summary/step3_icon_3.webp"
                        alt="icon_3"
                    />
                    <img
                        class="block absolute z-2 w-17.25 right-7.5 top-26.75 animate-bounce-vertical"
                        src="../../../assets/summary/step3_icon_4.webp"
                        alt="icon_4"
                    />
                </div>
            </Transition>
            <img
                v-if="!showSecondCover"
                class="block absolute z-2 w-17 right-[29.5px] top-[17.5px] cursor-pointer"
                src="../../../assets/summary/step3_icon_2.webp"
                alt="icon_2"
                @click="handleIcon2Click"
            />
            <Transition name="slide-up">
                <div
                    v-if="showText1"
                    class="absolute z-3 font-zcool right-2.5 top-47 text-right text-lg text-white leading-8.25"
                >
                    <span>{{ t('annualSummary.pastOneYear') }}</span>
                    <p>
                        {{ t('annualSummary.youUsed')
                        }}<span class="text-[28px] mx-1 usage-date">{{ totalUsageDays }}</span
                        >{{ t('annualSummary.days') }}
                    </p>
                </div>
            </Transition>
            <Transition name="second-slide-up">
                <div
                    v-if="showText2"
                    class="absolute z-3 font-zcool text-right text-xl left-7.5 top-126.25 text-white leading-8.25"
                >
                    <p class="text-left">{{ t('annualSummary.notOneTimeTool') }}</p>
                    <p class="ml-6">{{ t('annualSummary.stayInWorkflow') }}</p>
                </div>
            </Transition>
            <Transition name="second-slide-up">
                <img
                    v-if="showIcon5"
                    class="block absolute z-2 right-0 top-82.5 w-87.5"
                    src="../../../assets/summary/step3_icon_5.webp"
                    alt="icon_5"
                />
            </Transition>
            <!-- 下一步 -->
            <Transition name="slide-up">
                <img
                    v-if="showIcon6"
                    class="z-2 w-20.5 absolute right-3.5 top-159 cursor-pointer"
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
 * @file 年度总结模块step3
 */
import { ref, onMounted, computed } from 'vue';
import { useSafeWait } from '../hooks/useSafeWait';
import { getT } from '@/utils/i18n';
import type { UserMeData } from '@/api/bos/activity.bo';

const { safeWait } = useSafeWait();
const t = getT();
const isIconClicked = ref(false);

const props = defineProps<{
    stepId: number;
    userData: UserMeData | null;
}>();

const emit = defineEmits<{
    next: [];
    prev: [];
}>();

/**
 * 获取总使用天数
 */
const totalUsageDays = computed(() => {
    return props.userData?.totalUsageDays ?? '---';
});

const showIcon1 = ref(false);
const showIcon2 = ref(false);
const showSecondCover = ref(false);
const showText1 = ref(false);
const showText2 = ref(false);
const showIcon5 = ref(false);
const showIcon6 = ref(false);

const handleNext = () => {
    emit('next');
};

const handleIcon2Click = async () => {
    isIconClicked.value = true;
    showIcon1.value = false;
    showIcon2.value = false;
    showSecondCover.value = true;

    await safeWait(800);
    showText1.value = true;

    await safeWait(1000);
    showText2.value = true;
    showIcon5.value = true;

    await safeWait(1100);
    showIcon6.value = true;
};

onMounted(async () => {
    await safeWait(200);
    if (!isIconClicked.value) {
        showIcon1.value = true;
    }

    await safeWait(600);
    if (!isIconClicked.value) {
        showIcon2.value = true;
    }
});
</script>

<style scoped lang="less">
.fade-enter-active,
.fade-leave-active {
    transition: opacity 800ms ease-in-out;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}

.slide-up-enter-active,
.slide-up-leave-active {
    transition: opacity 300ms ease-in-out;
}

.slide-up-enter-from,
.slide-up-leave-to {
    opacity: 0;
}

.second-slide-up-enter-active,
.second-slide-up-leave-active {
    transition: opacity 500ms ease-in-out;
}

.second-slide-up-enter-from,
.second-slide-up-leave-to {
    opacity: 0;
}

@keyframes bounce-vertical {
    0%,
    100% {
        transform: translateY(0);
    }
    50% {
        transform: translateY(-22px);
    }
}

.animate-bounce-vertical {
    animation: bounce-vertical 400ms infinite;
}

.usage-date {
    background: linear-gradient(137deg, #00ffb7 2%, #f7fffd 22%, #ffffff 62%, #005eff 101%);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
}
</style>
