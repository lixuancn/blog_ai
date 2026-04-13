# Tasks

- [x] Task 1: 建立后端项目骨架（Go + 路由 + 中间件）
  - [x] SubTask 1.1: 初始化 Go 模块与基础目录结构（cmd, internal）
  - [x] SubTask 1.2: 接入路由与中间件（恢复、日志、鉴权占位）
  - [x] SubTask 1.3: 配置管理（读取固定管理员密钥）

- [x] Task 2: 实现后台登录与鉴权
  - [x] SubTask 2.1: 登录接口（校验 `lane` 与 `sha1(md5(\"******\"))`）
  - [x] SubTask 2.2: 登录态管理（cookie/session 或 token）
  - [x] SubTask 2.3: 鉴权中间件（保护后台接口）
  - [x] SubTask 2.4: 登出接口

- [x] Task 3: 分类管理接口（基于 `info_menu`）
  - [x] SubTask 3.1: 分类列表（支持名称搜索、父子关系展开）
  - [x] SubTask 3.2: 新增分类（站内/站外、url 校验）
  - [x] SubTask 3.3: 编辑分类（循环父子关系校验）
  - [x] SubTask 3.4: 删除分类（阻止删除存在子分类/文章的分类）

- [x] Task 4: 文章管理接口（基于 `info_article`）
  - [x] SubTask 4.1: 文章列表（标题搜索、分类筛选、推荐筛选、分页）
  - [x] SubTask 4.2: 新增文章（表单校验、入库）
  - [x] SubTask 4.3: 编辑文章（入库与字段更新）
  - [x] SubTask 4.4: 删除文章（联动标签统计）

- [x] Task 5: 标签管理与联动（基于 `info_tag`）
  - [x] SubTask 5.1: 标签列表（名称搜索、引用次数排序）
  - [x] SubTask 5.2: 保存/编辑文章时的标签解析与统计维护
  - [x] SubTask 5.3: 删除未引用标签接口（被引用则拒绝）
  - [x] SubTask 5.4: 提供标签全量重算服务方法（可选）

- [x] Task 6: 控制台统计接口
  - [x] SubTask 6.1: 总览统计（文章/分类/标签/推荐位数量）
  - [x] SubTask 6.2: 最近发布文章列表

- [x] Task 7: 后台前端骨架与页面
  - [x] SubTask 7.1: 后台布局（顶部/侧边栏/面包屑/主内容区）
  - [x] SubTask 7.2: 登录页
  - [x] SubTask 7.3: 控制台首页
  - [x] SubTask 7.4: 分类列表页 + 新增/编辑页
  - [x] SubTask 7.5: 文章列表页 + 新增/编辑页（带编辑器）
  - [x] SubTask 7.6: 标签列表页

- [x] Task 8: 验收与自测
  - [x] SubTask 8.1: 依据验收标准手动走查
  - [x] SubTask 8.2: 基于接口契约补充必要的单元或集成测试（聚焦服务层与数据校验）

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 1]
- [Task 4] depends on [Task 1]
- [Task 5] depends on [Task 4]
- [Task 6] depends on [Task 1, Task 4, Task 5, Task 3]
- [Task 7] depends on [Task 2, Task 3, Task 4, Task 6]
