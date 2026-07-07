#!/usr/bin/env python3
"""Renders this repo's Markdown docs (docs/index.md, docs/apps/*.md, and the
repo-root README.md) into static, responsive HTML pages alongside their
source files. No CDN, no JS build step: stdlib + the already-installed
`markdown` package.

Run from anywhere: `python3 docs/build.py`
"""

import pathlib

import markdown

REPO_ROOT = pathlib.Path(__file__).resolve().parent.parent
DOCS_DIR = REPO_ROOT / "docs"

TEMPLATE = """<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{title}</title>
<style>
  :root {{
    color-scheme: light dark;
  }}
  body {{
    max-width: 46rem;
    margin: 0 auto;
    padding: 1.5rem 1rem 4rem;
    font-family: -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    line-height: 1.55;
    background: #ffffff;
    color: #1a1a1a;
  }}
  @media (prefers-color-scheme: dark) {{
    body {{
      background: #1a1a1e;
      color: #e8e8ea;
    }}
    a {{ color: #7ab8ff; }}
    code, pre {{ background: #2a2a30; }}
    img {{ background: #ffffff10; }}
  }}
  h1, h2, h3 {{ line-height: 1.25; }}
  img {{
    max-width: 100%;
    height: auto;
    display: block;
    border-radius: 6px;
    margin: 0.75rem 0;
  }}
  code {{
    background: #f0f0f2;
    padding: 0.1em 0.35em;
    border-radius: 4px;
    font-size: 0.92em;
  }}
  pre {{
    background: #f0f0f2;
    padding: 0.75rem 1rem;
    border-radius: 6px;
    overflow-x: auto;
  }}
  pre code {{ background: none; padding: 0; }}
  @media (max-width: 600px) {{
    body {{ padding: 1rem 0.75rem 3rem; }}
  }}
</style>
</head>
<body>
{content}
</body>
</html>
"""


def render(md_path: pathlib.Path, html_path: pathlib.Path, title: str, text: str | None = None) -> None:
    if text is None:
        text = md_path.read_text(encoding="utf-8")
    content = markdown.markdown(text, extensions=["fenced_code"])
    html_path.write_text(TEMPLATE.format(title=title, content=content), encoding="utf-8")
    print(f"wrote {html_path.relative_to(REPO_ROOT)}")


def main() -> None:
    render(DOCS_DIR / "index.md", DOCS_DIR / "index.html", "Go + SDL3 example apps")

    apps_dir = DOCS_DIR / "apps"
    for md_path in sorted(apps_dir.glob("*.md")):
        title = md_path.stem.replace("-", " ").title()
        render(md_path, apps_dir / f"{md_path.stem}.html", f"{title} — Go + SDL3 examples")

    # README.md's own "docs/index.html" link is written for GitHub, where it
    # reads relative to the repo root. Rendered here it lives *inside*
    # docs/ alongside index.html, so that same link needs to drop the
    # "docs/" prefix or it'd point at docs/docs/index.html.
    readme_path = REPO_ROOT / "README.md"
    readme_text = readme_path.read_text(encoding="utf-8").replace("docs/index.html", "index.html")
    render(readme_path, DOCS_DIR / "readme.html", "README — go-sdl3", text=readme_text)


if __name__ == "__main__":
    main()
