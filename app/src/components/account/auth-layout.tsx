import Link from "next/link";

export function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="auth-page">
      <section className="auth-story">
        <Link className="brand" href="/">✦ 寰域</Link>
        <div className="story-copy">
          <h1>让世界通过寰点，真正认识你。</h1>
          <p>用爱好、专业、职业和持续参与构成新时代名片，在共同的域里找到内容与同行者。</p>
        </div>
      </section>
      <section className="auth-main">{children}</section>
    </main>
  );
}
