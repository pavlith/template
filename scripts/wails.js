import { spawnSync } from 'node:child_process'

const result = spawnSync('go', ['tool', 'wails3', ...process.argv.slice(2)], {
  env: { ...process.env, WAILS_EXP_USE_WAKE: '1' },
  stdio: 'inherit',
})

if (result.error) {
  console.error(result.error.message)
}
process.exitCode = result.status ?? 1
