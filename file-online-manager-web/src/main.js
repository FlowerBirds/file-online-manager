import Vue from 'vue';
import ElementUI from 'element-ui';
import 'element-ui/lib/theme-chalk/index.css';
import axios from 'axios';
import VueAxios from 'vue-axios';
import uploader from 'vue-simple-uploader'
import App from './App.vue';

Vue.use(ElementUI);
Vue.use(VueAxios, axios);
Vue.use(uploader)

Vue.config.productionTip = false

new Vue({
    render: h => h(App),
}).$mount('#app')