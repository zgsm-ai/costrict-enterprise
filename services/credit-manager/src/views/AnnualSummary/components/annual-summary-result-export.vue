<template>
    <div class="flex items-center justify-center h-full w-full">
        <div
            ref="posterRef"
            class="relative max-w-93.75 aspect-375/667 w-full bg-white overflow-hidden"
        >
            <img
                class="block absolute max-w-93.75 z-1 inset-0 w-full h-full object-cover"
                src="../../../assets/summary/result_bg.webp"
                alt="summary_cover"
            />

            <img
                class="block absolute z-2 w-full"
                :src="typeImageSrc"
                :alt="`export_${type}`"
                crossorigin="anonymous"
            />

            <div class="absolute z-2 max-w-67.75 w-full top-120.5 left-0 right-0 mx-auto">
                <p class="text-sm">{{ typeText }}</p>

                <div class="flex mt-7 justify-between items-start">
                    <div class="flex flex-col justify-start">
                        <div class="flex items-center">
                            <img
                                src="../../../assets/logo.webp"
                                alt="logo"
                                class="w-3.75"
                                crossorigin="anonymous"
                            />
                            <span class="text-xs ml-0.5 logo-text font-semibold">CoStrict</span>
                        </div>
                        <p class="font-semibold text-[8px] mt-1">
                            {{ t('annualSummary.annualReport') }}
                        </p>
                        <p class="font-semibold text-[8px] mt-0.5">
                            {{ t('annualSummary.scanQRCode') }}
                        </p>
                        <p
                            v-if="inviteCode"
                            class="font-semibold text-xs mt-2"
                        >
                            {{ t('annualSummary.inviteCode')
                            }}<span class="text-[#12FFBB]">{{ inviteCode }}</span>
                        </p>
                    </div>

                    <div class="flex gap-3 items-center">
                        <div
                            class="flex flex-col items-center"
                            v-if="qrCodeDataUrl"
                        >
                            <img
                                :src="qrCodeDataUrl"
                                alt="invite-qrcode"
                                class="w-15 h-15 block"
                            />
                            <span class="text-[10px] text-[#D1D1D1] mt-1.5 whitespace-nowrap">
                                {{ t('annualSummary.loginExperience') }}
                            </span>
                        </div>

                        <div class="flex flex-col items-center">
                            <img
                                src="../../../assets/summary/qr.webp"
                                alt="official-qrcode"
                                class="w-15 h-15 block"
                                crossorigin="anonymous"
                            />
                            <span class="text-[10px] text-[#D1D1D1] mt-1.5 whitespace-nowrap">
                                {{ t('annualSummary.learnProduct') }}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 年度总结结果导出组件
 */
import { ref, computed, watch } from 'vue';
import QRCode from 'qrcode';
import { getT } from '@/utils/i18n';
import { EXPORT_IMAGE_MAP, TYPE_TEXT_MAP, DEFAULT_RESULT_TYPE } from '../const';

const t = getT();

const props = withDefaults(
    defineProps<{
        type?: string;
        inviteCode?: string;
        loginUrl?: string;
    }>(),
    {
        type: DEFAULT_RESULT_TYPE,
        inviteCode: '',
        loginUrl: '',
    },
);

const typeImageSrc = computed(
    () => EXPORT_IMAGE_MAP[props.type] || EXPORT_IMAGE_MAP[DEFAULT_RESULT_TYPE],
);
const typeText = computed(() => TYPE_TEXT_MAP[props.type] || TYPE_TEXT_MAP[DEFAULT_RESULT_TYPE]);

const qrCodeDataUrl = ref('');

const generateQRCode = async (url: string) => {
    console.log(url);
    if (!url) return;
    try {
        const dataUrl = await QRCode.toDataURL(url, {
            width: 200,
            margin: 1,
            color: { dark: '#000000', light: '#FFFFFF' },
        });
        qrCodeDataUrl.value = dataUrl;
    } catch (error) {
        console.error(error);
    }
};

watch(
    () => props.loginUrl,
    (newUrl) => {
        if (newUrl) generateQRCode(newUrl);
    },
    { immediate: true },
);
</script>

<style scoped lang="less">
.logo-text {
    background: linear-gradient(0deg, #2b7df7 25%, rgba(47, 130, 255, 0.83) 70%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    color: transparent;
}
</style>
