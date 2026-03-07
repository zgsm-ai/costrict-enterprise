<template>
    <div class="operation-card relative">
        <img
            src="../../assets/card.png"
            alt=""
        />
        <span class="absolute top-[50%] transform translate-y-[-50%] left-4 font-medium">{{
            title
        }}</span>
    </div>
</template>

<script setup lang="ts">
withDefaults(
    defineProps<{
        title?: string;
    }>(),
    {
        title: '',
    },
);
</script>
