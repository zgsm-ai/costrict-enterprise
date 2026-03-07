<template>
    <div class="flex items-center justify-center h-full w-full">
        <!-- ================== 1. 正常展示区域 (保持不变) ================== -->
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
                :alt="`type_${type}`"
                crossorigin="anonymous"
            />

            <div class="absolute z-2 max-w-67.75 w-full top-107.5 left-0 right-0 mx-auto">
                <p class="text-sm">{{ typeText }}</p>

                <div class="flex mt-5 justify-between items-start">
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
                        <p class="font-semibold text-[8px] mt-1">
                            {{ t('annualSummary.scanQRCode') }}
                        </p>
                    </div>

                    <div class="flex flex-col items-center">
                        <img
                            v-if="qrCodeDataUrl"
                            :src="qrCodeDataUrl"
                            alt="invite-qrcode"
                            class="w-15 h-15 block"
                        />
                    </div>
                </div>

                <div data-html2canvas-ignore>
                    <div
                        class="h-10.75 w-full text-center leading-10.75 cursor-pointer mt-4.5 font-semibold save-btn rounded text-white select-none transition-opacity"
                        :class="{ 'opacity-70 pointer-events-none': isGenerating }"
                        @click="handleSavePoster"
                    >
                        {{
                            isGenerating
                                ? t('annualSummary.generating')
                                : t('annualSummary.savePoster')
                        }}
                    </div>
                    <p class="text-[8px] text-center mt-1 text-[#A9A9A9] font-medium">
                        {{ t('annualSummary.saveFailedTip') }}
                    </p>
                    <p
                        class="text-center text-sm mt-2.5 text-[#BCBBBB] cursor-pointer font-medium"
                        @click="handleReset"
                    >
                        {{ t('annualSummary.reviewAgain') }}
                    </p>
                </div>
            </div>
        </div>

        <!-- ================== 2. 导出专用区域 (关键修改) ================== -->
        <!-- 
            修改说明：
            1. 移除 v-if/hidden，确保 DOM 始终存在。
            2. 使用 fixed 将其移出屏幕左侧 (-9999px)，并置于底层。
            3. 必须保持 visibility: visible (默认) 和 opacity: 1，否则截图也是透明的。
        -->
        <div
            class="fixed top-0 left-[-9999px] w-[375px] z-[-1]"
            style="pointer-events: none"
        >
            <div
                ref="exportWrapperRef"
                class="w-full relative bg-white"
            >
                <!-- 
                    注意：请确保 AnnualSummaryResultExport 组件内部的所有 img 标签
                    如果是网络图片，都加上了 crossorigin="anonymous" 
                -->
                <AnnualSummaryResultExport
                    :type="type"
                    :invite-code="inviteCode"
                    :login-url="loginUrl"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 年度总结结果
 */
import { ref, computed, watch, nextTick } from 'vue';
import { toPng } from 'html-to-image';
import QRCode from 'qrcode';
import { getT } from '@/utils/i18n';
import { TYPE_IMAGE_MAP, TYPE_TEXT_MAP, DEFAULT_RESULT_TYPE } from '../const';
import AnnualSummaryResultExport from './annual-summary-result-export.vue';
import { recordShare } from '@/api/mods/activity.mod';

const t = getT();

const props = withDefaults(
    defineProps<{
        stepId: number;
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

const emit = defineEmits<{
    reset: [];
}>();

const posterRef = ref<HTMLElement>();
const exportWrapperRef = ref<HTMLElement>();
const isGenerating = ref(false);
const qrCodeDataUrl = ref('');

const typeImageSrc = computed(
    () => TYPE_IMAGE_MAP[props.type] || TYPE_IMAGE_MAP[DEFAULT_RESULT_TYPE],
);
const typeText = computed(() => TYPE_TEXT_MAP[props.type] || TYPE_TEXT_MAP[DEFAULT_RESULT_TYPE]);

// 生成二维码
const generateQRCode = async (url: string) => {
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

const handleReset = () => emit('reset');

/**
 * 辅助函数：检测目标元素内的所有图片是否加载完毕
 * 这是解决 Safari 白屏的关键步骤之一
 */
const ensureImagesLoaded = async (element: HTMLElement) => {
    const images = Array.from(element.querySelectorAll('img'));

    const promises = images.map((img) => {
        // 如果图片已经完成且有高度，直接成功
        if (img.complete && img.naturalHeight > 0) return Promise.resolve();

        // 否则等待加载事件
        return new Promise<void>((resolve) => {
            img.onload = () => resolve();
            img.onerror = () => resolve(); // 即使失败也 resolve，防止死锁
        });
    });

    // 最多等待 3 秒，防止某张图挂了导致程序卡死
    await Promise.race([
        Promise.all(promises),
        new Promise((resolve) => setTimeout(resolve, 3000)),
    ]);
};

const handleSavePoster = async () => {
    if (isGenerating.value) return;

    // 1. 确保二维码存在
    if (!qrCodeDataUrl.value && props.loginUrl) {
        await generateQRCode(props.loginUrl);
    }

    // 2. 埋点
    try {
        await recordShare();
    } catch (error) {
        console.error('Record share error:', error);
    }

    isGenerating.value = true;

    try {
        // 等待 Vue DOM 更新
        await nextTick();

        if (!exportWrapperRef.value) {
            throw new Error('Export wrapper not found');
        }

        // 3. 显式等待图片加载
        await ensureImagesLoaded(exportWrapperRef.value);

        // 4. 额外缓冲时间 (给 Safari 渲染管线一点时间)
        await new Promise((resolve) => setTimeout(resolve, 500));

        const targetNode = exportWrapperRef.value;
        const contentHeight = targetNode.scrollHeight || 667;

        // ★★★ 核心配置 ★★★
        const options = {
            // Safari 必须禁用 cacheBust，否则 CORS 会挂
            cacheBust: false,
            // 像素比：2 倍通常足够清晰，过高会导致 iOS Safari 内存溢出白屏
            pixelRatio: 2,
            backgroundColor: '#ffffff',
            width: 375,
            height: contentHeight,
            style: {
                transform: 'none', // 确保没有 transform 干扰
                margin: '0',
            },
            // 过滤掉不需要的标签
            filter: (node: HTMLElement) => {
                return (
                    node.tagName !== 'LINK' && node.tagName !== 'STYLE' && node.tagName !== 'SCRIPT'
                );
            },
        };

        // 5. 双重渲染机制 (Double Render)
        // 第一次渲染：强制浏览器“预热” Canvas 上下文，加载字体和解码图片
        try {
            await toPng(targetNode, options);
        } catch (e) {
            // 忽略第一次可能的报错
            console.warn('First pass render warning:', e);
        }

        // 稍微停顿
        await new Promise((resolve) => setTimeout(resolve, 100));

        // 第二次渲染：正式生成
        const dataUrl = await toPng(targetNode, options);

        // 6. 下载处理
        const link = document.createElement('a');
        link.download = `${t('annualSummary.yearlySummary')}_${new Date().getTime()}.png`;
        link.href = dataUrl;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    } catch (error) {
        console.error('海报生成失败:', error);
        // 这里可以加一个 Toast 提示用户重试
    } finally {
        isGenerating.value = false;
    }
};
</script>

<style scoped lang="less">
.logo-text {
    background: linear-gradient(0deg, #2b7df7 25%, rgba(47, 130, 255, 0.83) 70%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    color: transparent;
}
.save-btn {
    background: linear-gradient(0deg, #2b7df7 25%, rgba(47, 130, 255, 0.83) 70%);
}
</style>
