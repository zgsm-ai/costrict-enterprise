<template>
    <div class="credit-reward-plan mt-10">
        <div class="text-center text-[28px] font-semibold">{{ t('rewardPlan.title') }}</div>

        <div class="mt-8">
            <p>{{ t('rewardPlan.introTitle') }}</p>
            <p>
                {{ t('rewardPlan.introContent') }}
            </p>
        </div>

        <!-- 邀请人展示 -->
        <template v-if="!isInvite">
            <!-- 规则说明 -->
            <reward-card
                :title="t('rewardPlan.rulesTitle')"
                class="mt-10"
            >
                <template #content>
                    <p>
                        {{ t('rewardPlan.rules.rule1')
                        }}<span class="credit-unit">{{ t('rewardPlan.rules.rule1NewUser') }}</span
                        >{{ t('rewardPlan.rules.rule1Inviter')
                        }}<span class="credit-unit">{{
                            t('rewardPlan.rules.rule1InviterCredits')
                        }}</span
                        >{{ t('rewardPlan.rules.rule1End') }}
                    </p>
                    <div class="mt-1">
                        <p>
                            {{ t('rewardPlan.rules.rule2') }}
                        </p>
                        <p class="ml-4">
                            <span class="opacity-70">a. </span>
                            {{ t('rewardPlan.rules.rule2a')
                            }}<span class="credit-unit">{{
                                t('rewardPlan.rules.rule2aCredits')
                            }}</span
                            >{{ t('rewardPlan.rules.rule2aTotal')
                            }}<span class="credit-unit">{{
                                t('rewardPlan.rules.rule2aTotalCredits')
                            }}</span
                            >{{ t('rewardPlan.rules.rule2aEnd') }}
                        </p>
                        <p class="ml-4">
                            <span class="opacity-70">b. </span>
                            {{ t('rewardPlan.rules.rule2b')
                            }}<span class="credit-unit">{{
                                t('rewardPlan.rules.rule2bCredits')
                            }}</span
                            >{{ t('rewardPlan.rules.rule2bTotal')
                            }}<span class="credit-unit">{{
                                t('rewardPlan.rules.rule2bTotalCredits')
                            }}</span
                            >{{ t('rewardPlan.rules.rule2bEnd') }}
                        </p>
                    </div>
                    <p class="mt-1">{{ t('rewardPlan.rules.rule3') }}</p>
                </template>
            </reward-card>

            <!-- 邀请码 -->
            <reward-card
                :title="
                    inviteCode ? t('rewardPlan.invitationCode') : t('rewardPlan.invitationMethod')
                "
                class="mt-6"
                v-if="inviteCode"
            >
                <template
                    #label
                    v-if="inviteCode"
                >
                    <div class="flex items-center">
                        {{ t('rewardPlan.invitationCodeLabel') }}
                        <span
                            class="text-[#12FFBB]"
                            :class="{ 'ml-1': !isZh }"
                            >{{ inviteCode }}</span
                        >
                        <n-icon
                            size="14"
                            class="ml-2 cursor-pointer text-[#197DFF]"
                            @click="copyCode"
                        >
                            <copy-outline />
                        </n-icon>
                    </div>
                </template>
                <template #content>
                    <template v-if="inviteCode">
                        <p class="text-xs opacity-70 mt-2">
                            {{ t('rewardPlan.invitationCodeText') }}
                        </p>
                        <div class="mt-2 flex items-center">
                            <span :class="{ 'mr-1': !isZh }">{{
                                t('rewardPlan.exclusiveInvitationLink')
                            }}</span>
                            <n-button
                                type="info"
                                @click="copyInviteLink"
                                size="tiny"
                            >
                                {{ t('rewardPlan.copyLink') }}
                            </n-button>
                        </div>
                    </template>
                </template>
            </reward-card>

            <!-- 如何参与 -->
            <!-- <reward-card
                class="mt-6"
                :title="t('rewardPlan.howToParticipate')"
                content-class="mt-5"
            >
                <template #content>
                    <div class="flex items-center">
                        {{ t('rewardPlan.participationStep1') }}
                        <div class="ml-3">
                            <n-button
                                type="info"
                                @click="copyInviteLink"
                                size="tiny"
                            >
                                {{ t('rewardPlan.copyLink') }}
                            </n-button>
                        </div>
                    </div>
                    <p>{{ t('rewardPlan.participationStep2') }}</p>
                    <p>{{ t('rewardPlan.participationStep3') }}</p>
                </template>
            </reward-card> -->
        </template>

        <!-- 操作指引 -->
        <reward-card
            class="mt-10"
            content-class="mt-5"
        >
            <template #header>
                <div class="flex opertion-rule self-start items-center">
                    <span class="text-xl">{{ t('rewardPlan.operationGuideTitle') }}</span>
                    <!-- <span
                        class="text-sm"
                        :class="{ 'ml-1': !isZh }"
                        >{{ t('rewardPlan.operationGuideSubtitle') }}</span
                    > -->
                    <div
                        class="ml-2 mt-0.5"
                        v-if="inviteCode"
                    >
                        <n-button
                            type="info"
                            @click="copyLoginLink"
                            size="tiny"
                        >
                            {{ t('rewardPlan.copyLoginLink') }}
                        </n-button>
                    </div>
                </div>
            </template>

            <template #content>
                <div
                    v-for="(item, index) in operationGuide"
                    :key="item.title"
                    :class="{ 'mt-4': index }"
                >
                    <operation-card :title="item.title" />

                    <n-timeline
                        class="mt-4"
                        size="large"
                    >
                        <n-timeline-item
                            v-for="step in item.steps"
                            :key="step.index"
                            :title="step.header"
                            line-type="dashed"
                        >
                            <template
                                #header
                                v-if="step.headerRender"
                            >
                                <component :is="step.headerRender()"></component>
                            </template>
                            <template #icon>
                                <div class="custom-timeline-icon">{{ step.index }}</div>
                            </template>
                            <template #default>
                                <div class="text-white leading-5 font-medium">
                                    <p>{{ step.content }}</p>
                                    <p
                                        v-if="step.tips"
                                        class="font-normal text-white/70"
                                    >
                                        {{ step.tips }}
                                    </p>
                                </div>
                                <img
                                    v-if="step.image"
                                    class="mt-4 mb-10"
                                    :class="step.imageClass"
                                    :src="step.image"
                                    alt=""
                                />
                            </template>
                        </n-timeline-item>
                    </n-timeline>
                </div>
            </template>
        </reward-card>

        <!-- 被邀请人展示 -->
        <template v-if="isInvite">
            <!-- 邀请码 -->
            <reward-card
                :title="t('rewardPlan.invitationCode')"
                class="mt-6"
                :text="t('rewardPlan.invitationCodeText')"
            >
                <template #label>
                    <div class="flex items-center">
                        {{ t('rewardPlan.invitationCodeLabel') }}
                        <span
                            class="text-[#12FFBB]"
                            :class="{ 'ml-1': !isZh }"
                            >{{ inviteCode }}</span
                        >
                        <n-icon
                            size="14"
                            class="ml-2 cursor-pointer text-[#197DFF]"
                            @click="copyCode"
                        >
                            <copy-outline />
                        </n-icon>
                    </div>
                </template>
                <template #content>
                    <p class="text-xs opacity-70 mt-2">{{ t('rewardPlan.invitationCodeText') }}</p>
                    <div class="mt-2">
                        <n-button
                            type="info"
                            @click="copyInviteLink"
                            size="tiny"
                        >
                            {{ t('rewardPlan.copyLink') }}
                        </n-button>
                    </div>
                </template>
            </reward-card>

            <!-- 规则说明 -->
            <reward-card
                :title="t('rewardPlan.rulesTitle')"
                class="mt-10"
            >
                <template #content>
                    <p>
                        {{ t('rewardPlan.rules.rule1')
                        }}<span class="credit-unit">{{ t('rewardPlan.rules.rule1NewUser') }}</span
                        >{{ t('rewardPlan.rules.rule1Inviter')
                        }}<span class="credit-unit">{{
                            t('rewardPlan.rules.rule1InviterCredits')
                        }}</span
                        >{{ t('rewardPlan.rules.rule1End') }}
                    </p>
                    <div class="mt-1">
                        <p>
                            {{ t('rewardPlan.rules.rule2') }}
                        </p>
                        <p class="ml-4">
                            <span class="opacity-70">a. </span>
                            {{ t('rewardPlan.rules.rule2a')
                            }}<span class="credit-unit">{{
                                t('rewardPlan.rules.rule2aCredits')
                            }}</span
                            >{{ t('rewardPlan.rules.rule2aTotal')
                            }}<span class="credit-unit">{{
                                t('rewardPlan.rules.rule2aTotalCredits')
                            }}</span
                            >{{ t('rewardPlan.rules.rule2aEnd') }}
                        </p>
                        <p class="ml-4">
                            <span class="opacity-70">b. </span>
                            {{ t('rewardPlan.rules.rule2b')
                            }}<span class="credit-unit">{{
                                t('rewardPlan.rules.rule2bCredits')
                            }}</span
                            >{{ t('rewardPlan.rules.rule2bTotal')
                            }}<span class="credit-unit">{{
                                t('rewardPlan.rules.rule2bTotalCredits')
                            }}</span
                            >{{ t('rewardPlan.rules.rule2bEnd') }}
                        </p>
                    </div>
                    <p class="mt-1">{{ t('rewardPlan.rules.rule3') }}</p>
                </template>
            </reward-card>
        </template>

        <!-- 关于Credits -->
        <reward-card
            :title="t('rewardPlan.aboutCreditsTitle')"
            class="mt-12.5"
        >
            <template #content>
                <template
                    v-for="(item, index) in qaContent"
                    :key="index"
                >
                    <credit-qa-card
                        :question="item.question"
                        :answer="item.answer"
                        :class="{ 'mt-12': !!index }"
                    />
                </template>
            </template>
        </reward-card>

        <!-- 联系我们 -->
        <reward-card
            :title="t('rewardPlan.contactUsTitle')"
            :description="t('rewardPlan.contactUsDescription')"
            class="mt-12.5 mb-35"
            contentClass="mt-5"
        >
            <template #content>
                <div class="flex items-center mt-4">
                    <div class="flex flex-col items-center">
                        <img
                            class="w-25 h-25"
                            src="../../assets/qrcode/official_account.png"
                            alt=""
                        />
                        <span class="mt-3">{{ t('rewardPlan.officialAccount') }}</span>
                    </div>

                    <div class="flex flex-col ml-20 items-center">
                        <img
                            class="w-25 h-25"
                            src="../../assets/qrcode/communication_group.png"
                            alt=""
                        />
                        <span class="mt-3">{{ t('rewardPlan.joinGroup') }}</span>
                    </div>
                </div>
            </template>
        </reward-card>
    </div>
</template>

<script setup lang="ts">
/**
 * @file credit-reward-plan.vue
 * Credit奖励计划页面
 */
import { copyToClipboard } from '@/utils/copy';
import RewardCard from './reward-card.vue';
import CreditQaCard from './credit-qa-card.vue';
import { CopyOutline } from '@vicons/ionicons5';
import { NIcon, useMessage, NTimeline, NTimelineItem, NButton } from 'naive-ui';
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import OperationCard from './operation-card.vue';
import { createOperationGuide, createQaContent } from './const';
import { getLoginUrl } from '@/api/mods/quota.mod';
import { useMetaTags } from '@/composables/useMetaTags';
import { ref as refVue } from 'vue';
import type { SocialMetaConfig } from '@/config/meta';

// 为积分奖励计划页面配置特定的Open Graph标签
const customMeta = refVue<SocialMetaConfig>({
    og: {
        title: 'CoStrict 喊你来薅羊毛啦（送 Credits）！',
        description:
            'CoStrict-开源免费 AI 编程工具，为企业严肃编程量身打造。邀请好友注册，双方均可获得Credits奖励！',
        image: '/social-share.png',
        type: 'website',
        siteName: 'CoStrict',
        locale: 'zh_CN',
    },
    twitter: {
        card: 'summary_large_image',
        title: 'CoStrict 喊你来薅羊毛啦（送 Credits）！',
        description:
            'CoStrict-开源免费 AI 编程工具，为企业严肃编程量身打造。邀请好友注册，双方均可获得Credits奖励！',
        image: '/social-share.png',
    },
});

// 使用自定义meta标签
useMetaTags(customMeta);

const { t, locale } = useI18n();
const isZh = computed(() => locale.value === 'zh');
const message = useMessage();
const route = useRoute();

const inviteCode = ref('');
const isInvite = ref(false);
const loginUrl = ref('');

const fetchLoginUrl = async () => {
    try {
        const {
            data: { url },
        } = await getLoginUrl({
            inviter_code: inviteCode.value,
        });
        loginUrl.value = url;
    } catch (error) {
        console.error('获取登录链接失败:', error);
    }
};

// 从 URL 获取参数
onMounted(() => {
    const invite = route.query.invite as string;
    const code = route.query.code as string;

    isInvite.value = invite === 'true';
    inviteCode.value = code;

    fetchLoginUrl();
});

const copyCode = () => {
    copyToClipboard(inviteCode.value, {
        success: message.success,
        error: message.error,
    });
};

const copyLoginLink = () => {
    const currentUrl = window.location.origin + window.location.pathname;
    const inviteUrl = `${currentUrl}?invite=true&code=${inviteCode.value}`;
    copyToClipboard(inviteUrl, {
        success: message.success,
        error: message.error,
    });
};

const copyInviteLink = () => {
    copyToClipboard(loginUrl.value, {
        success: message.success,
        error: message.error,
    });
};

// 创建翻译后的数据
const operationGuide = computed(() => createOperationGuide(t, loginUrl.value, isZh.value));
const qaContent = computed(() => createQaContent(t, isZh.value));
</script>

<style lang="less" scoped>
.credit-reward-plan {
    // Credit奖励计划页面样式
    width: 1105px;
    margin: auto;
}

.opertion-rule {
    span {
        background: linear-gradient(91deg, #005eff -9%, #00ffb7 95%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }
}

.custom-timeline-icon {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: #00ca90;
    border: 2px solid #00ca90;
    font-size: 12px;
    color: #ffffff;
    font-weight: 500;
    display: flex;
    align-items: center;
    justify-content: center;
}

.credit-unit {
    color: rgba(18, 255, 187);
}
</style>
