/**
 * @file 年度总结常量定义
 */
import typeAgentImg from '@/assets/summary/type_agent.webp';
import typeBugImg from '@/assets/summary/type_bug.webp';
import typeFirstImg from '@/assets/summary/type_first.webp';
import typeSpeedImg from '@/assets/summary/type_speed.webp';
import exportAgentImg from '@/assets/summary/export_agent.webp';
import exportBugImg from '@/assets/summary/export_bug.webp';
import exportFirstImg from '@/assets/summary/export_first.webp';
import exportSpeedImg from '@/assets/summary/export_speed.webp';
import agentVideo from '@/assets/summary/video/agent.mp4';
import bugVideo from '@/assets/summary/video/bug.mp4';
import firstVideo from '@/assets/summary/video/first.mp4';
import speedVideo from '@/assets/summary/video/speed.mp4';

/**
 * 后台返回类型与项目内部类型的映射
 */
export const BACKEND_TYPE_MAP: Record<string, string> = {
    pioneer: 'first',
    high_freq: 'agent',
    problem_solver: 'bug',
    efficiency: 'speed',
};

/**
 * 结果类型图片映射
 */
export const TYPE_IMAGE_MAP: Record<string, string> = {
    agent: typeAgentImg,
    bug: typeBugImg,
    first: typeFirstImg,
    speed: typeSpeedImg,
};

/**
 * 结果导出类型图片映射
 */
export const EXPORT_IMAGE_MAP: Record<string, string> = {
    agent: exportAgentImg,
    bug: exportBugImg,
    first: exportFirstImg,
    speed: exportSpeedImg,
};

/**
 * 结果类型文案映射
 */
export const TYPE_TEXT_MAP: Record<string, string> = {
    bug: '报错、异常、奇怪现象,对你来说只是线索,Debug 对你而言不是痛苦,是解谜。',
    first: '当多数人还在观望时，你已把 AI 纳入生产力，后来成为标配的用法，对你只是顺手。',
    speed: '你很清楚哪些工作不值得从零写,重复与模板交给 AI,把时间留给真正需要判断的地方。',
    agent: '你不是一次性提问的人,而是通过多轮追问与调整,逐步把 AI 用到最优解。',
};

/**
 * 结果类型视频映射
 */
export const TYPE_VIDEO_MAP: Record<string, string> = {
    agent: agentVideo,
    bug: bugVideo,
    first: firstVideo,
    speed: speedVideo,
};

/**
 * 默认结果类型
 */
export const DEFAULT_RESULT_TYPE = 'speed';
