#!/bin/sh

html=$(cat index.html | sed -e 's/type="text\/x-template"/ /g' | minify --html-keep-document-tags --type html -- | base64 | sed 's/[\/&]/\\&/g')

sed -i '' -e "s/<html>/$html/g" main.go

go build -o kustomize-editor .

git checkout -f