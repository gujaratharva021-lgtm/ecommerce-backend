$file = "lib\screens\order_detail_screen.dart"
$content = Get-Content $file -Raw
$searchStr = "default:`r`n                  return ElevatedButton(`r`n                    onPressed: _loading ? null : () => _advanceDeliveryStatus('going_to_store'),"
$idx = $content.IndexOf($searchStr)

$insertText = @"
case 'returned':
                  return const Center(
                    child: Padding(
                      padding: EdgeInsets.symmetric(vertical: 16),
                      child: Text('Returned to store', style: TextStyle(color: Colors.orange, fontSize: 18, fontWeight: FontWeight.bold)),
                    ),
                  );
                
"@

$before = $content.Substring(0, $idx)
$after = $content.Substring($idx)
$updated = $before + $insertText + $after

[System.IO.File]::WriteAllText("$PWD\$file", $updated, (New-Object System.Text.UTF8Encoding $false))
Write-Host "SUCCESS"
