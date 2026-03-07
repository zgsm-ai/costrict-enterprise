/**
 * @file index.ts
 * @description 自定义指令统一导出文件
 */
import hoverDirective from './hover';

// 导出所有指令
export { hoverDirective };

// 默认导出所有指令的对象，方便注册
export default {
    hover: hoverDirective,
};
