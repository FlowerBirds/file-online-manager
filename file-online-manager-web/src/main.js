import Vue from 'vue';
import ElementUI from 'element-ui';
import 'element-ui/lib/theme-chalk/index.css';
import axios from 'axios';
import VueAxios from 'vue-axios';
import uploader from 'vue-simple-uploader'
import App from './App.vue';

// 拦截请求，添加前缀，使用于nginx代理
axios.interceptors.request.use(
    config => {
        config.url = '/file-online-manager/' + config.url;
        return config;
    },
    error => {
        return Promise.reject(error);
    }
);

Vue.use(ElementUI);
Vue.use(VueAxios, axios);
Vue.use(uploader)

Vue.config.productionTip = false

new Vue({
    render: h => h(App),
}).$mount('#app')