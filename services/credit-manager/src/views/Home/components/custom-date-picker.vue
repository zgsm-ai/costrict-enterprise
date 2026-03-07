<template>
    <div class="custom-date-picker">
        <!-- 自定义日期显示面板 -->
        <div
            class="custom-date-display"
            :class="{ 'select-time__group-item': true, active: isActive }"
            @click="showDatePicker"
        >
            <span class="date-text">{{ displayText }}</span>
            <img
                src="../../../assets/icon/calendar.svg"
                alt=""
            />
        </div>

        <!-- 隐藏的 n-date-picker，只用于弹出面板 -->
        <n-date-picker
            ref="datePickerRef"
            v-model:formatted-value="dateRange"
            type="daterange"
            value-format="yyyy-MM-dd"
            :show="isDatePickerVisible"
            @update:show="handleDatePickerVisibilityChange"
            @update:formatted-value="handleDateChange"
        />
    </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { NDatePicker } from 'naive-ui';
import { useI18n } from 'vue-i18n';

// 定义 props
interface Props {
    modelValue?: [string, string] | null;
    isSelected?: boolean; // 添加是否选中的状态
}

// 定义 emits
interface Emits {
    (e: 'update:modelValue', value: [string, string] | null): void;
}

const { t } = useI18n();

const props = withDefaults(defineProps<Props>(), {
    modelValue: null,
    isSelected: false,
});

const emit = defineEmits<Emits>();

// 内部状态
const datePickerRef = ref();
const dateRange = ref<[string, string] | null>(props.modelValue);
const isDatePickerVisible = ref(false);

// 计算是否处于激活状态（有选中值且父组件表示当前选中）
const isActive = computed(() => {
    return props.isSelected && dateRange.value && dateRange.value.length === 2;
});

// 计算显示文本
const displayText = computed(() => {
    if (!dateRange.value || dateRange.value.length !== 2) {
        return t('components.datePicker.displayText');
    }
    const [startDate, endDate] = dateRange.value;
    return `${startDate} ${t('components.datePicker.displayTextUnit')} ${endDate}`;
});

// 显示日期选择器面板
const showDatePicker = () => {
    isDatePickerVisible.value = true;
};

// 处理日期选择器面板的可见性变化
const handleDatePickerVisibilityChange = (visible: boolean) => {
    isDatePickerVisible.value = visible;
};

// 处理日期变化
const handleDateChange = (value: [string, string] | null) => {
    dateRange.value = value;
    emit('update:modelValue', value);
};

// 监听外部 modelValue 变化
watch(
    () => props.modelValue,
    (newValue: [string, string] | null) => {
        dateRange.value = newValue;
    },
);
</script>

<style scoped lang="less">
.custom-date-picker {
    position: relative;

    .custom-date-display {
        display: flex;
        align-items: center;
        padding: 0 8px;
        border-radius: 4px;
        cursor: pointer;
        transition: all 0.3s ease;
        height: 28px;
        font-size: 12px;
        margin-right: 1px;
        opacity: 0.7;
        background-color: rgba(255, 255, 255, 0.15);

        &:hover {
            opacity: 1;
            background-color: rgba(255, 255, 255, 0.1);
        }

        &.active {
            opacity: 1;
            background-color: rgba(255, 255, 255, 0.2);
        }

        .date-text {
            color: rgba(255, 255, 255, 0.9);
            margin-right: 6px;
            white-space: nowrap;
        }

        .calendar-icon {
            font-size: 14px;
            flex-shrink: 0;
        }
    }
}

:deep(.n-date-picker) {
    .n-input {
        position: absolute;
        opacity: 0;
        pointer-events: none;
        z-index: -1;
        width: 0;
        height: 0;
    }
}
</style>
