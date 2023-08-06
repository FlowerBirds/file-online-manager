<template>
    <div id="app">
        <!-- <img alt="Vue logo" src="./assets/logo.png"> -->
        <HomePage></HomePage>
        <!-- <Test></Test> -->
        <!-- <LargeFileUpload></LargeFileUpload> -->
    </div>
</template>

<script>
    import HomePage from './components/HomePage.vue';
    // import Test from './components/test.vue';
    // import LargeFileUpload from "./components/LargeFileUpload.vue";

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
            // Test,
            // LargeFileUpload
        }
    }
</script>

<style>
    #app {
        font-family: Avenir, Helvetica, Arial, sans-serif;
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        text-align: center;
        color: #2c3e50;
        margin-top: 60px;
    }
</style>