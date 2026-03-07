import { ref, computed, onMounted, onUnmounted } from 'vue';
import type { Ref, ComputedRef } from 'vue';
import { useI18n } from 'vue-i18n';

export interface UseResponsiveReturn {
    isMobileLayout: Ref<boolean>;
    screenWidth: Ref<number>;
    isSmallScreen: ComputedRef<boolean>;
    responsiveClass: ComputedRef<string>;
}

export function useResponsive(
    mobileBreakpoint = 1440,
    smallScreenZhBreakpoint = 768,
    smallScreenEnBreakpoint = 830,
): UseResponsiveReturn {
    const { locale } = useI18n();
    const isZh = computed(() => locale.value === 'zh');

    // 响应式布局状态
    const isMobileLayout = ref(false);
    const screenWidth = ref(window.innerWidth);

    // 检查屏幕宽度
    const checkScreenWidth = () => {
        const width = window.innerWidth;
        screenWidth.value = width;
        isMobileLayout.value = width <= mobileBreakpoint;
    };

    // 新增计算属性：判断页面宽度是否小于等于640px（中文版）或768px（英文版）
    const isSmallScreen = computed(() => {
        const threshold = isZh.value ? smallScreenZhBreakpoint : smallScreenEnBreakpoint;
        return screenWidth.value <= threshold;
    });

    // 新增计算属性：用于动态应用CSS类名
    const responsiveClass = computed(() => {
        return isZh.value ? 'responsive-zh' : 'responsive-en';
    });

    // 监听窗口大小变化
    onMounted(() => {
        checkScreenWidth();
        window.addEventListener('resize', checkScreenWidth);
    });

    onUnmounted(() => {
        window.removeEventListener('resize', checkScreenWidth);
    });

    return {
        isMobileLayout,
        screenWidth,
        isSmallScreen,
        responsiveClass,
    };
}
