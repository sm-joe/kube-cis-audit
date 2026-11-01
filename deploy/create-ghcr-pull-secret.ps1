[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [string]$GitHubUsername,

    [ValidateNotNullOrEmpty()]
    [string]$Namespace = "kube-system",

    [ValidateNotNullOrEmpty()]
    [string]$SecretName = "ghcr-pull-secret"
)

$secureToken = Read-Host "Enter GitHub PAT (classic) with read:packages" -AsSecureString
$tokenPointer = [IntPtr]::Zero
$plainToken = $null
$dockerConfigJson = $null
$dockerConfigBase64 = $null

try {
    $tokenPointer = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secureToken)
    $plainToken = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($tokenPointer)

    $auth = [Convert]::ToBase64String(
        [Text.Encoding]::UTF8.GetBytes("${GitHubUsername}:$plainToken")
    )

    $dockerConfig = @{
        auths = @{
            "ghcr.io" = @{
                username = $GitHubUsername
                password = $plainToken
                auth = $auth
            }
        }
    }
    $dockerConfigJson = $dockerConfig | ConvertTo-Json -Depth 5 -Compress
    $dockerConfigBase64 = [Convert]::ToBase64String(
        [Text.Encoding]::UTF8.GetBytes($dockerConfigJson)
    )

    $manifest = @"
apiVersion: v1
kind: Secret
metadata:
  name: $SecretName
  namespace: $Namespace
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: $dockerConfigBase64
"@

    $manifest | & kubectl apply -f -
    if ($LASTEXITCODE -ne 0) {
        throw "kubectl apply failed while creating the GHCR image pull secret."
    }

    Write-Host "Created/updated Secret '$SecretName' in namespace '$Namespace'."
    Write-Host "The token was not written to a file or included in kubectl command-line arguments."
}
finally {
    if ($tokenPointer -ne [IntPtr]::Zero) {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($tokenPointer)
    }
    $plainToken = $null
    $dockerConfigJson = $null
    $dockerConfigBase64 = $null
    $manifest = $null
    $secureToken = $null
}
