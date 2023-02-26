const childProcess = require('child_process')
const os = require('os')
const process = require('process')

const BINARY_BASE = 'pr-size-labeler'

function chooseBinary() {
    const platform = os.platform()
    const arch = os.arch()

    if (platform === 'linux' && arch === 'x64') {
        return `${BINARY_BASE}-linux-amd64`
    }
    if (platform === 'linux' && arch === 'arm64') {
        return `${BINARY_NAME}-linux-arm64`
    }
    if (platform === 'windows' && arch === 'x64') {
        return `${BINARY_NAME}-windows-amd64`
    }
    if (platform === 'windows' && arch === 'arm64') {
        return `${BINARY_NAME}-windows-arm64`
    }

    console.error(`Unsupported platform (${platform}) and architecture (${arch})`)
    process.exit(1)
}

function main() {
    const binary = chooseBinary()
    const mainScript = `${__dirname}/bin/${binary}`
    const spawnSyncReturns = childProcess.spawnSync(mainScript, { stdio: 'inherit' })
    const status = spawnSyncReturns.status
    if (typeof status === 'number') {
        process.exit(status)
    }
    process.exit(1)
}

if (require.main === module) {
    main()
}
