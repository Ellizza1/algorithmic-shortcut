package web

import "embed"

// Assets содержит встроенные статические ресурсы веб-интерфейса
//
//go:embed index.html style.css app.js
var Assets embed.FS
