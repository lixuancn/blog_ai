# LaneBlog 前台站点 Spec

## Why
LaneBlog 已完成后台管理系统与前台技术方案，但尚未形成可直接驱动实现的前台规格说明。需要将首页、文章详情页和前台渲染方式等内容沉淀为结构化 spec，作为后续开发与验收依据。

## What Changes
- 新增 LaneBlog 前台站点规格，覆盖首页、文章详情页、顶部导航、推荐区、点赞/点踩、点击计数
- 明确前台采用 Go + SSR + `html/template` 的实现方式
- 明确首页文章列表、全站推荐、类目推荐、正文截断和排序规则
- 明确文章详情页完整展示、点击自增、点赞/点踩交互
- 预留分类页与站内/站外导航跳转能力

## Impact
- Affected specs: 前台路由、前台 SSR 渲染、首页内容聚合、文章详情展示、推荐逻辑、SEO
- Affected code: 前台 handler、前台 service、前台模板、前台样式与脚本、文章查询逻辑、导航树查询逻辑

## ADDED Requirements

### Requirement: 前台首页渲染
系统 SHALL 提供前台首页，并使用服务端渲染输出完整 HTML。

#### Scenario: 打开首页
- **WHEN** 访客访问 `/`
- **THEN** 系统返回完整首页 HTML
- **THEN** 页面包含顶部导航、站点标题区、文章列表区、全站推荐区

#### Scenario: 首页文章按时间倒序展示
- **WHEN** 系统渲染首页文章列表
- **THEN** 文章按 `ctime` 倒序排列
- **THEN** 每篇文章展示标题、作者、发布时间、分类、标签、摘要和正文前 100 个字

#### Scenario: 首页文章列表支持分页
- **WHEN** 访客访问首页并携带 `page` 查询参数
- **THEN** 系统按页码返回对应页的文章列表
- **THEN** 页面展示上一页、下一页和当前页信息
- **THEN** 默认每页展示 10 篇文章

#### Scenario: 首页展示全站推荐
- **WHEN** 系统渲染首页右侧推荐区
- **THEN** 只展示 `recommend_type=1` 的文章
- **THEN** 推荐区只显示文章标题与跳转链接

### Requirement: 前台顶部导航
系统 SHALL 基于 `info_menu` 渲染支持层级关系的顶部导航。

#### Scenario: 渲染两级导航
- **WHEN** 系统读取 `info_menu`
- **THEN** `pid=0` 的记录作为顶级导航
- **THEN** `pid>0` 的记录作为下级导航项

#### Scenario: 点击站内导航
- **WHEN** 访客点击 `in_out=1` 的导航项
- **THEN** 系统跳转到站内页面或预留分类页

#### Scenario: 点击站外导航
- **WHEN** 访客点击 `in_out=2` 的导航项
- **THEN** 系统跳转到该项 `url`

### Requirement: 站点标题区
系统 SHALL 在导航栏下方展示站点主标题与副标题。

#### Scenario: 首页展示标题区
- **WHEN** 访客访问首页
- **THEN** 页面展示主标题 `LaneBlog`
- **THEN** 页面展示副标题 `每一个没有起舞的日子都是在辜负生命。`

### Requirement: 文章详情页渲染
系统 SHALL 提供文章详情页，并展示文章完整内容与附属信息。

#### Scenario: 打开文章详情页
- **WHEN** 访客访问 `/article/{id}`
- **THEN** 系统返回完整文章详情 HTML
- **THEN** 页面展示标题、作者、发布时间、分类、标签、正文、点击数、点赞数、点踩数

#### Scenario: 文章详情页展示类目推荐
- **WHEN** 系统渲染文章详情页右侧推荐区
- **THEN** 优先展示当前分类下的推荐文章
- **THEN** 过滤当前文章自身
- **THEN** 推荐区只展示标题与跳转链接

### Requirement: 点击计数
系统 SHALL 在文章详情页访问时自动增加文章点击数。

#### Scenario: 进入文章详情页
- **WHEN** 访客成功访问某篇文章详情页
- **THEN** 系统将该文章 `clicks` 加 1
- **THEN** 页面展示更新后的点击数或保证后续读取结果已累加

### Requirement: 点赞与点踩
系统 SHALL 提供文章点赞和点踩交互能力。

#### Scenario: 点赞文章
- **WHEN** 访客点击“赞一下”
- **THEN** 系统将该文章 `good_num` 加 1
- **THEN** 页面显示最新点赞数

#### Scenario: 点踩文章
- **WHEN** 访客点击“踩一下”
- **THEN** 系统将该文章 `bad_num` 加 1
- **THEN** 页面显示最新点踩数

### Requirement: 前台 SEO
系统 SHALL 为首页和文章详情页输出基础 SEO 信息。

#### Scenario: 首页 SEO 输出
- **WHEN** 系统渲染首页
- **THEN** 页面输出首页 `title`、`description`、`keywords`

#### Scenario: 文章详情 SEO 输出
- **WHEN** 系统渲染文章详情页
- **THEN** 优先使用文章的 `seo_title`、`seo_description`、`seo_keywords`
- **THEN** 若 SEO 字段为空，则使用标题、摘要或标签信息降级填充

## MODIFIED Requirements

### Requirement: 无
当前变更为新增前台规格，不修改已有后台 spec 要求。

## REMOVED Requirements

### Requirement: 无
**Reason**: 当前变更未移除任何既有规格要求。
**Migration**: 无需迁移。
