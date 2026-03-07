<template>
    <div class="flex items-center justify-center h-full w-full">
        <div class="relative max-w-93.75 aspect-375/667 w-full">
            <img
                class="block absolute max-w-93.75 z-1"
                src="../../../assets/summary/step2_cover.webp"
                alt="summary_cover"
            />
            <Transition name="fade">
                <img
                    v-if="showCoverIcon1"
                    class="block absolute z-2 w-44.5 top-71 left-28.25 max-w-93.75"
                    src="../../../assets/summary/step2_icon_1.webp"
                    alt="summary_cover_icon_1"
                />
            </Transition>
            <Transition name="fade">
                <img
                    v-if="showCoverIcon2"
                    class="block absolute max-w-93.75 z-2"
                    src="../../../assets/summary/step2_icon_2.webp"
                    alt="summary_cover_icon_2"
                />
            </Transition>
            <Transition name="fade">
                <span
                    v-if="showCoverDate"
                    class="font-zcool z-2 absolute top-24 right-9.5 text-2xl"
                    >{{ formattedDate }}</span
                >
            </Transition>
            <Transition name="fade">
                <div
                    v-if="showCoverText"
                    class="z-2 relative"
                >
                    <p
                        class="font-zcool z-2 absolute top-126.25 left-10.75 text-2xl truncate max-w-[70%]"
                    >
                        <span class="text-sm mr-3">{{ t('annualSummary.dear') }}</span
                        ><span>{{ userName }}</span>
                    </p>
                    <span class="font-zcool z-2 absolute top-136.75 left-10.75 text-2xl">{{
                        t('annualSummary.unlockItem')
                    }}</span>
                </div>
            </Transition>
            <Transition name="fade">
                <img
                    v-if="showCoverIcon3"
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
 * @file 年度总结模块step2
 */
import { ref, onMounted, computed } from 'vue';
import { useSafeWait } from '../hooks/useSafeWait';
import type { UserMeData } from '@/api/bos/activity.bo';
import { getT } from '@/utils/i18n';

const { safeWait } = useSafeWait();
const t = getT();

const props = defineProps<{
    stepId: number;
    userData: UserMeData | null;
}>();

const emit = defineEmits<{
    next: [];
}>();

/**
 * 格式化日期为 2025年XX月XX日 格式
 */
const formattedDate = computed(() => {
    if (!props.userData?.createdTime) {
        return '---';
    }
    const date = new Date(props.userData.createdTime);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}${t('annualSummary.year')}${month}${t('annualSummary.month')}${day}${t('annualSummary.day')}`;
});

const userName = computed(() => {
    if (!props.userData) {
        return '---';
    }
    return props.userData.username || props.userData.displayName || '---';
});

const showCoverIcon1 = ref(false);
const showCoverIcon2 = ref(false);
const showCoverDate = ref(false);
const showCoverText = ref(false);
const showCoverIcon3 = ref(false);

const handleNext = () => {
    emit('next');
};

onMounted(async () => {
    await safeWait(1000);
    showCoverIcon1.value = true;

    await safeWait(1200);
    showCoverIcon2.value = true;

    await safeWait(1200);
    showCoverDate.value = true;

    await safeWait(1200);
    showCoverText.value = true;

    await safeWait(1300);
    showCoverIcon3.value = true;
});
</script>

<style scoped lang="less">
.fade-enter-active {
    transition: opacity 300ms ease-in-out;
}

.fade-enter-from {
    opacity: 0;
}
</style>
