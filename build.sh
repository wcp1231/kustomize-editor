#!/bin/sh

html=$(minify --html-keep-document-tags --type html index.html | base64 | sed 's/[\/&]/\\&/g')

sed -i '' -e "s/<html>/$html/g" main.go

go build -o kustomize-editor .

git checkout -f