# The Stem - Windows installation script
# Run as Administrator from an extracted Stem Windows archive.
#
# Installs stem.exe to Program Files\Stem, registers the Stem service, opens
# the firewall for the web UI and starts it. Safe to run again over an
# existing install: configuration, credentials, certificates and data under
# %ProgramData%\stem are never overwritten.
#
# Keep this file ASCII: Windows PowerShell 5.1 reads a BOM-less script as the
# ANSI code page and mangles anything else.

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$InstallDir = Join-Path $env:ProgramFiles "Stem"
$DataDir = Join-Path $env:ProgramData "stem"
$BinaryName = "stem.exe"
$ServiceName = "Stem"
$Port = 8444
$FirewallRule = "Stem Web UI"
$SourceBinary = Join-Path $ScriptDir $BinaryName
$InstalledBinary = Join-Path $InstallDir $BinaryName

$Principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $Principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Error "Stem installation requires an elevated PowerShell session. Reopen PowerShell as Administrator and run install.ps1 again."
    exit 1
}

if (-not (Test-Path $SourceBinary)) {
    Write-Error "Cannot find $BinaryName next to install.ps1. Extract the complete archive before installing."
    exit 1
}

Write-Host "Installing Stem..." -ForegroundColor Green

# Judge the data directory before anything on the machine is changed: a refusal
# must leave the running service and the installed binary alone.
# The data directory holds the credentials, private key, license and
# configuration the LocalSystem service reads, so whoever can write it or
# replace anything in it can run code as SYSTEM. This script fails closed and
# never sanitises a tree it cannot trust: there is no recursive re-owning or
# ACL reset. It refuses, listing what it found, when the directory or anything
# under it is owned by anyone but SYSTEM, Administrators or TrustedInstaller,
# is a symbolic link or junction, carries an access entry of its own for an
# account that is not SYSTEM or Administrators, or when the directory's own
# DACL lets such an account write, delete or change permissions. Otherwise it
# replaces the directory's DACL wholesale (owner Administrators, inheritance
# cut, SYSTEM and Administrators full control, nothing else), lets everything
# inside inherit that, and reads the result back.
$SystemSid = New-Object Security.Principal.SecurityIdentifier("S-1-5-18")
$AdminsSid = New-Object Security.Principal.SecurityIdentifier("S-1-5-32-544")
$InstallerSid = New-Object Security.Principal.SecurityIdentifier("S-1-5-80-956008885-3418522649-1831038044-1853292631-2271478464")
$WriteRights = [Security.AccessControl.FileSystemRights]::WriteData -bor
    [Security.AccessControl.FileSystemRights]::AppendData -bor
    [Security.AccessControl.FileSystemRights]::WriteExtendedAttributes -bor
    [Security.AccessControl.FileSystemRights]::WriteAttributes -bor
    [Security.AccessControl.FileSystemRights]::Delete -bor
    [Security.AccessControl.FileSystemRights]::DeleteSubdirectoriesAndFiles -bor
    [Security.AccessControl.FileSystemRights]::ChangePermissions -bor
    [Security.AccessControl.FileSystemRights]::TakeOwnership

function Test-TrustedOwner($Sid) {
    return ($Sid -eq $SystemSid) -or ($Sid -eq $AdminsSid) -or ($Sid -eq $InstallerSid)
}

function Test-TrustedGrantee($Sid) {
    return ($Sid -eq $SystemSid) -or ($Sid -eq $AdminsSid)
}

# An account name when the SID resolves to one, else the SID, for messages.
function Get-AccountName($Sid) {
    try {
        return $Sid.Translate([Security.Principal.NTAccount]).Value
    } catch {
        return $Sid.Value
    }
}

function Get-RuleSid($Rule) {
    return $Rule.IdentityReference.Translate([Security.Principal.SecurityIdentifier])
}

# Everything wrong with $Root and what is under it. It never descends into a
# symbolic link or junction, so a planted link cannot lead the walk out of the
# directory. $Root's own DACL is judged separately, by what it lets others do.
function Get-DataDirProblems($Root) {
    $Problems = New-Object System.Collections.Generic.List[string]
    $Pending = New-Object System.Collections.Generic.Stack[string]
    $Pending.Push($Root)
    $First = $true
    while ($Pending.Count -gt 0) {
        $Path = $Pending.Pop()
        $Item = Get-Item -LiteralPath $Path -Force
        if ($Item.Attributes -band [IO.FileAttributes]::ReparsePoint) {
            $Problems.Add("$Path is a symbolic link or junction")
            continue
        }
        $Acl = Get-Acl -LiteralPath $Path
        $Owner = $Acl.GetOwner([Security.Principal.SecurityIdentifier])
        if (-not (Test-TrustedOwner $Owner)) {
            $Problems.Add("$Path is owned by $(Get-AccountName $Owner)")
        }
        foreach ($Rule in $Acl.Access) {
            if ($Rule.AccessControlType -ne "Allow") { continue }
            $RuleSid = Get-RuleSid $Rule
            if (Test-TrustedGrantee $RuleSid) { continue }
            if ($First -and ($Rule.FileSystemRights -band $WriteRights)) {
                $Problems.Add("$Path grants write access to $($Rule.IdentityReference)")
            } elseif (-not $First -and -not $Rule.IsInherited) {
                $Problems.Add("$Path has its own access entry for $($Rule.IdentityReference)")
            }
        }
        $First = $false
        if ($Item.PSIsContainer) {
            foreach ($Child in Get-ChildItem -LiteralPath $Path -Force) {
                $Pending.Push($Child.FullName)
            }
        }
    }
    return ,$Problems
}

function Stop-DataDirRefused($Problems) {
    Write-Error ("Refusing to use $DataDir, which cannot be trusted:`n  " + ($Problems -join "`n  ") +
        "`nInspect it, remove what is not yours, and run install.ps1 again.")
    exit 1
}

$Security = New-Object Security.AccessControl.DirectorySecurity
$Security.SetOwner($AdminsSid)
$Security.SetAccessRuleProtection($true, $false)
foreach ($Sid in @($SystemSid, $AdminsSid)) {
    $Security.AddAccessRule((New-Object Security.AccessControl.FileSystemAccessRule(
        $Sid, "FullControl", "ContainerInherit, ObjectInherit", "None", "Allow")))
}

if (Test-Path -LiteralPath $DataDir) {
    if (-not (Get-Item -LiteralPath $DataDir -Force).PSIsContainer) {
        Stop-DataDirRefused @("$DataDir is not a directory")
    }
    $Problems = Get-DataDirProblems $DataDir
    if ($Problems.Count -gt 0) {
        Stop-DataDirRefused $Problems
    }
    Set-Acl -LiteralPath $DataDir -AclObject $Security
} elseif ($PSVersionTable.PSEdition -eq "Core") {
    [IO.FileSystemAclExtensions]::Create((New-Object IO.DirectoryInfo($DataDir)), $Security)
} else {
    [IO.Directory]::CreateDirectory($DataDir, $Security) | Out-Null
}

# Look again, and read the result back: what was planted between the first
# look and the new DACL is caught here, and the directory must hold exactly
# SYSTEM and Administrators.
$Problems = Get-DataDirProblems $DataDir
$Top = Get-Acl -LiteralPath $DataDir
$TopRules = @($Top.Access)
if ($Top.GetOwner([Security.Principal.SecurityIdentifier]) -ne $AdminsSid) {
    $Problems.Add("$DataDir is not owned by Administrators")
}
if (-not $Top.AreAccessRulesProtected) {
    $Problems.Add("$DataDir inherits permissions from its parent")
}
if ($TopRules.Count -ne 2) {
    $Problems.Add("$DataDir has $($TopRules.Count) access entries, want exactly SYSTEM and Administrators")
}
foreach ($Rule in $TopRules) {
    if (-not (Test-TrustedGrantee (Get-RuleSid $Rule))) {
        $Problems.Add("$DataDir grants access to $($Rule.IdentityReference)")
    }
}
if ($Problems.Count -gt 0) {
    Stop-DataDirRefused $Problems
}

$ExistingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($null -ne $ExistingService -and $ExistingService.Status -ne "Stopped") {
    Write-Host "  Stopping the running service..."
    Stop-Service -Name $ServiceName -Force
    $ExistingService.WaitForStatus("Stopped", [TimeSpan]::FromSeconds(30))
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item $SourceBinary $InstalledBinary -Force

# A sample configuration, laid down once and never replaced.
$SampleConfig = Join-Path $ScriptDir "stem.yaml"
$DataConfig = Join-Path $DataDir "stem.yaml"
if ((Test-Path $SampleConfig) -and -not (Test-Path $DataConfig)) {
    Copy-Item $SampleConfig $DataConfig
}

$MachinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if (($MachinePath -split ";") -notcontains $InstallDir) {
    [Environment]::SetEnvironmentVariable("Path", "$MachinePath;$InstallDir", "Machine")
    Write-Host "  Added $InstallDir to the system PATH (open a new terminal to use it)"
}

# Idempotent: registers the service, or points an existing one at this binary.
Write-Host "  Registering the Windows service..."
& $InstalledBinary service install
if ($LASTEXITCODE -ne 0) {
    $ExitCode = $LASTEXITCODE
    Write-Error "Stem service registration failed with exit code $ExitCode."
    exit $ExitCode
}

if ($null -eq (Get-NetFirewallRule -DisplayName $FirewallRule -ErrorAction SilentlyContinue)) {
    Write-Host "  Adding firewall rule for port $Port..."
    New-NetFirewallRule -DisplayName $FirewallRule `
        -Direction Inbound -Protocol TCP -LocalPort $Port `
        -Action Allow | Out-Null
}

Write-Host "  Starting the service..."
& $InstalledBinary service start
if ($LASTEXITCODE -ne 0) {
    $ExitCode = $LASTEXITCODE
    Write-Error "Stem service startup failed with exit code $ExitCode. See $DataDir\logs\stem.log."
    exit $ExitCode
}

# The service reports Running as soon as it is launched, before the daemon has
# bound its port. Wait for the port, and give up if the service dies first.
$Deadline = (Get-Date).AddSeconds(30)
$Listening = $false
while (-not $Listening -and (Get-Date) -lt $Deadline) {
    if ((Get-Service -Name $ServiceName).Status -ne "Running") {
        break
    }
    $Client = New-Object System.Net.Sockets.TcpClient
    try {
        $Client.Connect("127.0.0.1", $Port)
        $Listening = $true
    } catch {
        Start-Sleep -Milliseconds 500
    } finally {
        $Client.Close()
    }
}
if (-not $Listening) {
    Write-Error "The Stem service is not listening on port $Port. See $DataDir\logs\stem.log."
    exit 1
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host "  Web UI:  https://localhost:$Port"
Write-Host "  Service: $ServiceName"
Write-Host "  Data:    $DataDir"
Write-Host ""
Write-Host "Commands:"
Write-Host "  stem service status  - Check service status"
Write-Host "  stem service stop    - Stop service"
Write-Host "  stem service start   - Start service"
