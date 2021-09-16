#!/usr/bin/env sh

# abort on errors
set -e

# navigate into the vuepress project directory
cd docs/vue_docs
# build
npm run build

# navigate into the build output directory
cd src/.vuepress/dist

git init
git add -A
git commit -m 'Deploy vuepress docs'

# if you are deploying to https://<USERNAME>.github.io/<REPO>
git push -f git@github.com:art9studio/Aureole.git master:docs

cd -