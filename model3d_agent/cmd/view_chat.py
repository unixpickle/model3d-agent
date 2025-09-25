"""
Render a chat transcript JSON file from run_simulations.py or run_agent*.py.

Wrote mostly by GPT-5 with a few fixes from myself.
"""

import html
import json
import os
import re
import sys
import tempfile
import webbrowser
from pathlib import Path
from typing import Any, Dict, List

# --- Minimal "markdown-ish" formatter for code fences and newlines ---
_CODE_FENCE_RE = re.compile(r"```([a-zA-Z0-9_\-]*)\n(.*?)```", re.DOTALL)


def _format_text(s: str) -> str:
    # Escape HTML
    s = html.escape(s)

    # Convert triple-backtick code blocks first
    def repl(m: re.Match) -> str:
        lang = m.group(1).strip()
        code = m.group(2)
        return f'<pre class="codeblock"><code class="lang-{lang}">{code}</code></pre>'

    s = _CODE_FENCE_RE.sub(repl, s)
    # Convert bare newlines to <br> but avoid inside code blocks (already handled)
    s = s.replace("\n", "<br>")
    return s


AssetDir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "view_chat_assets")


def render_message(msg: Dict[str, Any]) -> str:
    role = html.escape(msg.get("role", ""))
    role_class = role.lower().strip()
    content = msg.get("content", [])
    blocks: List[str] = []

    if isinstance(content, str):
        content = [dict(type="input_text", text=content)]

    for chunk in content:
        t = chunk.get("type")
        if t in ("input_text", "output_text"):
            text = _format_text(str(chunk.get("text", "")))
            blocks.append(
                f"""
              <div class="chunk">
                <h4>{"User text" if t == "input_text" else "Assistant text"}</h4>
                <div class="text">{text}</div>
              </div>
            """
            )
        elif t in ("input_image", "output_image"):
            # image_url could be data: URI or http(s)
            src = str(chunk.get("image_url", ""))
            cap = "Image"
            blocks.append(
                f"""
              <div class="chunk">
                <h4>{cap}</h4>
                <img class="chat-image" loading="lazy" src="{html.escape(src)}" alt="conversation image"/>
              </div>
            """
            )
        else:
            safe = _format_text(json.dumps(chunk, ensure_ascii=False, indent=2))
            blocks.append(
                f"""
              <div class="chunk">
                <h4>Unknown content ({html.escape(str(t))})</h4>
                <pre class="codeblock"><code>{safe}</code></pre>
              </div>
            """
            )

    return f"""
      <div class="msg {role_class}">
        <div class="role">{role}</div>
        <div class="body">
          {"".join(blocks)}
        </div>
      </div>
    """


def render_page(data: Dict[str, Any]) -> str:
    with open(os.path.join(AssetDir, "style.css"), "r") as f:
        CSS = f.read()
    with open(os.path.join(AssetDir, "script.js"), "r") as f:
        JS = f.read()

    conv = data["messages"]

    messages_html = "\n".join(render_message(m) for m in conv)

    raw_json = html.escape(json.dumps(data, ensure_ascii=False, indent=2))

    return f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Chat Render</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>{CSS}</style>
<script>{JS}</script>
</head>
<body>
  <div class="container">
    <div class="header">
      <div class="actions">
        <button id="copy-btn" class="button" onclick="copyRaw()">Copy raw JSON</button>
        <button class="button" onclick="toggleRaw()">Toggle raw</button>
      </div>
    </div>

    <div id="raw-wrap" style="display:none; margin-top:10px;">
      <div class="msg">
        <div class="role">RAW</div>
        <div class="body">
          <div class="chunk">
            <pre id="raw-json" class="codeblock"><code>{raw_json}</code></pre>
          </div>
        </div>
      </div>
    </div>

    <div class="chat">
      {messages_html}
    </div>

    <div class="footer">
      Generated locally • No external dependencies
    </div>
  </div>
</body>
</html>
"""


def main():
    # Read JSON from a file path arg or stdin
    if len(sys.argv) > 1 and sys.argv[1] != "-":
        src = Path(sys.argv[1]).read_text(encoding="utf-8")
    else:
        src = sys.stdin.read()

    try:
        data = json.loads(src)
    except json.JSONDecodeError as e:
        print(f"Invalid JSON: {e}", file=sys.stderr)
        sys.exit(1)

    html_doc = render_page(data)

    # Write to a temp file next to the JSON (if path provided) or system temp
    base_dir = None
    if len(sys.argv) > 1 and sys.argv[1] != "-":
        base_dir = str(Path(sys.argv[1]).resolve().parent)
    tmp = tempfile.NamedTemporaryFile(
        prefix="chat_render_", suffix=".html", delete=False, dir=base_dir
    )
    with open(tmp.name, "w", encoding="utf-8") as f:
        f.write(html_doc)

    # Open in default browser
    webbrowser.open("file://" + tmp.name)
    print(f"Wrote {tmp.name}")


if __name__ == "__main__":
    main()
