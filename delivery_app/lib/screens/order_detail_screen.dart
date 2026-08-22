import 'dart:async';
import 'package:flutter/material.dart';
import '../services/api_service.dart';

class OrderDetailScreen extends StatefulWidget {
  final Map<String, dynamic> order;
  const OrderDetailScreen({super.key, required this.order});

  @override
  State<OrderDetailScreen> createState() => _OrderDetailScreenState();
}

class _OrderDetailScreenState extends State<OrderDetailScreen> {
  late Map<String, dynamic> _order;
  bool _loading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _order = widget.order;
  }

  Future<void> _markShipped() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await ApiService.markShipped(_order['order_id']);
      setState(() {
        _order = data['order'];
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString().replaceFirst('Exception: ', '');
        _loading = false;
      });
    }
  }

  Future<void> _advanceDeliveryStatus(String status, {String? otp}) async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await ApiService.updateDeliveryStatus(_order['order_id'], status, otp: otp);
      setState(() {
        _order = data['order'];
      });
      if (status == 'delivered') {
        final confirmData = await ApiService.confirmDelivery(_order['order_id']);
        setState(() => _order = confirmData['order']);
      }
      setState(() => _loading = false);
    } catch (e) {
      setState(() {
        _error = e.toString().replaceFirst('Exception: ', '');
        _loading = false;
      });
    }
  }

  Future<void> _promptForOtpAndDeliver() async {
    final otpController = TextEditingController();
    final otp = await showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Enter delivery OTP'),
        content: TextField(
          controller: otpController,
          keyboardType: TextInputType.number,
          maxLength: 6,
          decoration: const InputDecoration(hintText: 'Ask the customer for their OTP'),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
          TextButton(
            onPressed: () => Navigator.pop(ctx, otpController.text.trim()),
            child: const Text('Verify & Deliver'),
          ),
        ],
      ),
    );
    if (otp == null || otp.isEmpty) return;
    await _advanceDeliveryStatus('delivered', otp: otp);
  }

  Future<void> _acceptAssignment() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await ApiService.acceptAssignment(_order['order_id']);
      setState(() {
        _order = data['order'];
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString().replaceFirst('Exception: ', '');
        _loading = false;
      });
    }
  }

  Future<void> _rejectAssignment() async {
    final reasonController = TextEditingController();
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Reject this delivery?'),
        content: TextField(
          controller: reasonController,
          decoration: const InputDecoration(hintText: 'Reason (optional)'),
          maxLines: 2,
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Reject', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await ApiService.rejectAssignment(_order['order_id'], reason: reasonController.text);
      setState(() {
        _order = data['order'];
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString().replaceFirst('Exception: ', '');
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final status = _order['status'] ?? '';
    final assignmentStatus = _order['assignment_status'];
    final rejectionReason = _order['rejection_reason'];
    final expiresAt = _order['assignment_expires_at'];
    final deliveryAddress = _order['delivery_address'] ?? '';
    final customerName = _order['customer_name'] ?? '';
    final customerPhone = _order['customer_phone'] ?? '';
    final itemCount = _order['item_count'] ?? 0;

    final isPendingResponse = assignmentStatus == 'assigned';
    final isRejected = assignmentStatus == 'rejected';
    final isExpired = assignmentStatus == 'expired';
    final canProgressOrder = assignmentStatus == null || assignmentStatus == 'accepted';

    return Scaffold(
      appBar: AppBar(title: Text('Order #${_order['order_id']}')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (isPendingResponse)
            Container(
              margin: const EdgeInsets.only(bottom: 12),
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: const Color(0xFFFFF4E5),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.orange.shade200),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.timer_outlined, color: Colors.orange, size: 20),
                      const SizedBox(width: 8),
                      const Text(
                        'New delivery request',
                        style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                      ),
                      const Spacer(),
                      _CountdownText(expiresAt: expiresAt?.toString()),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton(
                          onPressed: _loading ? null : _rejectAssignment,
                          style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.red,
                            side: const BorderSide(color: Colors.red),
                            padding: const EdgeInsets.symmetric(vertical: 12),
                          ),
                          child: const Text('Reject'),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: ElevatedButton(
                          onPressed: _loading ? null : _acceptAssignment,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.green,
                            padding: const EdgeInsets.symmetric(vertical: 12),
                          ),
                          child: _loading
                              ? const SizedBox(
                                  height: 18,
                                  width: 18,
                                  child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                                )
                              : const Text('Accept'),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          if (isRejected)
            Container(
              margin: const EdgeInsets.only(bottom: 12),
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFFDEAEA),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                rejectionReason != null && rejectionReason.toString().isNotEmpty
                    ? 'You rejected this delivery: $rejectionReason'
                    : 'You rejected this delivery.',
                style: const TextStyle(color: Colors.red),
              ),
            ),
          if (isExpired)
            Container(
              margin: const EdgeInsets.only(bottom: 12),
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFF0F0F0),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Text(
                'This assignment expired without a response.',
                style: TextStyle(color: Colors.black54),
              ),
            ),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    customerName,
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                  const SizedBox(height: 4),
                  Text(customerPhone),
                  const SizedBox(height: 8),
                  Text(deliveryAddress),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Items', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                  const Divider(),
                  Text('$itemCount item(s)'),
                  const Divider(),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text('Total', style: TextStyle(fontWeight: FontWeight.bold)),
                      Text(
                        '\u20B9${_order['total_amount']}',
                        style: const TextStyle(fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                  Text('Payment: ${_order['payment_method']?.toString().toUpperCase() ?? ''}'),
                ],
              ),
            ),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(_error!, style: const TextStyle(color: Colors.red)),
          ],
          const SizedBox(height: 20),
          if (canProgressOrder && status == 'confirmed')
            ElevatedButton(
              onPressed: _loading ? null : _markShipped,
              style: ElevatedButton.styleFrom(padding: const EdgeInsets.all(16)),
              child: _loading
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('Mark as Shipped (Picked Up)'),
            )
          else if (canProgressOrder && status == 'shipped')
            Builder(builder: (context) {
              final deliveryStatus = _order['delivery_status'];
              switch (deliveryStatus) {
                case 'picked_up':
                  return ElevatedButton(
                    onPressed: _loading ? null : () => _advanceDeliveryStatus('out_for_delivery'),
                    style: ElevatedButton.styleFrom(padding: const EdgeInsets.all(16)),
                    child: _loading
                        ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Text('Out for Delivery'),
                  );
                case 'out_for_delivery':
                  return ElevatedButton(
                    onPressed: _loading ? null : () => _advanceDeliveryStatus('arrived'),
                    style: ElevatedButton.styleFrom(padding: const EdgeInsets.all(16)),
                    child: _loading
                        ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Text('Arrived at Location'),
                  );
                case 'arrived':
                  return ElevatedButton(
                    onPressed: _loading ? null : _promptForOtpAndDeliver,
                    style: ElevatedButton.styleFrom(padding: const EdgeInsets.all(16), backgroundColor: Colors.green),
                    child: _loading
                        ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Text('Enter OTP & Confirm Delivery'),
                  );
                default:
                  return ElevatedButton(
                    onPressed: _loading ? null : () => _advanceDeliveryStatus('picked_up'),
                    style: ElevatedButton.styleFrom(padding: const EdgeInsets.all(16)),
                    child: _loading
                        ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Text('Mark Picked Up'),
                  );
              }
            })
          else if (status == 'delivered')
            const Center(
              child: Text('Delivered', style: TextStyle(color: Colors.green, fontSize: 18)),
            ),
        ],
      ),
    );
  }
}

class _CountdownText extends StatefulWidget {
  final String? expiresAt;
  const _CountdownText({required this.expiresAt});

  @override
  State<_CountdownText> createState() => _CountdownTextState();
}

class _CountdownTextState extends State<_CountdownText> {
  Timer? _timer;
  Duration? _remaining;

  @override
  void initState() {
    super.initState();
    _tick();
    _timer = Timer.periodic(const Duration(seconds: 1), (_) => _tick());
  }

  void _tick() {
    if (widget.expiresAt == null) {
      if (mounted) setState(() => _remaining = null);
      return;
    }
    try {
      final expiry = DateTime.parse(widget.expiresAt!).toLocal();
      final diff = expiry.difference(DateTime.now());
      if (mounted) setState(() => _remaining = diff.isNegative ? Duration.zero : diff);
    } catch (_) {
      if (mounted) setState(() => _remaining = null);
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_remaining == null) {
      return const Text(
        'Awaiting response',
        style: TextStyle(fontSize: 12, color: Colors.orange, fontWeight: FontWeight.w600),
      );
    }
    final m = _remaining!.inMinutes;
    final s = _remaining!.inSeconds % 60;
    final label = _remaining! == Duration.zero ? 'Expiring...' : '${m}m ${s.toString().padLeft(2, '0')}s left';
    return Text(
      label,
      style: TextStyle(
        fontSize: 12,
        fontWeight: FontWeight.w700,
        color: _remaining!.inSeconds < 30 ? Colors.red : Colors.orange,
      ),
    );
  }
}
