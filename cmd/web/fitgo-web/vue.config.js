const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  devServer: {
    port: 9093,
    proxy: {
      '/coros': {
        target: 'http://localhost:9092', // 后端API地址
        changeOrigin: true,
        pathRewrite: {
          '^/coros': '/coros' // 保持路径不变
        }
      }
    }
  },
  chainWebpack: config => {
    config.module
      .rule('md')
      .test(/\.md$/)
      .use('raw-loader')
      .loader('raw-loader')
      .end()
  }
})
