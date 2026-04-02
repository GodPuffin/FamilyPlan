package assets

import "embed"

// TemplatesFS contains the embedded HTML templates.
//
//go:embed templates/*.html
var TemplatesFS embed.FS

// StaticFS contains the embedded static assets.
//
//go:embed static/*
var StaticFS embed.FS
