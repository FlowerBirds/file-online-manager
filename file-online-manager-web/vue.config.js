const {
    defineConfig
} = require('@vue/cli-service')
module.exports = defineConfig({
    publicPath: "/",
    productionSourceMap: true,
    transpileDependencies: true,
    devServer: {
        proxy: 'http://localhost:8080'
    },
    outputDir: '../static'
})