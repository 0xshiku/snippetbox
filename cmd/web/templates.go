package main

import "github.com/0xshiku/snippetbox/internal/models"

// Define a templateData type to act as the holding structure for any dynamic data that we want to pass to our HTML templates
type templateData struct {
	Snippet *models.Snippet
}
