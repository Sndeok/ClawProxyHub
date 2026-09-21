import { ref, type Ref } from 'vue'

// 服务端分页列表的通用状态与取数逻辑：
// 把 page/pageSize/total/loading 与「翻页后重新拉取」收敛到一处，
// 各页只关心「怎么拼查询参数」和「怎么塞数据」。
export interface PagedResult<T> {
  items: T[]
  total: number
}

export interface PagedListOptions<T> {
  pageSize?: number
  fetch: (params: { page: number; pageSize: number }) => Promise<PagedResult<T>>
}

export function usePagedList<T>(opts: PagedListOptions<T>) {
  const items: Ref<T[]> = ref([]) as Ref<T[]>
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(opts.pageSize ?? 20)
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      const r = await opts.fetch({ page: page.value, pageSize: pageSize.value })
      items.value = r.items
      total.value = r.total
    } finally {
      loading.value = false
    }
  }

  function onPageSizeChange() {
    page.value = 1
    return load()
  }

  return { items, total, page, pageSize, loading, load, onPageSizeChange }
}
