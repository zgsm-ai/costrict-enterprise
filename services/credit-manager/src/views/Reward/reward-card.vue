<template>
    <div class="reward-card flex flex-col">
        <slot name="header">
            <div class="reward-card__title text-[20px] self-start">{{ title }}</div>
        </slot>

        <div
            class="reward-card__content mt-4"
            :class="contentClass"
        >
            <template v-if="content && content.length > 0">
                <div
                    v-for="item in content"
                    :key="item.key"
                    class="mb-8 last:mb-0"
                >
                    <component
                        v-if="item.labelRender"
                        :is="item.labelRender"
                        v-bind="item"
                    />
                    <div
                        v-else
                        class="reward-card__content-label"
                    >
                        {{ item.label }}
                    </div>

                    <component
                        v-if="item.descriptionRender"
                        :is="item.descriptionRender"
                        v-bind="item"
                    />
                    <div
                        v-else-if="item.description"
                        class="reward-card__content-description mt-2.5"
                    >
                        {{ item.description }}
                    </div>

                    <component
                        v-if="item.textRender"
                        :is="item.textRender"
                        v-bind="item"
                    />
                    <template v-else>
                        <div
                            v-for="(textItem, textIndex) in item.text"
                            :key="textIndex"
                            class="reward-card__content-text mt-2 text-white/70 text-xs leading-5.5"
                        >
                            {{ textItem }}
                        </div>
                    </template>
                </div>
            </template>

            <template v-else>
                <slot name="label">
                    <div class="reward-card__content-label">{{ label }}</div>
                </slot>

                <slot
                    name="description"
                    v-if="description"
                    class="mt-2.5"
                    >{{ description }}</slot
                >

                <slot name="content">
                    <div class="reward-card__content-text mt-2 text-white/70 text-xs">
                        {{ text }}
                    </div>
                </slot>
            </template>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file reward-card.vue
 * Credit奖励计划页面卡片组件
 */
import type { RewardCardProps } from './interface';

withDefaults(defineProps<RewardCardProps>(), {
    title: '',
    label: '',
    text: '',
    description: '',
    contentClass: '',
    content: () => [],
});
</script>

<style lang="less" scoped>
.reward-card {
    &__title {
        background: linear-gradient(94deg, #005eff -6%, #00ffb7 91%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }
}
</style>
