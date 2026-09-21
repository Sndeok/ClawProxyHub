import { createApp } from 'vue'
import { createPinia } from 'pinia'
// 按需注册：只引入实际用到的 43 个组件，避免 TDesign 全量打包（首屏 JS 从 1.5MB 降到 1/3 左右）。
// 新增组件时记得在这里补一行（忘了会报「未知组件」而不是静默失效）。
import {
  Alert, Aside, Avatar, Button, Card, Checkbox, CheckboxGroup, Col, Content,
  DatePicker, DateRangePicker, Descriptions, DescriptionsItem, Dialog, Drawer,
  Empty, Form, FormItem, Header, Input, InputNumber, Layout, Link, Loading,
  Menu, MenuItem, Option, Pagination, Popconfirm, Popup, RadioButton, RadioGroup,
  Row, Select, Space, Switch, TabPanel, Table, Tabs, Tag, Textarea, Tooltip, Upload,
} from 'tdesign-vue-next'
import App from './App.vue'
import i18n from './i18n'
import router from './router'
import 'tdesign-vue-next/es/style/index.css'
import './assets/theme.css'

const app = createApp(App)
app.use(createPinia()).use(router).use(i18n)

for (const c of [
  Alert, Aside, Avatar, Button, Card, Checkbox, CheckboxGroup, Col, Content,
  DatePicker, DateRangePicker, Descriptions, DescriptionsItem, Dialog, Drawer,
  Empty, Form, FormItem, Header, Input, InputNumber, Layout, Link, Loading,
  Menu, MenuItem, Option, Pagination, Popconfirm, Popup, RadioButton, RadioGroup,
  Row, Select, Space, Switch, TabPanel, Table, Tabs, Tag, Textarea, Tooltip, Upload,
]) {
  app.use(c as never)
}

app.mount('#app')