package tldraw

import _ "embed"

//go:generate bash -c "bun run build && cp dist/main.js ../statics/tldraw.min.js && cp dist/main.css ../statics/tldraw.min.css"
