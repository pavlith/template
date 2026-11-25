#!/usr/bin/env node

import { readFileSync, writeFileSync } from 'node:fs'
import { execFileSync } from 'node:child_process'
import { createInterface } from 'node:readline'

const SEMVER = /^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/

const FILES = {
  pkg: 'package.json',
  wails: 'wails.yaml',
}

function die(message) {
  console.error(`\n  ✗ ${message}\n`)
  process.exit(1)
}

function exec(file, args, opts = {}) {
  const display = [file, ...args].join(' ')
  console.log(`  $ ${display}`)
  execFileSync(file, args, { stdio: 'inherit', ...opts })
}

function ask(question) {
  const rl = createInterface({ input: process.stdin, output: process.stdout })
  return new Promise((resolve) => {
    rl.question(question, (answer) => {
      rl.close()
      resolve(answer.trim().toLowerCase())
    })
  })
}

// --- Parse version -----------------------------------------------------------

const raw = process.argv[2]
if (!raw) die('Usage: node scripts/prepare-release.js v1.0.0')

const match = raw.match(SEMVER)
if (!match) die(`"${raw}" is not a valid semver version (expected vX.Y.Z)`)

const version = `${match[1]}.${match[2]}.${match[3]}`
const tag = `v${version}`
console.log(`\n  Preparing release ${tag}\n`)

// --- Clean-tree gate ---------------------------------------------------------

const status = execFileSync('git', ['status', '--porcelain'], {
  encoding: 'utf-8',
}).trim()
if (status)
  die('Working tree is not clean. Commit or stash your changes first.')

// --- Quality gate ------------------------------------------------------------

console.log('\n  Running check:all…\n')
try {
  exec('pnpm', ['check:all'])
} catch {
  die('check:all failed — fix the errors before releasing.')
}

// --- Version sync ------------------------------------------------------------

console.log(`\n  Syncing version → ${version}\n`)

const pkg = JSON.parse(readFileSync(FILES.pkg, 'utf-8'))
pkg.version = version
writeFileSync(FILES.pkg, JSON.stringify(pkg, null, 2) + '\n')
console.log(`  ✓ ${FILES.pkg}`)

const manifest = readFileSync(FILES.wails, 'utf-8')
const versionField = /^ {2}version: ([^\r\n]+)$/m
if (!versionField.test(manifest))
  die(`Could not find project.version in ${FILES.wails}`)
writeFileSync(
  FILES.wails,
  manifest.replace(versionField, `  version: ${version}`),
)
console.log(`  ✓ ${FILES.wails}`)

// --- Prompt ------------------------------------------------------------------

console.log('\n  Ready to commit, tag, and push:\n')
console.log(`    git add -A`)
console.log(`    git commit -m "release: ${tag}"`)
console.log(`    git tag ${tag}`)
console.log(`    git push origin main --tags`)

const answer = await ask('\n  Execute these commands? [y/N] ')

if (answer === 'y' || answer === 'yes') {
  exec('git', ['add', '-A'])
  exec('git', ['commit', '-m', `release: ${tag}`])
  exec('git', ['tag', tag])
  exec('git', ['push', 'origin', 'main', '--tags'])
  console.log(`\n  ✓ ${tag} released\n`)
} else {
  console.log('\n  Skipped. Run the commands above manually when ready.\n')
}
