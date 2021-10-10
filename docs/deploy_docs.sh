#!/usr/bin/env sh

# abort on errors
set -e

python3 build_md_docs.py

# navigate into the vuepress project directory
cd vuepress_docs
# build
npm run docs:build

# navigate into the build output directory
cd docs/.vuepress/dist

git init
git add -A
git commit -m 'Deploy vuepress docs'

# if you are deploying to https://<USERNAME>.github.io/<REPO>
git push -f git@github.com:art9studio/Aureole.git master:docs

cd -