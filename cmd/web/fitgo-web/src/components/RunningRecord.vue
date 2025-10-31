<template>
  <n-data-table
    :columns="columns"
    :data="data"
    :pagination="pagination"
    :loading="loading"
    :bordered="false"
  />

  <n-drawer
    v-model:show="drawerVisible"
    :default-width="800"
    :placement="drawerPlacement"
    resizable
  >
    <n-drawer-content :title="drawerTitle">
      <div v-if="analysisLoading" class="loading-container">
        <n-spin size="large" />
        <p>正在加载分析结果...</p>
      </div>
      <div v-else-if="analysisError" class="error-message">
        <n-alert type="error" :title="analysisError" />
      </div>
      <div class="markdown-body" v-html="mdContent">
      </div>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup>
import { ref, onMounted, h, computed } from 'vue'
import { NButton, NIcon, useMessage, NSpin, NAlert } from 'naive-ui'
import MarkdownIt from 'markdown-it'

const message = useMessage()
const data = ref([])
const loading = ref(false)
const md = new MarkdownIt()

// 抽屉相关状态
const drawerVisible = ref(false)
const drawerPlacement = ref('right')
const analysisLoading = ref(false)
const analysisError = ref('')
const analysisResult = ref('')
const currentRow = ref(null)

// 计算 markdown 内容
const mdContent = computed(() => {
  return analysisResult.value ? md.render(analysisResult.value) : ''
})

// 计算抽屉标题
const drawerTitle = computed(() => {
  return currentRow.value ? `${currentRow.value.name || '运动记录'} 分析` : '运动分析'
})

// 先定义函数
// 处理分页变化
const handlePageChange = (page) => {
  pagination.value.page = page
  fetchData()
}

// 处理每页条数变化
const handlePageSizeChange = (pageSize) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  fetchData()
}

// 然后再定义分页配置
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  onUpdatePage: handlePageChange,        // 现在函数已经定义了
  onUpdatePageSize: handlePageSizeChange // 现在函数已经定义了
})

// 创建列
const columns = [
  {
    title: "日期",
    key: "date"
  },
  {
    title: "名称",
    key: "name"
  },
  {
    title: "图标",
    key: "imageUrl",
    render(row) {
      // 如果 imageUrl 是 SVG 字符串
      if (row.imageUrl && row.imageUrl.includes('<svg')) {
        return h('div', { 
          innerHTML: row.imageUrl,
          style: 'display: flex; align-items: center; justify-content: center;'
        })
      }
      // 如果 imageUrl 是图片 URL
      else if (row.imageUrl) {
        return h('img', {
          src: row.imageUrl,
          style: 'width: 24px; height: 24px; object-fit: contain;'
        })
      }
      // 如果没有图标数据，显示默认的加号图标
      return h(NIcon, { size: 24 }, () => 
        h('svg', { 
          xmlns: 'http://www.w3.org/2000/svg', 
          viewBox: '0 0 512 512',
          style: 'width: 24px; height: 24px;'
        }, [
          h('path', { 
            d: 'M368.5 240H272v-96.5c0-8.8-7.2-16-16-16s-16 7.2-16 16V240h-96.5c-8.8 0-16 7.2-16 16 0 4.4 1.8 8.4 4.7 11.3 2.9 2.9 6.9 4.7 11.3 4.7H240v96.5c0 4.4 1.8 8.4 4.7 11.3 2.9 2.9 6.9 4.7 11.3 4.7 8.8 0 16-7.2 16-16V272h96.5c8.8 0 16-7.2 16-16s-7.2-16-16-16z',
            fill: 'currentColor'
          })
        ])
      )
    }
  },
  {
    title: "操作",
    key: "actions",
    render(row) {
      return h(
        NButton,
        {
          strong: true,
          tertiary: true,
          size: "small",
          onClick: () => showAnalysisDrawer(row)
        },
        { default: () => "分析" }
      )
    }
  }
]

// 显示分析抽屉
const showAnalysisDrawer = (row) => {
  currentRow.value = row
  drawerVisible.value = true
  fetchAnalysisData(row)
}

// 获取分析数据
const fetchAnalysisData = async (row) => {
  analysisLoading.value = true
  analysisError.value = ''
  
  try {
    const response = await fetch(
      `http://localhost:9092/coros/ai/summary?labelId=${row.labelId}&sportType=${row.sportType}`
    )
    
    if (!response.ok) {
      throw new Error('获取分析结果失败')
    }
    
    analysisResult.value = await response.text()
  } catch (error) {
    console.error('获取分析结果出错:', error)
    analysisError.value = error.message || '获取分析结果失败，请稍后重试'
  } finally {
    analysisLoading.value = false
  }
}

// 获取数据
const fetchData = async () => {
  loading.value = true
  try {
    const { page, pageSize } = pagination.value
    const response = await fetch(
      `http://localhost:9092/coros/active?size=${pageSize}&pageNumber=${page}`
    )

    if (!response.ok) {
      throw new Error('获取数据失败')
    }

    const result = await response.json()

    if (result && result.data) {
      data.value = result.data.dataList || []
      pagination.value.itemCount = result.data.totalCount || data.value.length
    } else {
      data.value = []
      pagination.value.itemCount = 0
    }
  } catch (error) {
    message.error(error.message || '获取数据失败')
    console.error('Error fetching data:', error)
    data.value = []
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取第一页数据
onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  min-height: 200px;
}

.error-message {
  margin: 10px 0;
}

.analysis-content {
  padding: 10px 0;
  white-space: pre-wrap;
  line-height: 1.6;
}

.analysis-content :deep(h3) {
  margin: 20px 0 10px 0;
  color: #18a058;
}

.analysis-content :deep(ul) {
  padding-left: 20px;
  margin: 10px 0;
}

.analysis-content :deep(li) {
  margin-bottom: 5px;
}
</style>

<style scoped>
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  min-height: 200px;
}

.error-message {
  margin: 10px 0;
}

.analysis-content {
  padding: 10px 0;
  white-space: pre-wrap;
  line-height: 1.6;
}

.analysis-content :deep(h3) {
  margin: 20px 0 10px 0;
  color: #18a058;
}

.analysis-content :deep(ul) {
  padding-left: 20px;
  margin: 10px 0;
}

.analysis-content :deep(li) {
  margin-bottom: 5px;
}
</style>