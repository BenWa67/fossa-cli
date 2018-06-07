﻿#Requires -Version 5

<#
    .SYNOPSIS
    Download and install the latest available FOSSA release from GitHub.
#>

$ErrorActionPreference = "Stop"

$github = "https://github.com"
$latestUri = "$github/fossas/fossa-cli/releases/latest"
$extractDir = "$env:ALLUSERSPROFILE\fossa-cli"

Write-Verbose "Looking up latest release..."

[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12

$releasePage = Invoke-RestMethod $latestUri

if ($releasePage -inotmatch 'href=\"(.*?releases\/download\/.*?windows.*?)\"')
{
    throw "Did not find latest Windows release at $latestUri"
}

$downloadUri = "$github/$($Matches[1])"
Write-Verbose "Downloading from: $downloadUri"

$TempDir = Join-Path [System.IO.Path]::GetTempPath() "fossa-cli"
if (![System.IO.Directory]::Exists($TempDir)) {[void][System.IO.Directory]::CreateDirectory($TempDir)}

$zipFile = "$TempDir\fossa-cli.zip"

(New-Object System.Net.WebClient).DownloadFile($downloadUri, $zipFile)

Expand-Archive -Path $zipFile -DestinationPath $extractDir -Force

$ErrorActionPreference = "Continue"

Write-Verbose "Installed fossa-cli at: $extractDir\fossa-cli\fossa.exe"
Write-Verbose "Get started by running: fossa.exe --help"