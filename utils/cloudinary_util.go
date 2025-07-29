package utils

import (
	"path"
	"regexp"
	"strings"
)

func ExtractPublicIDFromURL(url string) string {
	parts := strings.Split(url, "/upload/")
	if len(parts) != 2 {
		return ""
	}
	
	pathParts := strings.Split(parts[1], "/")
	if len(pathParts) < 2 {
		return ""
	}

	publicPath := strings.Join(pathParts[1:], "/") // [go-chat-app, profiles, filename.jpg]
	publicPath = strings.TrimSuffix(publicPath, path.Ext(publicPath)) // remove .jpg

	return publicPath // go-chat-app/profiles/filename
}


func SanitizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile("[^a-zA-Z0-9_-]+")
	return reg.ReplaceAllString(name, "")
}