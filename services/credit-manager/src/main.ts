import './assets/main.css';
import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import i18n from './locales';
import { setupAuthGuard } from './router/guards/auth';
import { createHead } from '@vueuse/head';
// 导入自定义指令
import directives from './directives';

const app = createApp(App);
const head = createHead();

app.use(createPinia());
app.use(router);
app.use(i18n);
app.use(head);

// 注册自定义指令
Object.entries(directives).forEach(([name, directive]) => {
    app.directive(name, directive);
});

// 在Pinia初始化后设置路由守卫
setupAuthGuard(router);

app.mount('#app');
