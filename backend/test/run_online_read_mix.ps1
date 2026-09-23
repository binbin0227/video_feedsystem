param(
    [ValidateRange(1, 500)]
    [int]$Rps = 5,

    [string]$BaseUrl = 'http://47.80.23.81'
)

$reportDirectory = Join-Path $PSScriptRoot 'reports'
New-Item -ItemType Directory -Force -Path $reportDirectory | Out-Null

$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$reportPath = Join-Path $reportDirectory "online-read-mix-$($Rps)rps-$timestamp.json"
$testScript = Join-Path $PSScriptRoot 'performance\online_read_mix_test.js'

Write-Host "开始测试：$Rps RPS，持续 60 秒"
Write-Host "目标地址：$BaseUrl"
Write-Host "结果文件：$reportPath"
Write-Host '代理设置：已为本轮 k6 测试禁用'

$previousHttpProxy = $env:HTTP_PROXY
$previousHttpsProxy = $env:HTTPS_PROXY
$previousAllProxy = $env:ALL_PROXY
$previousNoProxy = $env:NO_PROXY

try {
    $env:HTTP_PROXY = $null
    $env:HTTPS_PROXY = $null
    $env:ALL_PROXY = $null
    $env:NO_PROXY = ([Uri]$BaseUrl).Host

    & k6 run `
        -e "BASE_URL=$BaseUrl" `
        -e "TARGET_RPS=$Rps" `
        -e 'DURATION=60s' `
        --summary-export $reportPath `
        $testScript

    $testExitCode = $LASTEXITCODE
} finally {
    $env:HTTP_PROXY = $previousHttpProxy
    $env:HTTPS_PROXY = $previousHttpsProxy
    $env:ALL_PROXY = $previousAllProxy
    $env:NO_PROXY = $previousNoProxy
}

exit $testExitCode
