/**
 * @file 运营活动模块类型文件
 */

import type { VNode } from 'vue';

// 操作指引步骤接口定义
export interface OperationStep {
    index: number;
    header?: string;
    headerRender?: () => VNode | null;
    content?: string;
    tips?: string;
    image?: string;
    imageClass?: string;
}

// 操作指引整体接口定义
export interface OperationGuide {
    title: string;
    steps: OperationStep[];
}

type AnswerFn = () => VNode;

export interface QaContent {
    question: string;
    answer: (string | AnswerFn)[];
}

export interface RewardCardProps {
    title?: string;
    label?: string;
    text?: string;
    description?: string;
    contentClass?: string;
    content?: Array<{
        key: string | number;
        label?: string;
        text: string[];
        description?: string;
        labelRender?: () => VNode;
        textRender?: () => VNode;
        descriptionRender?: () => VNode;
    }>;
}
