<template>
    <common-card :content-border="false">
        <template #header>
            <div class="flex items-center text-xl mb-5">
                <span class="text-white">{{ t('homePage.activityTitle') }}</span>
            </div>
        </template>
        <template #default>
            <!-- Card 1: 邀请有礼 -->
            <div class="activity-card-item">
                <div class="flex items-center justify-between gap-3 flex-wrap mb-4">
                    <div class="flex items-center gap-2.5 shrink-0">
                        <span
                            class="inline-flex items-center justify-center w-7 h-7 rounded-full bg-[rgba(0,102,255,0.2)] border border-[rgba(0,102,255,0.4)] text-[13px] font-bold text-[#197dff] shrink-0"
                            >1</span
                        >
                        <div
                            class="flex items-center flex-wrap gap-2.5 font-bold text-[18px] text-white/85"
                        >
                            {{ t('activityCard.inviteReward') }}
                            <span
                                class="invite-tag inline-flex items-center px-2.5 py-0.5 rounded-[40px] bg-[rgba(0,255,183,0.12)] border border-[rgba(0,255,183,0.3)] text-[13px] font-semibold text-[#00ffb7]"
                                >{{ t('activityCard.inviteCredits') }}</span
                            >
                        </div>
                    </div>
                    <div
                        class="activity-card-btn"
                        :class="{
                            'cursor-pointer': !isInviteLoading,
                            'cursor-not-allowed opacity-50': isInviteLoading,
                        }"
                        @click="toInvite"
                    >
                        {{ t('activityCard.goInvite') }}
                    </div>
                </div>
                <p class="text-sm text-white/70 leading-[1.7]">
                    {{ t('activityCard.inviteRewardDesc') }}
                </p>
                <ul class="flex flex-col gap-2 mt-3">
                    <li
                        v-for="rule in inviteRules"
                        :key="rule"
                        class="flex items-start gap-2 text-sm text-white/70 leading-[1.6] before:content-[''] before:block before:w-[5px] before:h-[5px] before:rounded-full before:bg-[#197dff] before:shrink-0 before:mt-[7px]"
                    >
                        {{ rule }}
                    </li>
                </ul>
            </div>

            <!-- Card 2: 开源贡献激励 -->
            <div class="activity-card-item mt-3">
                <div class="flex items-center justify-between gap-3 flex-wrap mb-4">
                    <div class="flex items-center gap-2.5 shrink-0">
                        <span
                            class="inline-flex items-center justify-center w-7 h-7 rounded-full bg-[rgba(0,102,255,0.2)] border border-[rgba(0,102,255,0.4)] text-[13px] font-bold text-[#197dff] shrink-0"
                            >2</span
                        >
                        <div class="font-bold text-[18px] text-white/85">
                            {{ t('activityCard.contributionTitle') }}
                        </div>
                    </div>
                    <a
                        class="activity-card-btn cursor-pointer"
                        target="_blank"
                        rel="noopener"
                        @click.prevent="toGithub"
                    >
                        {{ t('activityCard.contributionBtn') }}<span class="arrow">></span>
                    </a>
                </div>
                <p class="text-sm text-white/70 leading-[1.7]">
                    {{ t('activityCard.contributionDesc') }}
                </p>

                <NDataTable
                    class="contrib-table mt-4"
                    :columns="contribColumns"
                    :data="contribRows"
                    :row-class-name="rowClassName"
                    :bordered="false"
                    size="small"
                />
                <p class="text-xs text-white/50 mt-3 leading-[1.7]">
                    {{ t('activityCard.contributionNote') }}
                </p>
            </div>
        </template>
    </common-card>
</template>

<script setup lang="ts">
/**
 * @file activity-card.vue
 */
import { computed, ref, h } from 'vue';
import { useI18n } from 'vue-i18n';
import { NDataTable } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import CommonCard from '@/components/common-card.vue';
import { getInviteCode } from '@/api/mods/quota.mod';

const { t } = useI18n();

const GITHUB_URL = 'https://github.com/zgsm-ai/costrict';

const CONTRIB_ROW_HIGHLIGHTS = [true, true, false, false, false] as const;

const inviteRules = computed(() => [t('activityCard.inviteRule1'), t('activityCard.inviteRule2')]);

const contribRows = computed(() =>
    (['contrib1', 'contrib2', 'contrib3', 'contrib4', 'contrib5'] as const).map((key, idx) => ({
        action: t(`activityCard.${key}`),
        credits: t(`activityCard.${key}Credits`),
        highlight: CONTRIB_ROW_HIGHLIGHTS[idx],
    })),
);

type ContribRow = (typeof contribRows.value)[number];

const contribColumns = computed<DataTableColumns<ContribRow>>(() => [
    {
        key: 'action',
        title: t('activityCard.tableColAction'),
        render(row) {
            if (row.highlight) {
                return h('span', { class: 'contrib-action-highlight' }, [
                    h('span', { class: 'contrib-star' }, '\u2605'),
                    row.action,
                ]);
            }
            return h('span', { class: 'text-white/70' }, row.action);
        },
    },
    {
        key: 'credits',
        title: t('activityCard.tableColReward'),
        align: 'right',
        render(row) {
            return h(
                'span',
                {
                    class: [
                        'font-bold',
                        'text-[#00ffb7]',
                        'whitespace-nowrap',
                        row.highlight ? 'credits-glow' : '',
                    ],
                },
                row.credits,
            );
        },
    },
]);

const rowClassName = (row: ContribRow) => (row.highlight ? 'contrib-row-highlight' : '');

const toGithub = () => {
    window.open(GITHUB_URL);
};

const isInviteLoading = ref(false);

const toInvite = async () => {
    if (isInviteLoading.value) return;

    isInviteLoading.value = true;
    try {
        const {
            data: { invite_code = '' },
        } = await getInviteCode();
        window.open(`/credit/manager/credit-reward-plan?code=${invite_code}`);
    } catch (error) {
        console.error('获取邀请码失败:', error);
    } finally {
        isInviteLoading.value = false;
    }
};
</script>

<style scoped lang="less">
// Card item: pseudo-element + hover states (cannot be done with Tailwind in template)
.activity-card-item {
    position: relative;
    overflow: hidden;
    border-radius: 20px;
    padding: 32px 28px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.1);
    transition:
        background 0.25s ease-in-out,
        border-color 0.25s ease-in-out,
        transform 0.25s ease-in-out;

    &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        height: 2px;
        background: linear-gradient(90deg, #0066ff, #00ffb7);
        opacity: 0;
        transition: opacity 0.25s ease-in-out;
    }

    &:hover {
        background: rgba(255, 255, 255, 0.07);
        border-color: rgba(0, 102, 255, 0.35);
        transform: translateY(-2px);

        &::before {
            opacity: 1;
        }
    }
}

// CTA button: hover arrow animation
.activity-card-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 10px 22px;
    border-radius: 40px;
    font-size: 14px;
    font-weight: 600;
    color: #c3defa;
    border: 1px solid rgba(0, 102, 255, 0.5);
    text-decoration: none;
    white-space: nowrap;
    transition: all 0.25s ease-in-out;

    .arrow {
        font-size: 13px;
        transition: transform 0.25s ease-in-out;
    }

    &:hover {
        background: rgba(0, 102, 255, 0.12);
        border-color: #0066ff;
        color: #fff;
        transform: translateY(-1px);

        .arrow {
            transform: translateX(3px);
        }
    }
}

// NDataTable theme overrides
.contrib-table {
    :deep(.n-data-table),
    :deep(.n-data-table-table) {
        background: transparent;
    }

    :deep(.n-data-table-wrapper) {
        border-radius: 8px;
        overflow: hidden;
    }

    :deep(.n-data-table-thead) {
        background-color: transparent;

        .n-data-table-tr {
            background-color: transparent;
        }

        .n-data-table-th {
            background: transparent;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
            color: rgba(255, 255, 255, 0.5);
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            padding: 8px 12px;
        }
    }

    :deep(.n-data-table-tbody) {
        .n-data-table-tr {
            background: transparent;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
            transition: background 0.15s ease;

            &:last-child {
                border-bottom: none;
            }

            &:hover {
                background: rgba(255, 255, 255, 0.03) !important;
            }

            &.contrib-row-highlight {
                background: rgba(0, 255, 183, 0.06);
            }

            .n-data-table-td {
                background: transparent;
                border: none;
                padding: 10px 12px;
                color: rgba(255, 255, 255, 0.7);
                font-size: 13px;
            }
        }
    }
}

// Render function dynamic classes
:deep(.contrib-action-highlight) {
    position: relative;
    padding-left: 20px;
    color: rgba(255, 255, 255, 0.7);

    .contrib-star {
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%);
        color: #00ffb7;
        font-size: 11px;
    }
}

:deep(.credits-glow) {
    text-shadow: 0 0 8px rgba(0, 255, 183, 0.4);
}

.invite-tag {
    padding: 0 4px;
}
</style>
