<template>
  <main class="page">
    <section class="hero">
      <div class="badge">Code Token</div>
      <h1>Code Token</h1>
      <p>稳定&流畅的AI编程体验。。。</p>

      <div class="hero-actions">
        <a class="ghost-link" href="#tutorial">查看使用教程</a>
      </div>

      <div class="hero-grid">
        <article class="card">
          <div class="card-tag">套餐服务</div>
          <h2>购买与兑换</h2>
          <p>
            购买后可在站内控制台创建 API Key，
            如已有兑换码，也可以直接进入兑换页面完成激活。
          </p>
          <div class="card-links package-links">
            <a class="btn btn-primary" href="https://pay.ldxp.cn/shop/FED14QEA" target="_blank" rel="noreferrer">购买套餐</a>
            <a class="btn btn-secondary" href="/redeem">兑换入口</a>
          </div>
        </article>

        <article class="card">
          <div class="card-tag">控制台</div>
          <h2>登录后直接使用</h2>
          <p>
            进入控制台后可以创建 API Key、查看订阅与用量。
            若你已经登录，点击下方按钮会自动跳转到控制台页面。
          </p>
          <div class="card-links console-links">
            <a class="btn btn-primary" href="/login">进入控制台</a>
          </div>
        </article>
      </div>

      <div class="scroll-tip">向下滚动查看 CC-Switch 配置教程</div>
    </section>

    <section id="tutorial" class="tutorial">
      <div class="section-head">
        <h3>基于CC-Switch快速配置教程。</h3>
        <p>
          先下载并安装 Codex 与 CC-Switch，再按下面步骤完成配置。
          教程示例使用供应商名称
          <span class="inline-code">x</span>，
          API 地址固定为
          <span class="inline-code">https://api.xhhtop.top/v1</span>。
        </p>
        <div class="download-links">
          <a class="btn btn-secondary" href="https://www.codex-docs.com/using-codex/app/" target="_blank" rel="noreferrer">下载 Codex</a>
          <a class="btn btn-secondary" href="https://github.com/farion1231/cc-switch/releases/download/v3.10.3/CC-Switch-v3.10.3-Windows-Portable.zip" target="_blank" rel="noreferrer">下载 CC-Switch</a>
        </div>
      </div>

      <div class="guide-list">
        <article class="guide-card">
          <div class="guide-copy">
            <div class="step">1</div>
            <h4>先打开一次 Codex，然后完全退出</h4>
            <ul>
              <li>先安装并打开 <span class="inline-code">Codex</span> 一次。</li>
              <li>然后退出 <span class="inline-code">Codex</span>，一定是打开并完全退出一次。</li>
              <li>这一步完成后，再打开 <span class="inline-code">CC-Switch</span> 进行 API 配置。</li>
            </ul>
          </div>
          <div class="guide-image-wrap">
            <img src="/dis2.png" alt="CC-Switch 中选中 Codex 并添加接口">
          </div>
        </article>

        <article class="guide-card reverse">
          <div class="guide-image-wrap">
            <img src="/dis1.png" alt="填写自定义供应商配置">
          </div>
          <div class="guide-copy">
            <div class="step">2</div>
            <h4>在 CC-Switch 中填写配置并保存</h4>
            <ul>
              <li>打开 <span class="inline-code">CC-Switch</span> 进行 API 配置。</li>
              <li>在 CC-Switch 顶部选择 <span class="inline-code">Codex</span>。</li>
              <li>然后点击右上角的 <span class="inline-code">+</span>，新增一个供应商配置。</li>
              <li>如果左侧已经出现 <span class="inline-code">x</span> 且地址为 <span class="inline-code">https://api.xhhtop.top/v1</span>，说明你之前已经配置过。可以直接启用该配置。</li>
              <li>预设供应商选择 <span class="inline-code">自定义配置</span>。</li>
              <li>供应商名称填写 <span class="inline-code">x</span>。</li>
              <li>API Key 在本站控制台创建后填入。</li>
              <li>API 请求地址固定填写 <span class="inline-code">https://api.xhhtop.top/v1</span>。</li>
              <li>模型名称可先填写 <span class="inline-code">gpt-5.5</span>，后续也可以在客户端内切换。</li>
              <li>最后点击右下角 <span class="inline-code">添加</span> 完成保存。</li>
              <li>再次打开 <span class="inline-code">Codex</span>，开始体验吧。</li>
            </ul>
          </div>
        </article>
      </div>

      <div class="config-box">
        <div class="config-head">高级设置参考</div>
        <pre><code>model_provider = "custom"
model = "gpt-5.5"
model_reasoning_effort = "high"
disable_response_storage = true
service_tier = "fast"
approvals_reviewer = "user"

[model_providers.custom]
name = "custom"
wire_api = "responses"
requires_openai_auth = true
base_url = "https://api.xhhtop.top/v1"</code></pre>
      </div>
    </section>
  </main>
</template>

<style scoped>
:global(html) {
  scroll-behavior: smooth;
}

:global(body) {
  margin: 0;
  font-family: "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
  color: var(--text);
  background:
    radial-gradient(circle at top left, rgba(15, 118, 110, 0.14), transparent 32%),
    radial-gradient(circle at right, rgba(37, 99, 235, 0.12), transparent 28%),
    linear-gradient(180deg, var(--bg) 0%, var(--bg-soft) 100%);
}

:global(a) {
  color: inherit;
  text-decoration: none;
}

.page {
  width: min(1180px, calc(100% - 28px));
  margin: 0 auto;
  padding: 28px 0 72px;
}

.hero {
  min-height: 78vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  padding: 48px 0 28px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--panel);
  color: var(--muted);
  font-size: 13px;
  backdrop-filter: blur(10px);
  box-shadow: var(--shadow-soft);
}

.hero h1 {
  margin: 22px 0 14px;
  font-size: clamp(34px, 6vw, 64px);
  line-height: 1.04;
  letter-spacing: -0.04em;
}

.hero p {
  margin: 0 auto;
  max-width: 760px;
  font-size: 16px;
  line-height: 1.9;
  color: var(--muted);
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 14px;
  margin-top: 22px;
}

.ghost-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 46px;
  padding: 0 18px;
  border-radius: 14px;
  border: 1px solid var(--line);
  background: var(--panel);
  color: var(--text);
  box-shadow: var(--shadow-soft);
}

.hero-grid {
  width: 100%;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
  margin-top: 34px;
}

.card {
  position: relative;
  overflow: hidden;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 28px;
  padding: 28px;
  box-shadow: var(--shadow);
  backdrop-filter: blur(12px);
}

.card::before {
  content: "";
  position: absolute;
  inset: 0 auto auto 0;
  width: 100%;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,.75), transparent);
  opacity: 0.35;
  pointer-events: none;
}

.card-tag {
  display: inline-block;
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  background: var(--accent-soft);
  color: var(--accent);
  margin-bottom: 14px;
}

.card h2 {
  margin: 0 0 10px;
  font-size: 28px;
  line-height: 1.15;
  text-align: center;
}

.card p {
  margin: 0;
  color: var(--muted);
  line-height: 1.85;
  font-size: 15px;
  text-align: center;
}

.card-links {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 22px;
}

.card-links.package-links {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  align-items: center;
}

.card-links.console-links {
  justify-content: center;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 46px;
  padding: 0 18px;
  border-radius: 14px;
  font-size: 14px;
  font-weight: 600;
  transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease;
  text-align: center;
}

.btn:hover {
  transform: translateY(-1px);
}

.btn-primary {
  color: #fff;
  background: linear-gradient(135deg, var(--accent), #14b8a6);
  box-shadow: 0 12px 24px rgba(15, 118, 110, 0.18);
}

.btn-secondary {
  color: var(--text);
  background: var(--panel-strong);
  border: 1px solid var(--line);
}

.scroll-tip {
  margin-top: 26px;
  color: var(--muted);
  font-size: 13px;
  letter-spacing: 0.04em;
}

.tutorial {
  padding-top: 30px;
}

.section-head {
  text-align: center;
  margin-bottom: 24px;
}

.section-head h3 {
  margin: 0 0 10px;
  font-size: clamp(28px, 4vw, 40px);
  letter-spacing: -0.03em;
}

.section-head p {
  margin: 0 auto;
  max-width: 760px;
  color: var(--muted);
  line-height: 1.9;
  font-size: 15px;
}

.download-links {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
  margin-top: 18px;
}

.guide-list {
  display: grid;
  gap: 22px;
  margin-top: 24px;
}

.guide-card {
  display: grid;
  grid-template-columns: 0.92fr 1.08fr;
  gap: 24px;
  align-items: center;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 28px;
  padding: 22px;
  box-shadow: var(--shadow);
  backdrop-filter: blur(12px);
}

.guide-card.reverse {
  grid-template-columns: 1.08fr 0.92fr;
}

.guide-image-wrap {
  background: var(--panel-strong);
  border: 1px solid var(--line);
  border-radius: 22px;
  padding: 12px;
  box-shadow: var(--shadow-soft);
}

.guide-image-wrap img {
  display: block;
  width: 100%;
  height: auto;
  border-radius: 16px;
}

.step {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--accent-2), var(--accent));
  color: #fff;
  font-weight: 700;
  margin-bottom: 14px;
  box-shadow: 0 10px 20px rgba(37, 99, 235, 0.18);
}

.guide-card h4 {
  margin: 0 0 12px;
  font-size: 24px;
  line-height: 1.2;
}

.guide-card p,
.guide-card li {
  color: var(--muted);
  font-size: 15px;
  line-height: 1.85;
}

.guide-card ul {
  margin: 0;
  padding-left: 20px;
}

.guide-card li + li {
  margin-top: 6px;
}

.inline-code,
code {
  font-family: "Cascadia Code", "JetBrains Mono", Consolas, monospace;
  font-size: 0.95em;
}

.inline-code {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 8px;
  background: var(--accent-soft);
  color: var(--accent);
}

.config-box {
  margin-top: 22px;
  border: 1px solid var(--line);
  border-radius: 24px;
  overflow: hidden;
  box-shadow: var(--shadow);
  background: var(--panel);
}

.config-head {
  padding: 16px 20px;
  border-bottom: 1px solid var(--line);
  background: var(--panel-strong);
  font-weight: 600;
}

pre {
  margin: 0;
  padding: 20px;
  overflow: auto;
  background: var(--code-bg);
  color: var(--code-text);
  font-size: 13px;
  line-height: 1.75;
}

@media (max-width: 920px) {
  .hero {
    min-height: auto;
    padding-top: 40px;
  }

  .hero-grid,
  .guide-card,
  .guide-card.reverse,
  .card-links.package-links {
    grid-template-columns: 1fr;
  }

  .card,
  .guide-card {
    border-radius: 22px;
  }

  .card h2 {
    font-size: 24px;
  }

  .guide-card h4 {
    font-size: 22px;
  }
}
</style>
