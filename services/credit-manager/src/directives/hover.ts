/**
 * @file hover.ts
 * @description v-hover 自定义指令实现，用于在鼠标悬停时显示 Naive UI 的 Tooltip 提示框
 */
import { createApp, h, ref } from 'vue';
import type { DirectiveBinding } from 'vue';
import { NTooltip } from 'naive-ui';

interface HoverElement extends HTMLElement {
    _hoverApp?: {
        unmount: () => void;
    } | null;
    _tooltipVisible?: boolean;
    _mouseenterHandler?: ((event: MouseEvent) => void) | null;
    _mouseleaveHandler?: ((event: MouseEvent) => void) | null;
    _tooltipMouseEnterHandler?: ((event: MouseEvent) => void) | null;
    _tooltipMouseLeaveHandler?: ((event: MouseEvent) => void) | null;
    _value?: string;
    _tooltipContainer?: HTMLElement;
    _hideTimer?: ReturnType<typeof setTimeout> | null;
    _isMouseOverTooltip?: boolean;
    _isMouseOverElement?: boolean;
}

const hoverDirective = {
    mounted(el: HoverElement, binding: DirectiveBinding<string>) {
        // 存储指令值
        el._value = binding.value;

        // 存储状态
        el._hoverApp = null;
        el._tooltipVisible = false;
        el._hideTimer = null;
        el._isMouseOverTooltip = false;
        el._isMouseOverElement = false;

        // 创建 tooltip 容器元素，直接添加到 body
        const tooltipContainer = document.createElement('div');
        tooltipContainer.id = `tooltip-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        tooltipContainer.style.position = 'absolute';
        tooltipContainer.style.zIndex = '9999';
        tooltipContainer.style.pointerEvents = 'auto'; // 允许鼠标事件，这样用户可以与 tooltip 交互
        document.body.appendChild(tooltipContainer);

        // 创建 tooltip 应用函数
        const createTooltipApp = (
            content: string,
            isUpdated: boolean = false,
            checkAndHideTooltipFn?: () => void,
        ) => {
            if (!content || content.trim() === '') return null;

            return createApp({
                setup() {
                    const showTooltip = ref(true);
                    return () =>
                        h(
                            NTooltip,
                            {
                                show: showTooltip.value,
                                trigger: 'manual',
                                placement: 'top',
                                // 添加鼠标事件监听
                                onMouseenter: () => {
                                    el._isMouseOverTooltip = true;
                                },
                                onMouseleave: () => {
                                    el._isMouseOverTooltip = false;
                                    // 延迟检查是否应该隐藏
                                    setTimeout(() => {
                                        if (isUpdated) {
                                            // updated方法中的特殊处理
                                            if (!el._isMouseOverElement) {
                                                // 立即隐藏 tooltip
                                                el._tooltipVisible = false;
                                                if (el._hoverApp) {
                                                    el._hoverApp.unmount();
                                                    el._hoverApp = null;
                                                }
                                                if (el._tooltipContainer) {
                                                    el._tooltipContainer.style.display = 'none';
                                                    // 确保所有子元素被正确清理
                                                    while (el._tooltipContainer.firstChild) {
                                                        el._tooltipContainer.removeChild(
                                                            el._tooltipContainer.firstChild,
                                                        );
                                                    }
                                                }
                                            }
                                        } else {
                                            // mounted方法中的处理
                                            if (checkAndHideTooltipFn) {
                                                checkAndHideTooltipFn();
                                            }
                                        }
                                    }, 100);
                                },
                            },
                            {
                                default: () => content,
                                trigger: () =>
                                    h('span', {
                                        style: {
                                            display: 'inline-block',
                                            width: '100%',
                                            height: '100%',
                                        },
                                    }),
                            },
                        );
                },
            });
        };

        // 获取元素位置并更新容器位置
        const updateTooltipPosition = () => {
            const rect = el.getBoundingClientRect();

            // 设置 tooltip 容器位置
            if (el._tooltipContainer) {
                el._tooltipContainer.style.position = 'absolute';
                el._tooltipContainer.style.top = `${rect.top + window.scrollY}px`;
                el._tooltipContainer.style.left = `${rect.left + window.scrollX}px`;
                el._tooltipContainer.style.width = `${rect.width}px`;
                el._tooltipContainer.style.height = `${rect.height}px`;
                el._tooltipContainer.style.display = 'block';
            }
        };

        // 检查是否应该隐藏 tooltip
        const checkAndHideTooltip = () => {
            // 只有当鼠标既不在元素上也不在 tooltip 上时才隐藏
            if (!el._isMouseOverElement && !el._isMouseOverTooltip) {
                el._tooltipVisible = false;

                // 卸载应用实例
                if (el._hoverApp) {
                    el._hoverApp.unmount();
                    el._hoverApp = null;
                }

                // 隐藏容器
                if (el._tooltipContainer) {
                    el._tooltipContainer.style.display = 'none';
                    el._tooltipContainer.innerHTML = '';
                }
            }
        };

        // 隐藏 tooltip 的函数
        const hideTooltip = (immediate = false) => {
            if (!el._tooltipVisible) return;

            // 清除之前的隐藏定时器
            if (el._hideTimer !== null) {
                clearTimeout(el._hideTimer);
                el._hideTimer = null;
            }

            if (immediate) {
                checkAndHideTooltip();
            } else {
                // 设置 1 秒延迟隐藏
                // 使用类型断言确保兼容性
                el._hideTimer = setTimeout(checkAndHideTooltip, 1000) as unknown as ReturnType<
                    typeof setTimeout
                >;
            }
        };

        // 显示 tooltip 的函数
        const showTooltip = () => {
            if (el._tooltipVisible) return;

            // 获取当前值
            const currentValue = el._value || '';
            if (!currentValue.trim()) return;

            el._tooltipVisible = true;

            // 卸载之前的应用实例（如果存在）
            if (el._hoverApp) {
                el._hoverApp.unmount();
                el._hoverApp = null;
            }

            // 清除之前的隐藏定时器
            if (el._hideTimer !== null) {
                clearTimeout(el._hideTimer);
                el._hideTimer = null;
            }

            // 更新位置
            updateTooltipPosition();

            // 创建并挂载新的 tooltip 应用
            const tooltipApp = createTooltipApp(currentValue, false, checkAndHideTooltip);
            if (tooltipApp && el._tooltipContainer) {
                tooltipApp.mount(el._tooltipContainer);
                el._hoverApp = tooltipApp;
            }
        };

        // 鼠标进入元素事件处理
        const handleMouseEnter = () => {
            el._isMouseOverElement = true;
            showTooltip();
        };

        // 鼠标离开元素事件处理
        const handleMouseLeave = () => {
            el._isMouseOverElement = false;
            // 延迟检查是否应该隐藏，给用户一些时间移动到 tooltip
            setTimeout(() => {
                hideTooltip();
            }, 200);
        };

        // 鼠标进入 tooltip 容器事件处理
        const handleTooltipMouseEnter = () => {
            el._isMouseOverTooltip = true;
        };

        // 鼠标离开 tooltip 容器事件处理
        const handleTooltipMouseLeave = () => {
            el._isMouseOverTooltip = false;
            // 延迟检查是否应该隐藏
            setTimeout(() => {
                hideTooltip();
            }, 200);
        };

        // 添加事件监听器
        el.addEventListener('mouseenter', handleMouseEnter);
        el.addEventListener('mouseleave', handleMouseLeave);
        tooltipContainer.addEventListener('mouseenter', handleTooltipMouseEnter);
        tooltipContainer.addEventListener('mouseleave', handleTooltipMouseLeave);

        // 存储事件处理函数和容器以便在卸载时移除
        el._mouseenterHandler = handleMouseEnter;
        el._mouseleaveHandler = handleMouseLeave;
        el._tooltipMouseEnterHandler = handleTooltipMouseEnter;
        el._tooltipMouseLeaveHandler = handleTooltipMouseLeave;
        el._tooltipContainer = tooltipContainer;
    },

    updated(el: HoverElement, binding: DirectiveBinding<string>) {
        // 更新存储的值
        el._value = binding.value;

        // 如果当前 tooltip 正在显示，需要更新其内容
        if (el._tooltipVisible) {
            // 卸载之前的应用实例
            if (el._hoverApp) {
                el._hoverApp.unmount();
                el._hoverApp = null;
            }

            const currentValue = binding.value || '';

            // 检查值是否为非空字符串
            if (currentValue.trim()) {
                // 更新位置
                const rect = el.getBoundingClientRect();
                if (el._tooltipContainer) {
                    el._tooltipContainer.style.position = 'absolute';
                    el._tooltipContainer.style.top = `${rect.top + window.scrollY}px`;
                    el._tooltipContainer.style.left = `${rect.left + window.scrollX}px`;
                    el._tooltipContainer.style.width = `${rect.width}px`;
                    el._tooltipContainer.style.height = `${rect.height}px`;
                    el._tooltipContainer.style.display = 'block';
                }

                // 创建新的 tooltip 应用实例
                // 创建一个局部函数来处理 updated 中的检查逻辑
                const checkAndHideForUpdated = () => {
                    if (!el._isMouseOverElement) {
                        // 立即隐藏 tooltip
                        el._tooltipVisible = false;
                        if (el._hoverApp) {
                            el._hoverApp.unmount();
                            el._hoverApp = null;
                        }
                        if (el._tooltipContainer) {
                            el._tooltipContainer.style.display = 'none';
                            // 确保所有子元素被正确清理
                            while (el._tooltipContainer.firstChild) {
                                el._tooltipContainer.removeChild(el._tooltipContainer.firstChild);
                            }
                        }
                    }
                };

                const createTooltipAppForUpdated = (content: string) => {
                    if (!content || content.trim() === '') return null;

                    return createApp({
                        setup() {
                            const showTooltip = ref(true);
                            return () =>
                                h(
                                    NTooltip,
                                    {
                                        show: showTooltip.value,
                                        trigger: 'manual',
                                        placement: 'top',
                                        // 添加鼠标事件监听
                                        onMouseenter: () => {
                                            el._isMouseOverTooltip = true;
                                        },
                                        onMouseleave: () => {
                                            el._isMouseOverTooltip = false;
                                            // 延迟检查是否应该隐藏
                                            setTimeout(() => {
                                                checkAndHideForUpdated();
                                            }, 100);
                                        },
                                    },
                                    {
                                        default: () => content,
                                        trigger: () =>
                                            h('span', {
                                                style: {
                                                    display: 'inline-block',
                                                    width: '100%',
                                                    height: '100%',
                                                },
                                            }),
                                    },
                                );
                        },
                    });
                };

                const tooltipApp = createTooltipAppForUpdated(currentValue);

                // 挂载新的 tooltip 应用
                if (tooltipApp && el._tooltipContainer) {
                    tooltipApp.mount(el._tooltipContainer);
                    el._hoverApp = tooltipApp;
                }
            } else {
                // 隐藏容器 - 立即隐藏
                if (el._hideTimer !== null) {
                    clearTimeout(el._hideTimer);
                    el._hideTimer = null;
                }

                el._tooltipVisible = false;

                if (el._hoverApp) {
                    (el._hoverApp as { unmount: () => void }).unmount();
                    el._hoverApp = null;
                }

                if (el._tooltipContainer) {
                    el._tooltipContainer.style.display = 'none';
                    // 确保所有子元素被正确清理
                    while (el._tooltipContainer.firstChild) {
                        el._tooltipContainer.removeChild(el._tooltipContainer.firstChild);
                    }
                }
            }
        }
    },

    unmounted(el: HoverElement) {
        // 移除事件监听器
        if (el._mouseenterHandler) {
            el.removeEventListener('mouseenter', el._mouseenterHandler);
        }
        if (el._mouseleaveHandler) {
            el.removeEventListener('mouseleave', el._mouseleaveHandler);
        }

        // 移除 tooltip 容器的事件监听器
        if (el._tooltipContainer && el._tooltipMouseEnterHandler) {
            el._tooltipContainer.removeEventListener('mouseenter', el._tooltipMouseEnterHandler);
        }
        if (el._tooltipContainer && el._tooltipMouseLeaveHandler) {
            el._tooltipContainer.removeEventListener('mouseleave', el._tooltipMouseLeaveHandler);
        }

        // 卸载 tooltip 应用
        if (el._hoverApp) {
            el._hoverApp.unmount();
        }

        // 清除隐藏定时器
        if (el._hideTimer !== null) {
            clearTimeout(el._hideTimer);
            el._hideTimer = null;
        }

        // 从 DOM 中移除 tooltip 容器
        if (el._tooltipContainer) {
            el._tooltipContainer.remove();
        }

        // 清理存储的引用
        el._hoverApp = null;
        el._tooltipContainer = undefined;
        el._tooltipVisible = false;
        el._hideTimer = null;
        el._isMouseOverTooltip = false;
        el._isMouseOverElement = false;
        el._mouseenterHandler = null;
        el._mouseleaveHandler = null;
        el._tooltipMouseEnterHandler = null;
        el._tooltipMouseLeaveHandler = null;
        delete el._value;
    },
};

// 使用命名导出
export { hoverDirective };
// 同时使用默认导出
export default hoverDirective;
