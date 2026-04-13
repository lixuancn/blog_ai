package handler

import (
	"html/template"
	"net/http"
	"strconv"

	"example.com/laneblog/internal/config"
	"example.com/laneblog/internal/http/middleware"
)

type AdminUIHandler struct {
	cfg  config.Config
	tmpl *template.Template
}

type crumb struct {
	Name string
	Path string
}

type adminPageData struct {
	Title       string
	Heading     string
	Username    string
	ActiveNav   string
	PageID      string
	RecordID    int64
	Redirect    string
	Content     template.HTML
	IsLoginPage bool
	Breadcrumbs []crumb
}

func NewAdminUIHandler(cfg config.Config) *AdminUIHandler {
	return &AdminUIHandler{
		cfg:  cfg,
		tmpl: template.Must(template.New("admin-ui").Parse(adminPageTemplate)),
	}
}

func (h *AdminUIHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if middleware.Authorized(h.cfg, r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	h.render(w, adminPageData{
		Title:       "登录",
		Heading:     "LaneBlog Admin",
		PageID:      "login",
		Redirect:    r.URL.Query().Get("redirect"),
		Content:     template.HTML(loginMarkup),
		IsLoginPage: true,
	})
}

func (h *AdminUIHandler) DashboardPage(w http.ResponseWriter, _ *http.Request) {
	h.render(w, adminPageData{
		Title:     "控制台",
		Heading:   "控制台",
		Username:  h.cfg.Auth.Username,
		ActiveNav: "dashboard",
		PageID:    "dashboard",
		Content:   template.HTML(dashboardMarkup),
		Breadcrumbs: []crumb{
			{Name: "控制台"},
		},
	})
}

func (h *AdminUIHandler) CategoryListPage(w http.ResponseWriter, _ *http.Request) {
	h.render(w, adminPageData{
		Title:     "分类管理",
		Heading:   "分类管理",
		Username:  h.cfg.Auth.Username,
		ActiveNav: "category",
		PageID:    "category-list",
		Content:   template.HTML(categoryListMarkup),
		Breadcrumbs: []crumb{
			{Name: "控制台", Path: "/admin"},
			{Name: "分类管理"},
		},
	})
}

func (h *AdminUIHandler) CategoryCreatePage(w http.ResponseWriter, _ *http.Request) {
	h.render(w, adminPageData{
		Title:     "新增分类",
		Heading:   "新增分类",
		Username:  h.cfg.Auth.Username,
		ActiveNav: "category",
		PageID:    "category-form",
		Content:   template.HTML(categoryFormMarkup),
		Breadcrumbs: []crumb{
			{Name: "控制台", Path: "/admin"},
			{Name: "分类管理", Path: "/admin/category"},
			{Name: "新增分类"},
		},
	})
}

func (h *AdminUIHandler) CategoryEditPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.render(w, adminPageData{
		Title:     "编辑分类",
		Heading:   "编辑分类",
		Username:  h.cfg.Auth.Username,
		ActiveNav: "category",
		PageID:    "category-form",
		RecordID:  id,
		Content:   template.HTML(categoryFormMarkup),
		Breadcrumbs: []crumb{
			{Name: "控制台", Path: "/admin"},
			{Name: "分类管理", Path: "/admin/category"},
			{Name: "编辑分类"},
		},
	})
}

func (h *AdminUIHandler) ArticleListPage(w http.ResponseWriter, _ *http.Request) {
	h.render(w, adminPageData{
		Title:     "文章管理",
		Heading:   "文章管理",
		Username:  h.cfg.Auth.Username,
		ActiveNav: "article",
		PageID:    "article-list",
		Content:   template.HTML(articleListMarkup),
		Breadcrumbs: []crumb{
			{Name: "控制台", Path: "/admin"},
			{Name: "文章管理"},
		},
	})
}

func (h *AdminUIHandler) ArticleCreatePage(w http.ResponseWriter, _ *http.Request) {
	h.render(w, adminPageData{
		Title:     "新建文章",
		Heading:   "新建文章",
		Username:  h.cfg.Auth.Username,
		ActiveNav: "article",
		PageID:    "article-form",
		Content:   template.HTML(articleFormMarkup),
		Breadcrumbs: []crumb{
			{Name: "控制台", Path: "/admin"},
			{Name: "文章管理", Path: "/admin/article"},
			{Name: "新建文章"},
		},
	})
}

func (h *AdminUIHandler) ArticleEditPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.render(w, adminPageData{
		Title:     "编辑文章",
		Heading:   "编辑文章",
		Username:  h.cfg.Auth.Username,
		ActiveNav: "article",
		PageID:    "article-form",
		RecordID:  id,
		Content:   template.HTML(articleFormMarkup),
		Breadcrumbs: []crumb{
			{Name: "控制台", Path: "/admin"},
			{Name: "文章管理", Path: "/admin/article"},
			{Name: "编辑文章"},
		},
	})
}

func (h *AdminUIHandler) TagListPage(w http.ResponseWriter, _ *http.Request) {
	h.render(w, adminPageData{
		Title:     "标签管理",
		Heading:   "标签管理",
		Username:  h.cfg.Auth.Username,
		ActiveNav: "tag",
		PageID:    "tag-list",
		Content:   template.HTML(tagListMarkup),
		Breadcrumbs: []crumb{
			{Name: "控制台", Path: "/admin"},
			{Name: "标签管理"},
		},
	})
}

func (h *AdminUIHandler) render(w http.ResponseWriter, data adminPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.Execute(w, data); err != nil {
		http.Error(w, "render page failed", http.StatusInternalServerError)
	}
}

const adminPageTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}} - LaneBlog Admin</title>
  <link rel="stylesheet" href="/admin/assets/app.css">
</head>
<body class="{{if .IsLoginPage}}login-page{{else}}admin-page{{end}}" data-page="{{.PageID}}" data-record-id="{{.RecordID}}" data-redirect="{{.Redirect}}">
  {{if .IsLoginPage}}
  <main class="login-shell">
    <section class="login-card">
      <h1>{{.Heading}}</h1>
      <p class="page-subtitle">固定管理员登录入口</p>
      <div id="page-banner" class="banner" hidden></div>
      {{.Content}}
    </section>
  </main>
  {{else}}
  <div class="admin-shell">
    <aside class="sidebar">
      <a class="brand" href="/admin">LaneBlog Admin</a>
      <nav class="nav-list">
        <a class="nav-item {{if eq .ActiveNav "dashboard"}}active{{end}}" href="/admin">控制台</a>
        <a class="nav-item {{if eq .ActiveNav "category"}}active{{end}}" href="/admin/category">分类管理</a>
        <a class="nav-item {{if eq .ActiveNav "article"}}active{{end}}" href="/admin/article">文章管理</a>
        <a class="nav-item {{if eq .ActiveNav "tag"}}active{{end}}" href="/admin/tag">标签管理</a>
      </nav>
      <button id="logout-button" class="sidebar-logout" type="button">退出登录</button>
    </aside>
    <div class="main-shell">
      <header class="topbar">
        <div>
          <div class="page-title">{{.Heading}}</div>
          <div class="page-subtitle">当前账号：{{.Username}}</div>
        </div>
      </header>
      <main class="content-shell">
        <nav class="breadcrumbs">
          {{range $index, $item := .Breadcrumbs}}
            {{if gt $index 0}}<span class="crumb-separator">/</span>{{end}}
            {{if $item.Path}}<a href="{{$item.Path}}">{{$item.Name}}</a>{{else}}<span>{{$item.Name}}</span>{{end}}
          {{end}}
        </nav>
        <div id="page-banner" class="banner" hidden></div>
        {{.Content}}
      </main>
    </div>
  </div>
  {{end}}
  <script src="/admin/assets/app.js"></script>
</body>
</html>
`

const loginMarkup = `
<form id="login-form" class="panel-form">
  <label class="form-field">
    <span>用户名</span>
    <input id="login-username" name="username" type="text" autocomplete="username" placeholder="请输入用户名" required>
  </label>
  <label class="form-field">
    <span>密码</span>
    <input id="login-password" name="password" type="password" autocomplete="current-password" placeholder="请输入密码" required>
  </label>
  <button class="primary-button full-width" type="submit">登录</button>
</form>
`

const dashboardMarkup = `
<section class="card-grid" id="dashboard-summary"></section>
<section class="panel">
  <div class="panel-header">
    <h2>最近发布文章</h2>
  </div>
  <div id="recent-articles"></div>
</section>
<section class="panel">
  <div class="panel-header">
    <h2>快捷入口</h2>
  </div>
  <div id="quick-links" class="quick-links"></div>
</section>
`

const categoryListMarkup = `
<section class="panel">
  <form id="category-search-form" class="toolbar">
    <input name="name" type="text" placeholder="按分类名称搜索">
    <button class="primary-button" type="submit">搜索</button>
    <a class="secondary-button" href="/admin/category/create">新增分类</a>
  </form>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>分类名</th>
          <th>父级分类</th>
          <th>类型</th>
          <th>链接地址</th>
          <th>SEO 标题</th>
          <th>Item</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody id="category-table-body"></tbody>
    </table>
  </div>
</section>
`

const categoryFormMarkup = `
<section class="panel">
  <form id="category-form" class="panel-form">
    <div class="form-grid two-columns">
      <label class="form-field">
        <span>分类名称</span>
        <input name="name" type="text" maxlength="50" required>
        <small data-error-for="name"></small>
      </label>
      <label class="form-field">
        <span>父级分类</span>
        <select name="pid" id="category-parent"></select>
        <small data-error-for="pid"></small>
      </label>
      <label class="form-field">
        <span>链接类型</span>
        <select name="in_out" id="category-link-type">
          <option value="1">站内</option>
          <option value="2">站外</option>
        </select>
        <small data-error-for="in_out"></small>
      </label>
      <label class="form-field" id="category-url-field">
        <span>站外链接</span>
        <input name="url" type="url" placeholder="https://example.com">
        <small data-error-for="url"></small>
      </label>
      <label class="form-field">
        <span>SEO 标题</span>
        <input name="seo_title" type="text" maxlength="100" required>
        <small data-error-for="seo_title"></small>
      </label>
      <label class="form-field">
        <span>分类标识</span>
        <input name="item" type="text" maxlength="20" required>
        <small data-error-for="item"></small>
      </label>
    </div>
    <label class="form-field">
      <span>SEO 描述</span>
      <textarea name="seo_description" rows="3" maxlength="500" required></textarea>
      <small data-error-for="seo_description"></small>
    </label>
    <label class="form-field">
      <span>SEO 关键词</span>
      <input name="seo_keywords" type="text" maxlength="200" required>
      <small data-error-for="seo_keywords"></small>
    </label>
    <div class="form-actions">
      <button class="primary-button" type="submit" data-submit-mode="stay">保存</button>
      <button class="secondary-button" type="submit" data-submit-mode="back">保存并返回</button>
      <a class="text-button" href="/admin/category">取消</a>
      <button id="category-delete-button" class="danger-button hidden" type="button">删除分类</button>
    </div>
  </form>
</section>
`

const articleListMarkup = `
<section class="panel">
  <form id="article-search-form" class="toolbar toolbar-wrap">
    <input name="title" type="text" placeholder="按标题搜索">
    <select name="mid" id="article-filter-category"></select>
    <select name="recommend_type">
      <option value="">全部推荐</option>
      <option value="1">全站推荐</option>
      <option value="2">首页推荐</option>
    </select>
    <button class="primary-button" type="submit">搜索</button>
    <a class="secondary-button" href="/admin/article/create">新建文章</a>
  </form>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>标题</th>
          <th>分类</th>
          <th>作者</th>
          <th>标签</th>
          <th>点击</th>
          <th>点赞</th>
          <th>点踩</th>
          <th>推荐</th>
          <th>发布时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody id="article-table-body"></tbody>
    </table>
  </div>
  <div id="article-pagination" class="pagination"></div>
</section>
`

const articleFormMarkup = `
<section class="panel">
  <form id="article-form" class="panel-form">
    <div class="form-grid two-columns">
      <label class="form-field">
        <span>标题</span>
        <input name="title" type="text" maxlength="100" required>
        <small data-error-for="title"></small>
      </label>
      <label class="form-field">
        <span>所属分类</span>
        <select name="mid" id="article-category-select"></select>
        <small data-error-for="mid"></small>
      </label>
      <label class="form-field">
        <span>作者</span>
        <input name="author" type="text" maxlength="50" required>
        <small data-error-for="author"></small>
      </label>
      <label class="form-field">
        <span>标签</span>
        <input name="tag" type="text" placeholder="Go,MySQL,后台管理" required>
        <small data-error-for="tag"></small>
      </label>
      <label class="form-field">
        <span>SEO 标题</span>
        <input name="seo_title" type="text" maxlength="100" required>
        <small data-error-for="seo_title"></small>
      </label>
      <label class="form-field">
        <span>SEO 关键词</span>
        <input name="seo_keywords" type="text" maxlength="200" required>
        <small data-error-for="seo_keywords"></small>
      </label>
      <label class="form-field">
        <span>点击数</span>
        <input name="clicks" type="number" min="0" value="0">
        <small data-error-for="clicks"></small>
      </label>
      <label class="form-field">
        <span>推荐类型</span>
        <select name="recommend_type">
          <option value="1">全站推荐</option>
          <option value="2">首页推荐</option>
        </select>
        <small data-error-for="recommend_type"></small>
      </label>
      <label class="form-field">
        <span>点赞数</span>
        <input name="good_num" type="number" min="0" value="0">
        <small data-error-for="good_num"></small>
      </label>
      <label class="form-field">
        <span>点踩数</span>
        <input name="bad_num" type="number" min="0" value="0">
        <small data-error-for="bad_num"></small>
      </label>
      <label class="form-field">
        <span>发布时间（Unix 时间戳）</span>
        <input name="ctime" type="number" min="0" placeholder="留空则自动写入当前时间">
        <small data-error-for="ctime"></small>
      </label>
    </div>
    <label class="form-field">
      <span>摘要</span>
      <textarea name="description" rows="3" maxlength="500" required></textarea>
      <small data-error-for="description"></small>
    </label>
    <label class="form-field">
      <span>SEO 描述</span>
      <textarea name="seo_description" rows="3" maxlength="500" required></textarea>
      <small data-error-for="seo_description"></small>
    </label>
    <label class="form-field">
      <span>正文内容（支持纯文本或 HTML）</span>
      <textarea name="content" rows="16" class="editor-area" required></textarea>
      <small data-error-for="content"></small>
    </label>
    <div class="form-actions">
      <button class="primary-button" type="submit" data-submit-mode="stay">保存</button>
      <button class="secondary-button" type="submit" data-submit-mode="back">保存并返回</button>
      <a class="text-button" href="/admin/article">取消</a>
      <button id="article-delete-button" class="danger-button hidden" type="button">删除文章</button>
    </div>
  </form>
</section>
`

const tagListMarkup = `
<section class="panel">
  <form id="tag-search-form" class="toolbar">
    <input name="name" type="text" placeholder="按标签名称搜索">
    <button class="primary-button" type="submit">搜索</button>
  </form>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>标签名</th>
          <th>引用次数</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody id="tag-table-body"></tbody>
    </table>
  </div>
</section>
`
