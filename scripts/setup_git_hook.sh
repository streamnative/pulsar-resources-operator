#!/usr/bin/bash
# Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

# Install husky locally
npm install husky -D
# Enable git hooks
npm set-script prepare "husky install"
npm run prepare

# # Install commitlint locally
# npm install --save-dev @commitlint/{config-conventional,cli}
# echo "module.exports = {extends: ['@commitlint/config-conventional']}" > commitlint.config.js
# # Add commitlint for commit-msg hook. 
# # This hook will check commit message,if you want to bypass the check, use --no-verify options
# # eg: git commit -m "hell" --no-verify
# npx husky add .husky/commit-msg 'npx --no -- commitlint --edit "$1"'


# Add pre-commit hook to check license
npx husky add .husky/pre-commit 'echo "Check License"'
npx husky add .husky/pre-commit 'go test license_test.go'

# Add pre-commit hook to lint code
npx husky add .husky/pre-commit 'echo "Check Lint"'
npx husky add .husky/pre-commit './scripts/lint.sh ./...'

# Add pre-commit hook to fmt code
npx husky add .husky/pre-commit 'echo "Go fmt"'
npx husky add .husky/pre-commit './scripts/verify_gofmt.sh ./...'

# Add pre-commit hook to go vet code
npx husky add .husky/pre-commit 'echo "Go vet"'
npx husky add .husky/pre-commit './scripts/verify_govet.sh ./...'