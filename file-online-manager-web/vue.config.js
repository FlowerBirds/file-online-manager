const {
    defineConfig
} = require('@vue/cli-service')
const { createProxyMiddleware } = require('http-proxy-middleware');
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin')
const proxy = createProxyMiddleware({
    target: 'http://localhost:8080',
    changeOrigin: true,
    router: function (req) {
        if (req.headers['x-custom-header'] === 'special') {
            return 'http://localhost:4000';  // 如果请求头中的 x-custom-header 是 'special'，则代理到 http://localhost:4000
        } else {
            return 'http://localhost:5000';  // 否则代理到 http://localhost:5000
        }
    },
});

module.exports = defineConfig({
    publicPath: "/fm",
    // 产品源码映射，true前端可以看到源码
    productionSourceMap: true,
    // 将es6转化为es5
    transpileDependencies: true,
    // 开发时的代理
    devServer: {
        port: 8081,
        proxy: {
            '/api': {
                // target: 'http://172.29.190.147:30001/',
                target: "http://localhost:8080/",
                changeOrigin: true
            }
        }
    },
    // 编译后输出路径
    outputDir: '../static',
    configureWebpack: config => {
        config.devtool = 'source-map';
        config.plugins.push(new MonacoWebpackPlugin())
    }
})