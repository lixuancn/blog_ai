# Tasks

- [x] Task 1: 建立前台页面骨架与路由
  - [x] SubTask 1.1: 新增前台首页与文章详情页路由
  - [x] SubTask 1.2: 建立前台模板布局、头部导航、主内容区、侧栏区
  - [x] SubTask 1.3: 建立前台静态资源目录与基础样式、脚本

- [x] Task 2: 实现前台数据服务
  - [x] SubTask 2.1: 实现导航树查询与两级菜单组装
  - [x] SubTask 2.2: 实现首页文章列表查询与正文前 100 字预览处理
  - [x] SubTask 2.3: 实现首页全站推荐文章查询
  - [x] SubTask 2.4: 实现文章详情查询与类目推荐查询

- [x] Task 3: 实现首页 SSR 页面
  - [x] SubTask 3.1: 渲染站点标题 `LaneBlog`
  - [x] SubTask 3.2: 渲染副标题 `每一个没有起舞的日子都是在辜负生命。`
  - [x] SubTask 3.3: 渲染文章列表双栏布局
  - [x] SubTask 3.4: 渲染全站推荐侧栏
  - [x] SubTask 3.5: 首页文章列表支持分页

- [x] Task 4: 实现文章详情页 SSR 页面
  - [x] SubTask 4.1: 渲染文章完整信息
  - [x] SubTask 4.2: 页面访问时累加点击数
  - [x] SubTask 4.3: 渲染同分类推荐文章侧栏
  - [x] SubTask 4.4: 渲染点赞与点踩按钮

- [x] Task 5: 实现前台交互接口
  - [x] SubTask 5.1: 实现文章点赞接口
  - [x] SubTask 5.2: 实现文章点踩接口
  - [x] SubTask 5.3: 前端按钮调用接口并刷新数值

- [x] Task 6: 实现 SEO 与页面细节
  - [x] SubTask 6.1: 首页输出基础 SEO 信息
  - [x] SubTask 6.2: 文章详情页输出文章 SEO 信息
  - [x] SubTask 6.3: 响应式处理导航与侧栏布局

- [x] Task 7: 验收与测试
  - [x] SubTask 7.1: 补充首页与详情页的 handler/service 测试
  - [x] SubTask 7.2: 验证文章排序、正文截断、点击自增、推荐逻辑
  - [x] SubTask 7.3: 验证点赞与点踩接口行为

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 1, Task 2]
- [Task 4] depends on [Task 1, Task 2]
- [Task 5] depends on [Task 4]
- [Task 6] depends on [Task 3, Task 4]
- [Task 7] depends on [Task 3, Task 4, Task 5, Task 6]
