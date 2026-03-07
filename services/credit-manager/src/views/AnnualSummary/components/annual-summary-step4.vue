<template>
    <div class="flex items-center justify-center h-full w-full">
        <div class="relative max-w-93.75 aspect-375/667 w-full overflow-hidden">
            <!-- 背景图 -->
            <img
                class="block absolute max-w-93.75 z-1"
                src="../../../assets/summary/step4_cover.webp"
                alt="summary_cover"
            />

            <Transition name="slide-in-left">
                <img
                    v-if="showFireball"
                    class="block absolute z-2 w-13 top-18 fireball-pos"
                    src="../../../assets/summary/fire.gif"
                    alt="fireball"
                />
            </Transition>

            <Transition name="text-fade">
                <div
                    v-if="showText"
                    class="absolute z-3 font-zcool top-26.5 text-lg text-white w-full"
                >
                    <p class="ml-4.75 truncate max-w-[70%]">
                        {{ t('annualSummary.youInitiated') }}
                        <span class="mx-1 text-3xl">{{ userData?.totalRequests ?? '---' }}</span>
                        {{ t('annualSummary.challenges') }}
                    </p>
                </div>
            </Transition>

            <Transition name="slide-down">
                <img
                    v-if="showIcon1"
                    class="block absolute z-2 icon-pos left-0 right-0 mx-auto"
                    src="../../../assets/summary/step4_icon_1.webp"
                    alt="step4_icon_1"
                />
            </Transition>

            <Transition name="dissolve">
                <img
                    v-if="showStar"
                    class="block absolute z-2 mx-auto w-36.5 left-28.75 top-56.5 cursor-pointer"
                    src="../../../assets/summary/star.gif"
                    alt="star"
                    @click="handleStarClick"
                />
            </Transition>

            <Transition name="dissolve">
                <img
                    v-if="showIcon2"
                    class="block absolute z-2 mx-auto w-55.75 right-2 top-67.5"
                    src="../../../assets/summary/step4_icon_2.webp"
                    alt="step4_icon_2"
                />
            </Transition>

            <Transition name="text2-dissolve">
                <div
                    v-if="showText2"
                    class="absolute z-3 font-zcool left-9 top-109.25 text-lg text-white max-w-80 leading-normal"
                >
                    <p class="mb-2">{{ t('annualSummary.skillUnlocked') }}</p>
                    <p class="mb-2 truncate max-w-full">
                        {{ t('annualSummary.codeModeSkill') }}
                        <span
                            :class="codePercentage !== '---' ? 'text-[#00FF63]' : ''"
                            class="ml-1"
                            >{{ codePercentage }}</span
                        >
                    </p>
                    <p class="mb-2 truncate max-w-full">
                        {{ t('annualSummary.askModeSkill') }}
                        <span
                            :class="askPercentage !== '---' ? 'text-[#00FF63]' : ''"
                            class="ml-1"
                            >{{ askPercentage }}</span
                        >
                    </p>
                    <p class="mb-2 truncate max-w-full">
                        {{ t('annualSummary.architectModeSkill') }}
                        <span
                            :class="architectPercentage !== '---' ? 'text-[#00FF63]' : ''"
                            class="ml-1"
                            >{{ architectPercentage }}</span
                        >
                    </p>
                    <p class="mb-2 truncate max-w-full">
                        {{ t('annualSummary.tokenConsumption')
                        }}<span
                            :class="
                                userData?.totalTokens !== undefined &&
                                userData?.totalTokens !== null
                                    ? 'text-[#00FF63]'
                                    : ''
                            "
                            class="ml-1"
                            >{{ userData?.totalTokens ?? '---' }}</span
                        >
                    </p>
                </div>
            </Transition>

            <Transition name="next-dissolve">
                <img
                    v-if="showNextIcon"
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
 * @file 年度总结模块step4
 */
import { ref, onMounted, computed } from 'vue';
import { useSafeWait } from '../hooks/useSafeWait';
import { getT } from '@/utils/i18n';
import type { UserMeData } from '@/api/bos/activity.bo';

const { safeWait } = useSafeWait();
const t = getT();

/**
 * 模型统计数据结构
 */
interface ModelStats {
    code: number;
    ask: number;
    architect: number;
}

interface Props {
    stepId: number;
    userData?: UserMeData | null;
}

const props = defineProps<Props>();
const emit = defineEmits<{ next: []; prev: [] }>();

const showFireball = ref(false);
const showText = ref(false);
const showIcon1 = ref(false);
const showStar = ref(false);
const showIcon2 = ref(false);
const showText2 = ref(false);
const showNextIcon = ref(false);

/**
 * 解析模型统计数据
 */
const modelStatsData = computed<ModelStats | null>(() => {
    try {
        if (!props.userData?.modeStatsJson) {
            return null;
        }
        return props.userData.modeStatsJson as unknown as ModelStats;
    } catch {
        return null;
    }
});

/**
 * 计算模式的百分比
 */
const calculatePercentage = (
    modeStats: ModelStats | null,
    mode: 'code' | 'ask' | 'architect',
): string => {
    if (!modeStats) {
        return '---';
    }

    // 获取当前模式的数值，如果不存在则返回 ---
    const modeCount = modeStats[mode];
    if (modeCount === undefined || modeCount === null) {
        return '---';
    }

    // 计算总数（忽略不存在的值）
    const codeCount = modeStats.code ?? 0;
    const askCount = modeStats.ask ?? 0;
    const architectCount = modeStats.architect ?? 0;

    const total = codeCount + askCount + architectCount;
    if (total === 0) {
        return '0%';
    }

    const percentage = Math.round((modeCount / total) * 100);
    return `${percentage}%`;
};

/**
 * Code模式百分比
 */
const codePercentage = computed(() => calculatePercentage(modelStatsData.value, 'code'));

/**
 * Ask模式百分比
 */
const askPercentage = computed(() => calculatePercentage(modelStatsData.value, 'ask'));

/**
 * Architect模式百分比
 */
const architectPercentage = computed(() => calculatePercentage(modelStatsData.value, 'architect'));

onMounted(async () => {
    await safeWait(1000);
    showFireball.value = true;

    await safeWait(1400);
    showText.value = true;

    await safeWait(500);
    showIcon1.value = true;

    await safeWait(1000);
    showStar.value = true;
});

const handleStarClick = async () => {
    showIcon1.value = false;
    showStar.value = false;

    await safeWait(300);
    showIcon2.value = true;

    await safeWait(500);
    showText2.value = true;

    await safeWait(1600);
    showNextIcon.value = true;
};

const handleNext = () => {
    emit('next');
};
</script>

<style scoped lang="less">
.fireball-pos {
    left: calc(100% - 0.8125rem - 3.25rem);
}

.icon-pos {
    top: 13.25rem;
}

.slide-in-left-enter-active,
.slide-in-left-leave-active {
    transition: left 800ms ease-in-out;
}
.slide-in-left-enter-from,
.slide-in-left-leave-to {
    left: -3.25rem;
}

.text-fade-enter-active,
.text-fade-leave-active {
    transition: opacity 300ms ease-in-out;
}
.text-fade-enter-from,
.text-fade-leave-to {
    opacity: 0;
}

.slide-down-enter-active {
    transition: top 500ms ease-in-out;
}

.slide-down-leave-active {
    transition: opacity 300ms ease-in-out;
}

.slide-down-enter-from {
    top: -13.25rem;
}

.slide-down-leave-to {
    opacity: 0;
}

.dissolve-enter-active {
    transition: opacity 300ms ease-in-out;
}

.dissolve-leave-active {
    transition: opacity 300ms ease-in-out;
}

.dissolve-enter-from,
.dissolve-leave-to {
    opacity: 0;
}

.text2-dissolve-enter-active {
    transition: opacity 500ms ease-in-out;
}

.text2-dissolve-leave-active {
    transition: opacity 300ms ease-in-out;
}

.text2-dissolve-enter-from,
.text2-dissolve-leave-to {
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

.next-icon-pos {
    bottom: 2.5rem;
    right: 4.75rem;
}
</style>
