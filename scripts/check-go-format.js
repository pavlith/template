import { execFileSync } from 'node:child_process'

const unformatted = execFileSync('gofmt', ['-l', '.'], {
  encoding: 'utf8',
}).trim()
if (unformatted) {
  console.error(unformatted)
  process.exitCode = 1
}
