#!/bin/sh

cp ./git-hooks/pre-commit.sh .git/hooks/pre-commit
cp ./git-hooks/pre-push.sh .git/hooks/pre-push
chmod 755 .git/hooks/pre-commit
chmod 755 .git/hooks/pre-push