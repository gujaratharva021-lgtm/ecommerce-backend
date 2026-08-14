$categoryGST = @{
    1  = 0
    2  = 18
    3  = 18
    4  = 18
    5  = 18
    6  = 18
    7  = 18
    8  = 18
    9  = 12
    10 = 18
    11 = 18
    12 = 18
    13 = 12
    14 = 0
    15 = 12
    16 = 5
    33 = 5
    34 = 5
    35 = 5
    36 = 12
    37 = 18
    38 = 0
    39 = 5
    40 = 12
    41 = 18
    42 = 18
    43 = 0
    44 = 18
    45 = 12
    46 = 18
    47 = 18
}

$overrides = @{
    273 = 18   # Camphor
    274 = 5    # Agarbatti
    271 = 12   # Bell (Ghanti)
    272 = 12   # Brass Pooja Thali
}

$allProducts = @()
$page = 1
do {
    $resp = Invoke-RestMethod -Uri "$base/api/v1/products?limit=100&page=$page" -Method GET -Headers $headers
    $allProducts += $resp.products
    $page++
} while ($allProducts.Count -lt $resp.total)

Write-Host "Fetched $($allProducts.Count) products" -ForegroundColor Cyan

$successCount = 0
$failCount = 0
$failedIds = @()

foreach ($p in $allProducts) {
    if ($overrides.ContainsKey($p.id)) {
        $gst = $overrides[$p.id]
    } else {
        $gst = $categoryGST[$p.category_id]
        if ($null -eq $gst) { $gst = 0 }
    }

    $updateBody = @{
        name        = $p.name
        description = $p.description
        price       = $p.price
        gst_percent = $gst
        image_url   = $p.image_url
        category_id = $p.category_id
        stock       = 0
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "$base/api/v1/admin/products/$($p.id)" -Method PUT -Headers $headers -Body $updateBody -ContentType "application/json" -TimeoutSec 30 | Out-Null
        $successCount++
    } catch {
        $failCount++
        $failedIds += $p.id
        Write-Host "Failed ID $($p.id): $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "Done. Success: $successCount, Failed: $failCount" -ForegroundColor Green
if ($failCount -gt 0) {
    Write-Host "Failed product IDs: $($failedIds -join ', ')" -ForegroundColor Red
}