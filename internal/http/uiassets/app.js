(function () {
  const body = document.body;
  if (!body) {
    return;
  }

  const page = body.dataset.page || "";
  const recordID = Number(body.dataset.recordId || 0);
  const redirectAfterLogin = body.dataset.redirect || "";

  document.addEventListener("DOMContentLoaded", function () {
    bindLogout();
    renderSavedHint();

    switch (page) {
      case "login":
        initLogin();
        break;
      case "dashboard":
        initDashboard();
        break;
      case "category-list":
        initCategoryList();
        break;
      case "category-form":
        initCategoryForm();
        break;
      case "article-list":
        initArticleList();
        break;
      case "article-form":
        initArticleForm();
        break;
      case "tag-list":
        initTagList();
        break;
      default:
        break;
    }
  });

  function bindLogout() {
    const button = document.getElementById("logout-button");
    if (!button) {
      return;
    }
    button.addEventListener("click", async function () {
      await fetch("/api/v1/auth/logout", {
        method: "POST",
        credentials: "include",
      });
      window.location.href = "/admin/login";
    });
  }

  function renderSavedHint() {
    const url = new URL(window.location.href);
    const message = url.searchParams.get("message");
    if (message) {
      showBanner("success", decodeURIComponent(message));
      url.searchParams.delete("message");
      history.replaceState({}, "", url.toString());
    }
  }

  async function apiFetch(url, options) {
    const response = await fetch(url, Object.assign({
      credentials: "include",
      headers: {},
    }, options || {}));

    if (response.status === 401) {
      const next = encodeURIComponent(window.location.pathname + window.location.search);
      window.location.href = "/admin/login?redirect=" + next;
      throw new Error("unauthorized");
    }
    return response;
  }

  async function apiJSON(url, options) {
    const response = await apiFetch(url, options);
    const text = await response.text();
    let data = {};
    if (text) {
      try {
        data = JSON.parse(text);
      } catch (error) {
        throw new Error("响应解析失败");
      }
    }

    if (!response.ok) {
      const err = new Error(data.error || "请求失败");
      err.fields = data.fields || {};
      err.status = response.status;
      throw err;
    }

    return data;
  }

  function initLogin() {
    const form = document.getElementById("login-form");
    if (!form) {
      return;
    }

    form.addEventListener("submit", async function (event) {
      event.preventDefault();
      const username = form.username.value.trim();
      const password = form.password.value;

      if (!username || !password) {
        showBanner("error", "请输入用户名和密码");
        return;
      }

      setPending(form, true);
      try {
        await apiJSON("/api/v1/auth/login", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            username: username,
            password: password,
          }),
        });
        window.location.href = redirectAfterLogin || "/admin";
      } catch (error) {
        showBanner("error", error.message || "登录失败");
      } finally {
        setPending(form, false);
      }
    });
  }

  async function initDashboard() {
    const data = await apiJSON("/api/v1/admin/dashboard");
    const summaryRoot = document.getElementById("dashboard-summary");
    const recentRoot = document.getElementById("recent-articles");
    const quickLinksRoot = document.getElementById("quick-links");

    const summaryItems = [
      { label: "文章总数", value: data.summary.article_count },
      { label: "分类总数", value: data.summary.category_count },
      { label: "标签总数", value: data.summary.tag_count },
      { label: "全站推荐", value: data.summary.recommended_site_count },
      { label: "首页推荐", value: data.summary.recommended_home_count },
    ];

    summaryRoot.innerHTML = summaryItems.map(function (item) {
      return [
        '<article class="card">',
        '<div class="card-label">' + escapeHTML(String(item.label)) + "</div>",
        '<div class="card-value">' + escapeHTML(String(item.value)) + "</div>",
        "</article>"
      ].join("");
    }).join("");

    if (data.recent_articles.length === 0) {
      recentRoot.innerHTML = '<div class="empty-state">暂无最近文章</div>';
    } else {
      recentRoot.innerHTML = '<ol class="recent-list">' + data.recent_articles.map(function (item) {
        return '<li><strong>' + escapeHTML(item.title) + '</strong><div class="muted">' + escapeHTML(item.category_name || "-") + " | " + formatUnix(item.ctime) + "</div></li>";
      }).join("") + "</ol>";
    }

    quickLinksRoot.innerHTML = data.quick_links.map(function (item) {
      return '<a class="quick-link" href="' + escapeHTML(item.path) + '">' + escapeHTML(item.name) + "</a>";
    }).join("");
  }

  function initCategoryList() {
    const form = document.getElementById("category-search-form");
    if (!form) {
      return;
    }

    hydrateSearchForm(form);
    form.addEventListener("submit", function (event) {
      event.preventDefault();
      const params = new URLSearchParams();
      const name = form.name.value.trim();
      if (name) {
        params.set("name", name);
      }
      window.location.search = params.toString();
    });

    loadCategories();
  }

  async function loadCategories() {
    const tableBody = document.getElementById("category-table-body");
    const params = new URLSearchParams(window.location.search);
    const data = await apiJSON("/api/v1/admin/categories?" + params.toString());

    if (data.items.length === 0) {
      tableBody.innerHTML = '<tr><td colspan="8" class="empty-state">暂无分类数据</td></tr>';
      return;
    }

    tableBody.innerHTML = data.items.map(function (item) {
      const levelClass = item.level > 0 ? "indent-" + item.level : "";
      const name = item.level > 0 ? "&boxur; " + escapeHTML(item.name) : escapeHTML(item.name);
      return [
        "<tr>",
        "<td>" + item.id + "</td>",
        '<td class="' + levelClass + '">' + name + "</td>",
        "<td>" + escapeHTML(item.parent_name || "顶级") + "</td>",
        "<td>" + escapeHTML(item.in_out === 2 ? "站外" : "站内") + "</td>",
        "<td>" + escapeHTML(item.url || "-") + "</td>",
        "<td>" + escapeHTML(item.seo_title) + "</td>",
        "<td>" + escapeHTML(item.item) + "</td>",
        '<td class="actions"><a class="table-link" href="/admin/category/edit/' + item.id + '">编辑</a><button class="table-danger" type="button" data-category-delete="' + item.id + '">删除</button></td>',
        "</tr>"
      ].join("");
    }).join("");

    tableBody.querySelectorAll("[data-category-delete]").forEach(function (button) {
      button.addEventListener("click", async function () {
        const id = button.getAttribute("data-category-delete");
        if (!window.confirm("确认删除该分类吗？")) {
          return;
        }
        try {
          await apiJSON("/api/v1/admin/categories/" + id, { method: "DELETE" });
          showBanner("success", "分类已删除");
          loadCategories();
        } catch (error) {
          showBanner("error", error.message || "删除失败");
        }
      });
    });
  }

  async function initCategoryForm() {
    const form = document.getElementById("category-form");
    if (!form) {
      return;
    }

    const submitMode = { value: "stay" };
    form.querySelectorAll("[data-submit-mode]").forEach(function (button) {
      button.addEventListener("click", function () {
        submitMode.value = button.dataset.submitMode || "stay";
      });
    });

    const categories = await apiJSON("/api/v1/admin/categories");
    fillCategoryParentOptions(categories.items, recordID);
    toggleCategoryURLField();
    document.getElementById("category-link-type").addEventListener("change", toggleCategoryURLField);

    if (recordID > 0) {
      const detail = await apiJSON("/api/v1/admin/categories/" + recordID);
      fillCategoryForm(form, detail);
      document.getElementById("category-delete-button").classList.remove("hidden");
      document.getElementById("category-delete-button").addEventListener("click", async function () {
        if (!window.confirm("确认删除该分类吗？")) {
          return;
        }
        try {
          await apiJSON("/api/v1/admin/categories/" + recordID, { method: "DELETE" });
          window.location.href = "/admin/category?message=" + encodeURIComponent("分类已删除");
        } catch (error) {
          showBanner("error", error.message || "删除失败");
        }
      });
    }

    form.addEventListener("submit", async function (event) {
      event.preventDefault();
      clearFieldErrors(form);
      showBanner("", "");

      const payload = {
        name: form.name.value.trim(),
        seo_title: form.seo_title.value.trim(),
        seo_description: form.seo_description.value.trim(),
        seo_keywords: form.seo_keywords.value.trim(),
        in_out: Number(form.in_out.value),
        pid: Number(form.pid.value),
        url: form.url.value.trim(),
        item: form.item.value.trim(),
      };

      setPending(form, true);
      try {
        const method = recordID > 0 ? "PUT" : "POST";
        const url = recordID > 0 ? "/api/v1/admin/categories/" + recordID : "/api/v1/admin/categories";
        const result = await apiJSON(url, {
          method: method,
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(payload),
        });
        if (submitMode.value === "back") {
          window.location.href = "/admin/category?message=" + encodeURIComponent("分类保存成功");
          return;
        }
        if (recordID > 0) {
          showBanner("success", "分类保存成功");
        } else {
          window.location.href = "/admin/category/edit/" + result.id + "?message=" + encodeURIComponent("分类创建成功");
        }
      } catch (error) {
        setFieldErrors(form, error.fields || {});
        showBanner("error", error.message || "保存失败");
      } finally {
        setPending(form, false);
      }
    });
  }

  function fillCategoryParentOptions(items, currentID) {
    const select = document.getElementById("category-parent");
    select.innerHTML = '<option value="0">顶级分类</option>';
    items.filter(function (item) {
      return item.level === 0 && item.id !== currentID;
    }).forEach(function (item) {
      const option = document.createElement("option");
      option.value = String(item.id);
      option.textContent = item.name;
      select.appendChild(option);
    });
  }

  function fillCategoryForm(form, detail) {
    form.name.value = detail.name || "";
    form.seo_title.value = detail.seo_title || "";
    form.seo_description.value = detail.seo_description || "";
    form.seo_keywords.value = detail.seo_keywords || "";
    form.in_out.value = String(detail.in_out || 1);
    form.pid.value = String(detail.pid || 0);
    form.url.value = detail.url || "";
    form.item.value = detail.item || "";
    toggleCategoryURLField();
  }

  function toggleCategoryURLField() {
    const field = document.getElementById("category-url-field");
    const type = document.getElementById("category-link-type").value;
    field.classList.toggle("hidden", type !== "2");
  }

  async function initArticleList() {
    const form = document.getElementById("article-search-form");
    if (!form) {
      return;
    }

    hydrateSearchForm(form);
    const categories = await apiJSON("/api/v1/admin/categories");
    const categorySelect = document.getElementById("article-filter-category");
    categorySelect.innerHTML = '<option value="">全部分类</option>';
    categories.items.filter(function (item) {
      return item.in_out === 1;
    }).forEach(function (item) {
      const option = document.createElement("option");
      option.value = String(item.id);
      option.textContent = item.name;
      categorySelect.appendChild(option);
    });
    form.mid.value = new URLSearchParams(window.location.search).get("mid") || "";

    form.addEventListener("submit", function (event) {
      event.preventDefault();
      const params = new URLSearchParams();
      [
        ["title", form.title.value.trim()],
        ["mid", form.mid.value],
        ["recommend_type", form.recommend_type.value],
      ].forEach(function (entry) {
        if (entry[1]) {
          params.set(entry[0], entry[1]);
        }
      });
      window.location.search = params.toString();
    });

    loadArticles();
  }

  async function loadArticles() {
    const params = new URLSearchParams(window.location.search);
    if (!params.get("page")) {
      params.set("page", "1");
    }
    if (!params.get("page_size")) {
      params.set("page_size", "10");
    }

    const data = await apiJSON("/api/v1/admin/articles?" + params.toString());
    const tableBody = document.getElementById("article-table-body");

    if (data.items.length === 0) {
      tableBody.innerHTML = '<tr><td colspan="11" class="empty-state">暂无文章数据</td></tr>';
    } else {
      tableBody.innerHTML = data.items.map(function (item) {
        return [
          "<tr>",
          "<td>" + item.id + "</td>",
          "<td>" + escapeHTML(item.title) + "</td>",
          "<td>" + escapeHTML(item.category_name || "-") + "</td>",
          "<td>" + escapeHTML(item.author) + "</td>",
          "<td>" + escapeHTML(item.tag) + "</td>",
          "<td>" + item.clicks + "</td>",
          "<td>" + item.good_num + "</td>",
          "<td>" + item.bad_num + "</td>",
          "<td>" + escapeHTML(item.recommend_type === 2 ? "首页推荐" : "全站推荐") + "</td>",
          "<td>" + formatUnix(item.ctime) + "</td>",
          '<td class="actions"><a class="table-link" href="/admin/article/edit/' + item.id + '">编辑</a><button class="table-danger" type="button" data-article-delete="' + item.id + '">删除</button></td>',
          "</tr>"
        ].join("");
      }).join("");
    }

    tableBody.querySelectorAll("[data-article-delete]").forEach(function (button) {
      button.addEventListener("click", async function () {
        if (!window.confirm("确认删除该文章吗？")) {
          return;
        }
        try {
          await apiJSON("/api/v1/admin/articles/" + button.dataset.articleDelete, { method: "DELETE" });
          showBanner("success", "文章已删除");
          loadArticles();
        } catch (error) {
          showBanner("error", error.message || "删除失败");
        }
      });
    });

    renderPagination(data.page, data.total_pages);
  }

  async function initArticleForm() {
    const form = document.getElementById("article-form");
    if (!form) {
      return;
    }

    const submitMode = { value: "stay" };
    form.querySelectorAll("[data-submit-mode]").forEach(function (button) {
      button.addEventListener("click", function () {
        submitMode.value = button.dataset.submitMode || "stay";
      });
    });

    const categories = await apiJSON("/api/v1/admin/categories");
    const options = categories.items.filter(function (item) {
      return item.in_out === 1;
    });
    const select = document.getElementById("article-category-select");
    select.innerHTML = '<option value="">请选择分类</option>';
    options.forEach(function (item) {
      const option = document.createElement("option");
      option.value = String(item.id);
      option.textContent = item.level > 0 ? "  - " + item.name : item.name;
      select.appendChild(option);
    });

    if (recordID > 0) {
      const detail = await apiJSON("/api/v1/admin/articles/" + recordID);
      fillArticleForm(form, detail);
      document.getElementById("article-delete-button").classList.remove("hidden");
      document.getElementById("article-delete-button").addEventListener("click", async function () {
        if (!window.confirm("确认删除该文章吗？")) {
          return;
        }
        try {
          await apiJSON("/api/v1/admin/articles/" + recordID, { method: "DELETE" });
          window.location.href = "/admin/article?message=" + encodeURIComponent("文章已删除");
        } catch (error) {
          showBanner("error", error.message || "删除失败");
        }
      });
    }

    form.addEventListener("submit", async function (event) {
      event.preventDefault();
      clearFieldErrors(form);
      showBanner("", "");

      const payload = {
        mid: Number(form.mid.value),
        author: form.author.value.trim(),
        title: form.title.value.trim(),
        description: form.description.value.trim(),
        seo_title: form.seo_title.value.trim(),
        seo_description: form.seo_description.value.trim(),
        seo_keywords: form.seo_keywords.value.trim(),
        tag: form.tag.value.trim(),
        clicks: Number(form.clicks.value || 0),
        content: form.content.value,
        ctime: Number(form.ctime.value || 0),
        good_num: Number(form.good_num.value || 0),
        bad_num: Number(form.bad_num.value || 0),
        recommend_type: Number(form.recommend_type.value),
      };

      setPending(form, true);
      try {
        const method = recordID > 0 ? "PUT" : "POST";
        const url = recordID > 0 ? "/api/v1/admin/articles/" + recordID : "/api/v1/admin/articles";
        const result = await apiJSON(url, {
          method: method,
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(payload),
        });

        if (submitMode.value === "back") {
          window.location.href = "/admin/article?message=" + encodeURIComponent("文章保存成功");
          return;
        }
        if (recordID > 0) {
          showBanner("success", "文章保存成功");
        } else {
          window.location.href = "/admin/article/edit/" + result.id + "?message=" + encodeURIComponent("文章创建成功");
        }
      } catch (error) {
        setFieldErrors(form, error.fields || {});
        showBanner("error", error.message || "保存失败");
      } finally {
        setPending(form, false);
      }
    });
  }

  function fillArticleForm(form, detail) {
    form.mid.value = String(detail.mid || "");
    form.author.value = detail.author || "";
    form.title.value = detail.title || "";
    form.description.value = detail.description || "";
    form.seo_title.value = detail.seo_title || "";
    form.seo_description.value = detail.seo_description || "";
    form.seo_keywords.value = detail.seo_keywords || "";
    form.tag.value = detail.tag || "";
    form.clicks.value = detail.clicks || 0;
    form.content.value = detail.content || "";
    form.ctime.value = detail.ctime || "";
    form.good_num.value = detail.good_num || 0;
    form.bad_num.value = detail.bad_num || 0;
    form.recommend_type.value = String(detail.recommend_type || 1);
  }

  function initTagList() {
    const form = document.getElementById("tag-search-form");
    if (!form) {
      return;
    }

    hydrateSearchForm(form);
    form.addEventListener("submit", function (event) {
      event.preventDefault();
      const params = new URLSearchParams();
      const name = form.name.value.trim();
      if (name) {
        params.set("name", name);
      }
      window.location.search = params.toString();
    });

    loadTags();
  }

  async function loadTags() {
    const params = new URLSearchParams(window.location.search);
    const data = await apiJSON("/api/v1/admin/tags?" + params.toString());
    const tableBody = document.getElementById("tag-table-body");

    if (data.items.length === 0) {
      tableBody.innerHTML = '<tr><td colspan="4" class="empty-state">暂无标签数据</td></tr>';
      return;
    }

    tableBody.innerHTML = data.items.map(function (item) {
      const actionText = item.num > 0 ? "已被引用" : "删除";
      const actionDisabled = item.num > 0 ? "disabled" : "";
      return [
        "<tr>",
        "<td>" + item.id + "</td>",
        "<td>" + escapeHTML(item.tag) + "</td>",
        '<td><span class="chip">' + item.num + "</span></td>",
        '<td class="actions"><button class="table-danger" type="button" data-tag-delete="' + item.id + '" ' + actionDisabled + ">" + actionText + "</button></td>",
        "</tr>"
      ].join("");
    }).join("");

    tableBody.querySelectorAll("[data-tag-delete]").forEach(function (button) {
      if (button.disabled) {
        return;
      }
      button.addEventListener("click", async function () {
        if (!window.confirm("确认删除该标签吗？")) {
          return;
        }
        try {
          await apiJSON("/api/v1/admin/tags/" + button.dataset.tagDelete, { method: "DELETE" });
          showBanner("success", "标签已删除");
          loadTags();
        } catch (error) {
          showBanner("error", error.message || "删除失败");
        }
      });
    });
  }

  function renderPagination(pageNumber, totalPages) {
    const root = document.getElementById("article-pagination");
    if (!root) {
      return;
    }
    if (totalPages <= 1) {
      root.innerHTML = "";
      return;
    }

    const prevDisabled = pageNumber <= 1 ? "disabled" : "";
    const nextDisabled = pageNumber >= totalPages ? "disabled" : "";
    root.innerHTML = [
      '<button class="secondary-button" type="button" data-page-nav="prev" ' + prevDisabled + ">上一页</button>",
      "<span>第 " + pageNumber + " / " + totalPages + " 页</span>",
      '<button class="secondary-button" type="button" data-page-nav="next" ' + nextDisabled + ">下一页</button>",
    ].join("");

    root.querySelectorAll("[data-page-nav]").forEach(function (button) {
      button.addEventListener("click", function () {
        const params = new URLSearchParams(window.location.search);
        const current = Number(params.get("page") || 1);
        const direction = button.dataset.pageNav;
        const target = direction === "prev" ? current - 1 : current + 1;
        params.set("page", String(target));
        window.location.search = params.toString();
      });
    });
  }

  function hydrateSearchForm(form) {
    const params = new URLSearchParams(window.location.search);
    Array.from(form.elements).forEach(function (element) {
      if (!element.name) {
        return;
      }
      const value = params.get(element.name);
      if (value !== null) {
        element.value = value;
      }
    });
  }

  function setFieldErrors(form, fields) {
    Object.keys(fields).forEach(function (name) {
      const target = form.querySelector('[data-error-for="' + name + '"]');
      if (target) {
        target.textContent = fields[name];
      }
    });
  }

  function clearFieldErrors(form) {
    form.querySelectorAll("[data-error-for]").forEach(function (node) {
      node.textContent = "";
    });
  }

  function showBanner(type, message) {
    const banner = document.getElementById("page-banner");
    if (!banner) {
      return;
    }

    if (!message) {
      banner.hidden = true;
      banner.className = "banner";
      banner.textContent = "";
      return;
    }

    banner.hidden = false;
    banner.className = "banner " + type;
    banner.textContent = message;
  }

  function setPending(form, pending) {
    form.querySelectorAll("button").forEach(function (button) {
      button.disabled = pending;
    });
  }

  function formatUnix(value) {
    const timestamp = Number(value || 0);
    if (!timestamp) {
      return "-";
    }
    return new Date(timestamp * 1000).toLocaleString("zh-CN", {
      hour12: false,
    });
  }

  function escapeHTML(value) {
    return String(value)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }
})();
