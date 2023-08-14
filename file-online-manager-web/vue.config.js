const {
    defineConfig
} = require('@vue/cli-service')
module.exports = defineConfig({
    publicPath: "/",
    // 产品源码映射，true前端可以看到源码
    productionSourceMap: true,
    // 将es6转化为es5
    transpileDependencies: true,
    // 开发时的代理
    devServer: {
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true
            }
        }
    },
    // 编译后输出路径
    outputDir: '../static'
})