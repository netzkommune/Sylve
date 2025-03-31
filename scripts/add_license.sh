#!/bin/sh

addlicense -f docs/CODE_LICENSE \
  -ignore internal/assets/web-files/** \
  cmd/** internal/** web/src/lib/**/*.ts web/src/lib/*.ts pkg/utils/** pkg/crypto/** pkg/disk/**
