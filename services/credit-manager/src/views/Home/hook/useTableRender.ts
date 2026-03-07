/**
 * @file useTableRender.ts
 * @description 表格渲染相关的 hook，处理空值显示
 */
import type { DataTableColumn } from 'naive-ui';
import type { VNodeChild } from 'vue';

/**
 * 为表格列配置添加默认的 render 函数
 * @param defaultValue 当值为空时的默认显示值，默认为 '-'
 * @returns 一个处理函数，接受列配置数组并返回处理后的配置
 */
export function withDefaultRender<T = Record<string, unknown>>(defaultValue = '-') {
    return (columns: DataTableColumn<T>[]): DataTableColumn<T>[] => {
        return columns.map((column) => {
            // 如果已经有 render 函数，保持不变
            if ('render' in column && column.render) {
                return column;
            }

            // 确保列有 key 属性
            if (!('key' in column)) {
                return column;
            }

            // 为没有 render 函数的列添加默认渲染逻辑
            return {
                width: 120,
                ...column,
                render: (row: T): VNodeChild => {
                    const value = row[column.key as keyof T];
                    // 处理值为 null、undefined、空字符串的情况
                    // 但保留值为 0 的情况
                    return value || value === 0 ? String(value) : defaultValue;
                },
            } as DataTableColumn<T>;
        });
    };
}

/**
 * 格式化金额显示
 * @param amount 金额数值
 * @param currency 货币符号，默认为 '￥'
 * @param defaultValue 当值为空时的默认显示值
 * @returns 格式化后的金额字符串
 */
export function formatAmount(
    amount: number | string | null | undefined,
    currency = '￥',
    defaultValue = '-',
): string {
    if (amount === null || amount === undefined || amount === '') {
        return defaultValue;
    }
    return `${currency}${amount}`;
}
