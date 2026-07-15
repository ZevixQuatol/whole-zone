import Link from "next/link";

export default function Home() {
  return (
    <main className="auth-page">
      <section className="auth-story">
        <div className="brand">✦ 寰域</div>
        <div className="story-copy">
          <h1>新时代的个人名片，新时代的沟通方式。</h1>
          <p>首个 Account 模块已经建立。受邀用户可以注册、登录并维护公开资料。</p>
        </div>
      </section>
      <section className="auth-main">
        <div className="form-panel">
          <h2>开始建立你的寰点名片</h2>
          <p>当前为邀请内测。使用邀请码创建账号，或登录已有账号。</p>
          <div style={{ display: "flex", gap: 12 }}>
            <Link className="button" href="/register">受邀注册</Link>
            <Link className="button secondary" href="/login">登录</Link>
          </div>
        </div>
      </section>
    </main>
  );
}
