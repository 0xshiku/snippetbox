package ui

import "embed"

// This is an important line. It looks like a comment, but it is actually a special comment directive
// This comment directive instructs Go to store the files from our ui/html and ui/static folders in an embed.FS
// embedded filesystem referenced by the global variable Files
//
//go:embed "html" "static"
var Files embed.FS
