#!/bin/bash

# Copy all files (after removing them)
rm -rf ../RPout-Github/"(.git|node_modules)"
rsync -ax --exclude='.git' --exclude='node_modules' . ../RPout-Github/

# Replace all "git.rpjosh.de" references with Github URL
#find ../RPout-Github/ -type f -name '*' -exec sed -i 's/git.rpjosh.de\/RPJosh\/RPout/github.com\/RPJoshL\/RPout/g' {} \;
find ../RPout-Github/ -type f -name '*' -exec sed -i 's/git.rpjosh.de\/RPJosh\/go-logger/github.com\/RPJoshL\/go-logger/g' {} \;
