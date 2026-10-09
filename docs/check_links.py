"""Check generated site links while respecting the configured Pages base path."""

from html.parser import HTMLParser
from pathlib import Path
import sys
from urllib.parse import unquote, urlsplit


class Links(HTMLParser):
    def __init__(self):
        super().__init__()
        self.links = []

    def handle_starttag(self, tag, attrs):
        for name, value in attrs:
            if name in ("href", "src") and value:
                self.links.append(value)


root = Path(sys.argv[1])
prefix = urlsplit(sys.argv[2]).path.rstrip("/") + "/"
errors = []
checked = 0
pages = list(root.rglob("*.html"))
if not pages:
    sys.exit(f"No generated HTML pages in {root}")
for page in pages:
    parser = Links()
    parser.feed(page.read_text(encoding="utf-8"))
    for link in parser.links:
        url = urlsplit(link)
        if url.scheme or url.netloc or not url.path:
            continue
        if url.path.startswith("/"):
            if not url.path.startswith(prefix):
                errors.append(f"{page}: link {link} is outside the site base path {prefix}")
                continue
            target = root / unquote(url.path.removeprefix(prefix))
        else:
            target = page.parent / unquote(url.path)
        if target.is_dir():
            target /= "index.html"
        if not target.is_file():
            errors.append(f"{page}: broken link {link}")
        checked += 1
if errors:
    sys.exit("\n".join(errors))
print(f"Checked {checked} internal links across {len(pages)} HTML pages")
