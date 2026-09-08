$file = "lib\screens\order_detail_screen.dart"
$content = Get-Content $file -Raw
$idxStart = $content.IndexOf("Widget _buildStepper(String? deliveryStatus, String orderStatus) {")
$idxEnd = $content.IndexOf("@override`r`n  Widget build(BuildContext context) {")

$newFunc = @"
Widget _buildStepper(String? deliveryStatus, String orderStatus) {
    final currentIndex = _currentStepIndex(deliveryStatus, orderStatus);
    if (currentIndex < 0) return const SizedBox.shrink();

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: List.generate(_stepDefs.length, (stepIndex) {
          final isDone = stepIndex < currentIndex;
          final isCurrent = stepIndex == currentIndex;
          final isLast = stepIndex == _stepDefs.length - 1;
          final label = _stepDefs[stepIndex]['label']!;
          final lineDone = stepIndex < currentIndex;

          return IntrinsicHeight(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Column(
                  children: [
                    Container(
                      width: 26,
                      height: 26,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: isDone || isCurrent ? const Color(0xFF5B2A9E) : Colors.white,
                        border: Border.all(
                          color: isDone || isCurrent ? const Color(0xFF5B2A9E) : const Color(0xFFE0DAF0),
                          width: 2,
                        ),
                      ),
                      child: isDone
                          ? const Icon(Icons.check, size: 14, color: Colors.white)
                          : isCurrent
                              ? Center(
                                  child: Container(
                                    width: 8,
                                    height: 8,
                                    decoration: const BoxDecoration(shape: BoxShape.circle, color: Colors.white),
                                  ),
                                )
                              : null,
                    ),
                    if (!isLast)
                      Expanded(
                        child: Container(
                          width: 2,
                          margin: const EdgeInsets.symmetric(vertical: 2),
                          color: lineDone ? const Color(0xFF5B2A9E) : const Color(0xFFE0DAF0),
                        ),
                      ),
                  ],
                ),
                const SizedBox(width: 12),
                Padding(
                  padding: const EdgeInsets.only(bottom: 20, top: 3),
                  child: Text(
                    label,
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: isCurrent ? FontWeight.bold : FontWeight.normal,
                      color: isDone || isCurrent ? Colors.black87 : Colors.black45,
                    ),
                  ),
                ),
              ],
            ),
          );
        }),
      ),
    );
  }

  
"@

$before = $content.Substring(0, $idxStart)
$after = $content.Substring($idxEnd)
$updated = $before + $newFunc + $after
[System.IO.File]::WriteAllText("$PWD\$file", $updated, (New-Object System.Text.UTF8Encoding $false))
Write-Host "SUCCESS"
