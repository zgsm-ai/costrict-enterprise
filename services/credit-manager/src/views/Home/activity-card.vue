<template>
    <common-card :content-border="false">
        <template #header>
            <div class="activity-title text-xl mb-5">
                <span>{{ t('homePage.activityTitle') }}</span>
            </div>
        </template>
        <template #default>
            <div class="activity-content">
                <div class="content-title text-base flex">
                    {{ t('homePage.creditPlan')
                    }}<img
                        src="../../assets/activity.svg"
                        alt=""
                        class="ml-2"
                    />
                </div>
                <div class="content-desc">{{ t('homePage.planDesc') }}</div>
            </div>
            <common-card class="mt-3 relative">
                <template #default>
                    <div class="activity-github-star">
                        <div class="activity-github-star__header flex">
                            <img
                                class="w-3"
                                src="../../assets/star.svg"
                                alt=""
                            />
                            <span class="text-white ml-3 text-sm">{{
                                t('activityCard.githubStar')
                            }}</span>
                        </div>
                        <div
                            class="activity-github-star__content text-sm ml-6 text-white mt-3 max-w-[70%]"
                        >
                            {{ t('activityCard.githubStarDesc') }}
                        </div>
                        <div
                            class="activity-github-star__btn rounded-[20px] absolute right-4 top-7.5 px-4 py-1 w-29 h-7 flex items-center justify-center cursor-pointer"
                            @click="toGithub"
                        >
                            <span class="text-sm">{{ t('activityCard.gitStarNow') }}</span>
                        </div>
                    </div>
                </template>
            </common-card>
            <common-card class="mt-3 relative">
                <template #default>
                    <div class="activity-invite-user">
                        <div class="activity-invite-user__header flex">
                            <img
                                class="w-3"
                                src="../../assets/user.svg"
                                alt=""
                            />
                            <span class="text-white ml-3 text-sm">{{
                                t('activityCard.inviteNewUser')
                            }}</span>
                        </div>
                        <div class="activity-invite-user__content text-sm ml-6 text-white mt-3">
                            {{ t('activityCard.inviteNewUserDesc') }}
                        </div>
                        <div
                            class="activity-invite-user__btn rounded-[20px] absolute right-4 top-7.5 px-4 py-1 w-27 h-7 flex items-center justify-center"
                            :class="{
                                'cursor-pointer': !isInviteLoading,
                                'cursor-not-allowed opacity-50': isInviteLoading,
                            }"
                            @click="toInvite"
                        >
                            <span class="text-sm">{{ t('activityCard.goInvite') }}</span>
                        </div>
                    </div>
                    <ul class="invite-ser__list ml-6 text-white/70 leading-7">
                        <template
                            v-for="(item, index) in inviteRules"
                            :key="index"
                        >
                            <li>{{ item.text }}</li>
                            <template v-if="item.children">
                                <li
                                    v-for="(child, childIndex) in item.children"
                                    :key="`${index}-${childIndex}`"
                                    class="invite-ser__subitem indent-6"
                                >
                                    {{ child.text }}
                                </li>
                            </template>
                        </template>
                    </ul>
                </template>
            </common-card>
            <common-card
                v-if="isZh"
                class="mt-3 relative"
            >
                <template #default>
                    <div class="activity-annual-report">
                        <div class="activity-annual-report__header flex">
                            <img
                                class="w-3"
                                src="../../assets/calendar.svg"
                                alt=""
                            />
                            <span class="text-white ml-3 text-sm">{{
                                t('activityCard.annualReport')
                            }}</span>
                        </div>
                        <div
                            class="activity-annual-report__content text-sm ml-6 text-white mt-3 max-w-[70%]"
                        >
                            {{ t('activityCard.annualReportDesc') }}
                        </div>
                        <div
                            class="activity-annual-report__btn rounded-[20px] absolute right-4 top-7.5 px-4 py-1 w-29 h-7 flex items-center justify-center cursor-pointer"
                            @click="toAnnualReport"
                        >
                            <span class="text-sm">{{ t('activityCard.goAnnualReport') }}</span>
                        </div>
                    </div>
                </template>
            </common-card>
        </template>
    </common-card>
</template>

<script setup lang="ts">
/**
 * @file activity-card.vue
 */
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import CommonCard from '@/components/common-card.vue';
import { getInviteCode } from '@/api/mods/quota.mod';

const { t, locale } = useI18n();

const isZh = computed(() => locale.value === 'zh');

const inviteRules = computed(() => [
    {
        text: t('activityCard.inviteRules.rule1'),
    },
    {
        text: t('activityCard.inviteRules.rule2'),
        children: [
            { text: t('activityCard.inviteRules.rule2a') },
            { text: t('activityCard.inviteRules.rule2b') },
        ],
    },
    {
        text: t('activityCard.inviteRules.rule3'),
    },
]);

const toGithub = () => {
    window.open('https://github.com/zgsm-ai/costrict');
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

const toAnnualReport = () => {
    window.location.href = '/credit/manager/annual-summary';
};
</script>

<style scoped lang="less">
.activity-title {
    display: flex;
    align-items: center;

    span {
        color: #fff;
    }

    img {
        margin-left: 10px;
    }
}

.activity-content {
    display: flex;
    flex-direction: column;

    .content-desc {
        opacity: 0.7;
        margin-top: 8px;
    }
}

.activity-tips {
    margin-top: 16px;
    padding: 7px 8px 7px 12px;
    box-sizing: border-box;
    border-radius: 20px;
    display: flex;
    align-items: center;
    background: url('../../assets/tips-bg.png') no-repeat;
    background-size: cover;
    cursor: pointer;

    .tips-content {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .enter-code {
        margin-left: 10px;
        color: #1876f2;
        cursor: pointer;
        font-weight: 600;
        white-space: nowrap;
    }

    img {
        margin-left: 20px;
    }
}

.activity-github-star {
    &__header {
        span {
            letter-spacing: normal;
            background: linear-gradient(95deg, #00ffb7 0%, #ffffff 68%, #c5dbff 101%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }

    &__btn {
        border: 0.5px solid transparent;
        background:
            linear-gradient(#000, #000) padding-box,
            linear-gradient(84deg, #0167ff -2%, #48ffcb 52%, #005eff 115%) border-box;

        span {
            letter-spacing: 0.2px;
            background: linear-gradient(96deg, #00ffb7 0%, #ffffff 68%, #c5dbff 101%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }
}

.activity-invite-user {
    &__header {
        span {
            letter-spacing: normal;
            background: linear-gradient(95deg, #00ffb7 0%, #ffffff 68%, #c5dbff 101%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }

    &__btn {
        border: 0.5px solid transparent;
        background:
            linear-gradient(#000, #000) padding-box,
            linear-gradient(84deg, #0167ff -2%, #48ffcb 52%, #005eff 115%) border-box;

        span {
            letter-spacing: 0.2px;
            background: linear-gradient(96deg, #00ffb7 0%, #ffffff 68%, #c5dbff 101%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }

    &__btn {
        border: 0.5px solid transparent;
        background:
            linear-gradient(#000, #000) padding-box,
            linear-gradient(84deg, #0167ff -2%, #48ffcb 52%, #005eff 115%) border-box;

        span {
            letter-spacing: 0.2px;
            background: linear-gradient(96deg, #00ffb7 0%, #ffffff 68%, #c5dbff 101%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }
}

.activity-annual-report {
    &__header {
        span {
            letter-spacing: normal;
            background: linear-gradient(95deg, #00ffb7 0%, #ffffff 68%, #c5dbff 101%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }

    &__btn {
        border: 0.5px solid transparent;
        background:
            linear-gradient(#000, #000) padding-box,
            linear-gradient(84deg, #0167ff -2%, #48ffcb 52%, #005eff 115%) border-box;

        span {
            letter-spacing: 0.2px;
            background: linear-gradient(96deg, #00ffb7 0%, #ffffff 68%, #c5dbff 101%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }
}
</style>
