<template>
    <div id="app">
        <HomePage></HomePage>
    </div>
</template>

<script>
import HomePage from './components/HomePage.vue';

export default {
    name: 'App',
    data() {
        return {
            msg: {
                name: "Welcome to Your Vue.js App"
            }
        }
    },
    components: {
        HomePage,
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