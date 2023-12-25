<template>
    <div id="app">
      <el-tabs v-model="activeTab" @tab-click="handleClick">
        <el-tab-pane label="文件" name="tab1">
          <HomePage></HomePage>
        </el-tab-pane>
        <el-tab-pane label="服务" name="tab2">
          <K8sServicePage ref="k8s"></K8sServicePage>
        </el-tab-pane>
      </el-tabs>
    </div>
</template>

<script>
import HomePage from './components/HomePage.vue';
import K8sServicePage from "@/components/K8sServicePage";

export default {
    name: 'App',
    data() {
        return {
          activeTab: 'tab1'
        }
    },
    components: {
        HomePage,
      K8sServicePage,
    },
  methods: {
    handleClick(tab, event) {
      if (tab.name == "tab2") {
        this.$refs.k8s.init()
      }
    }
  }
}
</script>

<style>
#app {
    font-family: Avenir, Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    /*text-align: center;*/
    color: #2c3e50;
    margin-top: 0px;
}
</style>