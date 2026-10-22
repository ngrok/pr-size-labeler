const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const test = require('node:test')
const vm = require('node:vm')

const source = fs.readFileSync(path.join(__dirname, 'invoke-binary.js'), 'utf8')

for (const [platform, arch, binary] of [
    ['linux', 'x64', 'pr-size-labeler-linux-amd64'],
    ['linux', 'arm64', 'pr-size-labeler-linux-arm64'],
    ['win32', 'x64', 'pr-size-labeler-windows-amd64.exe'],
    ['win32', 'arm64', 'pr-size-labeler-windows-arm64.exe'],
]) {
    test(`launches ${binary}`, () => {
        const exitSignal = new Error('process exited')
        const mockRequire = (name) => {
            if (name === 'os') {
                return { platform: () => platform, arch: () => arch }
            }
            if (name === 'child_process') {
                return {
                    spawnSync: (filename, options) => {
                        assert.equal(filename, `${__dirname}/bin/${binary}`)
                        assert.equal(options.stdio, 'inherit')
                        return { status: 0 }
                    },
                }
            }
            if (name === 'process') {
                return {
                    exit: (status) => {
                        assert.equal(status, 0)
                        throw exitSignal
                    },
                }
            }
            throw new Error(`Unexpected module: ${name}`)
        }
        const module = {}
        mockRequire.main = module

        assert.throws(() => vm.runInNewContext(source, {
            require: mockRequire,
            module,
            __dirname,
        }), (error) => error === exitSignal)
    })
}