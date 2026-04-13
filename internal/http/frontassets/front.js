(function () {
  function initNav() {
    var toggle = document.querySelector("[data-nav-toggle]");
    var panel = document.querySelector("[data-nav-panel]");
    if (!toggle || !panel) {
      return;
    }
    toggle.addEventListener("click", function () {
      panel.classList.toggle("is-open");
    });
  }

  function initVotes() {
    var actions = document.querySelector("[data-article-id]");
    if (!actions) {
      return;
    }

    var articleID = actions.getAttribute("data-article-id");
    var buttons = document.querySelectorAll("[data-vote]");
    var goodNum = document.querySelector("[data-good-num]");
    var badNum = document.querySelector("[data-bad-num]");
    var message = document.querySelector("[data-action-message]");

    buttons.forEach(function (button) {
      button.addEventListener("click", async function () {
        var type = button.getAttribute("data-vote");
        var endpoint = type === "good" ? "good" : "bad";
        buttons.forEach(function (item) { item.disabled = true; });
        if (message) {
          message.textContent = "提交中...";
        }
        try {
          var response = await fetch("/api/v1/front/articles/" + articleID + "/" + endpoint, {
            method: "POST",
            headers: {
              "Content-Type": "application/json"
            }
          });
          var data = await response.json();
          if (!response.ok) {
            throw new Error(data.error || "请求失败");
          }
          if (goodNum) {
            goodNum.textContent = String(data.good_num);
          }
          if (badNum) {
            badNum.textContent = String(data.bad_num);
          }
          if (message) {
            message.textContent = type === "good" ? "感谢点赞" : "已记录反馈";
          }
        } catch (error) {
          if (message) {
            message.textContent = error.message || "请求失败";
          }
        } finally {
          buttons.forEach(function (item) { item.disabled = false; });
        }
      });
    });
  }

  initNav();
  initVotes();
})();
