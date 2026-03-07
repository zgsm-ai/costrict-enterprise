<template>
    <div
        class="common-footer"
        v-if="!isRewardPlanPage"
    >
        <div class="contract flex items-center">
            <!-- <div class="flex flex-col items-center">
                <img
                    class="w-25 h-25"
                    src="../assets/qrcode/official_account.png"
                    alt=""
                />
                <span class="mt-3">{{ t('rewardPlan.officialAccount') }}</span>
            </div>

            <div class="flex flex-col ml-20 items-center">
                <img
                    class="w-25 h-25"
                    src="../assets/qrcode/communication_group.png"
                    alt=""
                />
                <span class="mt-3">{{ t('rewardPlan.joinGroup') }}</span>
            </div> -->
        </div>
        <div class="community mt-5">
            <div class="community-item">
                <p class="item-desc">
                    <span class="label">{{ t('common.footer.emailContactLabel') }}</span>
                    <span class="divider">|</span>
                    <span class="contact">zgsm@sangfor.com.cn</span>
                </p>
            </div>
        </div>
        <div class="copyright">
            <p>{{ t('common.footer.copyright') }}</p>
            <template v-if="isZh">
                <p>
                    <img src="@/assets/copyright_btn.png" />
                    <span>粤ICP备08126214号</span>
                </p>
                <p>
                    <img src="@/assets/copyright_btn.png" />
                    <span>粤公网安备44030502002384号</span>
                </p>
                <p>合字B1.B2-20220041</p>
            </template>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import { useRouter } from 'vue-router';

const { locale, t } = useI18n();

const isZh = computed(() => locale.value === 'zh');

const router = useRouter();

const isRewardPlanPage = computed(() => router.currentRoute.value.path === '/credit-reward-plan');
</script>

<style scoped lang="less">
.common-footer {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-top: 30px;

    .community {
        .community-item {
            .item-desc {
                display: flex;
                align-items: center;

                .label {
                    color: #a1a7b3;
                }

                .divider {
                    margin: 0 12px;
                }

                .conteact {
                    font-weight: 500;
                }
            }
        }
    }

    .copyright {
        display: flex;
        align-items: center;
        font-size: 12px;
        color: #a1a7b3;
        margin-bottom: 20px;
        margin-top: 12px;
        flex-wrap: wrap;

        @media (max-width: 830px) {
            flex-direction: column;
        }

        p {
            display: flex;
            align-items: center;
            margin: 0 12px;

            img {
                width: 16px;
                height: 16px;
            }
        }
    }
}
</style>
