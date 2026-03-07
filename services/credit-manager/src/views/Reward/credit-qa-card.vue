<template>
    <div class="credit-qa-content">
        <div class="question flex text-[#4394FF]">
            <div
                class="question-icon w-6 h-6 text-white bg-[#4394FF] rounded-full flex items-center justify-center shrink-0 grow-0"
            >
                {{ t('rewardPlan.qaCard.question') }}
            </div>
            <div class="ml-2">
                <slot name="question">
                    <p class="max-w-242 leading-5">{{ question }}</p>
                </slot>
            </div>
        </div>

        <div class="question flex mt-2.5">
            <div
                class="question-icon w-6 h-6 text-white bg-[#36D5A7] rounded-full flex items-center justify-center shrink-0 grow-0"
            >
                {{ t('rewardPlan.qaCard.answer') }}
            </div>
            <div class="ml-2 self-center max-w-242">
                <slot name="answer">
                    <template
                        v-for="(item, index) in processedAnswer"
                        :key="index"
                    >
                        <p
                            v-if="isString(item)"
                            class="font-medium leading-5"
                        >
                            {{ item }}
                        </p>
                        <component
                            v-else-if="isFunction(item)"
                            :is="item()"
                        />
                    </template>
                </slot>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file credit-qa-card.vue
 */

import { type VNode, computed } from 'vue';
import { useI18n } from 'vue-i18n';

type AnswerItem = string | (() => VNode);
type AnswerType = string | AnswerItem[];

interface Props {
    question?: string;
    answer?: AnswerType;
}

const { t } = useI18n();

const props = withDefaults(defineProps<Props>(), {
    question: '',
    answer: '',
});

// 判断是否为字符串
const isString = (value: unknown) => {
    return typeof value === 'string';
};

// 判断是否为函数
const isFunction = (value: unknown) => {
    return typeof value === 'function';
};

// 将answer统一转换为数组进行处理
const processedAnswer = computed(() => {
    if (isString(props.answer)) {
        return [props.answer];
    } else if (Array.isArray(props.answer)) {
        return props.answer;
    }
    return [];
});
</script>
