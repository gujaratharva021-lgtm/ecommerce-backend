import 'dart:async';
import 'package:flutter/material.dart';
import '../services/api_service.dart';
import 'notifications_screen.dart';
import 'order_detail_screen.dart';

class OrdersScreen extends StatefulWidget {
  final void Function(int)? onSwitchTab;
  const OrdersScreen({super.key, this.onSwitchTab});

  @override
  State<OrdersScreen> createState() => _OrdersScreenState();
}

class _OrdersScreenState extends State<OrdersScreen> {
  List<dynamic> _orders = [];
  bool _loading = true;
  String? _error;
  String _filter = 'active';
  final Set<int> _actingOrderIds = {};

  static const Color primaryPurple = Color(0xFF5B2A9E);
  static const Color pageBg = Color(0xFFF7F1FB);
  static const Color cardBg = Color(0xFFF3EDFA);

  @override
  void initState() {
    super.initState();
    _loadAll();
  }

  Future<void> _loadAll() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final orders = await ApiService.getMyDeliveries();
      setState(() {
        _orders = orders;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Failed to load orders';
        _loading = false;
      });
    }
  }

  Future<void> _acceptOrder(int orderId) async {
    setState(() => _actingOrderIds.add(orderId));
    try {
      await ApiService.acceptAssignment(orderId);
      await _loadAll();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString().replaceFirst('Exception: ', ''))),
        );
      }
    } finally {
      if (mounted) setState(() => _actingOrderIds.remove(orderId));
    }
  }

  Future<void> _rejectOrder(int orderId) async {
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
    setState(() => _actingOrderIds.add(orderId));
    try {
      await ApiService.rejectAssignment(orderId, reason: reasonController.text);
      await _loadAll();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString().replaceFirst('Exception: ', ''))),
        );
      }
    } finally {
      if (mounted) setState(() => _actingOrderIds.remove(orderId));
    }
  }

  List<dynamic> get _filteredOrders {
    if (_filter == 'active') {
      return _orders.where((o) => (o['status'] ?? '') != 'delivered' && (o['status'] ?? '') != 'cancelled').toList();
    }
    if (_filter == 'completed') {
      return _orders.where((o) => (o['status'] ?? '') == 'delivered').toList();
    }
    return _orders;
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'confirmed':
        return const Color(0xFF3B82F6);
      case 'shipped':
        return const Color(0xFFF59E0B);
      case 'delivered':
        return const Color(0xFF22C55E);
      default:
        return Colors.grey;
    }
  }

  IconData _statusIcon(String status) {
    switch (status) {
      case 'delivered':
        return Icons.check_circle_outline;
      case 'shipped':
        return Icons.local_shipping_outlined;
      default:
        return Icons.assignment_outlined;
    }
  }

  String? _formatTime(dynamic raw) {
    if (raw == null) return null;
    try {
      final dt = DateTime.parse(raw.toString()).toLocal();
      final now = DateTime.now();
      final isToday = dt.year == now.year && dt.month == now.month && dt.day == now.day;
      final hour = dt.hour % 12 == 0 ? 12 : dt.hour % 12;
      final min = dt.minute.toString().padLeft(2, '0');
      final ampm = dt.hour >= 12 ? 'PM' : 'AM';
      final time = '$hour:$min $ampm';
      return isToday ? 'Today, $time' : '${dt.day}/${dt.month}, $time';
    } catch (_) {
      return null;
    }
  }

  void _openDetail(dynamic order) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => OrderDetailScreen(order: order)),
    ).then((_) => _loadAll());
  }

  @override
  Widget build(BuildContext context) {
    final orders = _filteredOrders;
    return Scaffold(
      backgroundColor: pageBg,
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              child: Row(
                children: [
                  const Expanded(
                    child: Text(
                      'My Deliveries',
                      style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.black87),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.notifications_none, color: Colors.black87),
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(builder: (_) => const NotificationsScreen()),
                      );
                    },
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Row(
                children: [
                  _filterChip('New', 'new'),
                  const SizedBox(width: 8),
                  _filterChip('Active', 'active'),
                  const SizedBox(width: 8),
                  _filterChip('Completed', 'completed'),
                  const SizedBox(width: 8),
                  _filterChip('All', 'all'),
                ],
              ),
            ),
            Expanded(
              child: _loading
                  ? const Center(child: CircularProgressIndicator())
                  : _error != null
                      ? Center(child: Text(_error!))
                      : orders.isEmpty
                          ? const Center(child: Text('No deliveries here', style: TextStyle(color: Colors.black45)))
                          : RefreshIndicator(
                              onRefresh: _loadAll,
                              child: ListView.builder(
                                padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                                itemCount: orders.length,
                                itemBuilder: (context, index) {
                                  final o = orders[index];
                                  final orderId = o['order_id'] ?? o['id'];
                                  final status = (o['status'] ?? '').toString();
                                  final isAssignedPending = status == 'confirmed' &&
                                      (o['delivery_status'] == null || o['delivery_status'] == 'assigned');
                                  final isActing = _actingOrderIds.contains(orderId);

                                  return GestureDetector(
                                    onTap: () => _openDetail(o),
                                    child: Container(
                                      margin: const EdgeInsets.only(bottom: 10),
                                      padding: const EdgeInsets.all(14),
                                      decoration: BoxDecoration(color: cardBg, borderRadius: BorderRadius.circular(16)),
                                      child: Column(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Row(
                                            children: [
                                              Container(
                                                padding: const EdgeInsets.all(8),
                                                decoration: BoxDecoration(
                                                  color: Colors.white,
                                                  borderRadius: BorderRadius.circular(10),
                                                ),
                                                child: Icon(_statusIcon(status), size: 18, color: _statusColor(status)),
                                              ),
                                              const SizedBox(width: 10),
                                              Expanded(
                                                child: Column(
                                                  crossAxisAlignment: CrossAxisAlignment.start,
                                                  children: [
                                                    Text('Order #$orderId', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
                                                    if (_formatTime(o['created_at']) != null)
                                                      Text(
                                                        _formatTime(o['created_at'])!,
                                                        style: const TextStyle(fontSize: 11, color: Colors.black45),
                                                      ),
                                                  ],
                                                ),
                                              ),
                                              Container(
                                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                                decoration: BoxDecoration(
                                                  color: _statusColor(status).withValues(alpha: 0.12),
                                                  borderRadius: BorderRadius.circular(10),
                                                ),
                                                child: Text(
                                                  status.toUpperCase(),
                                                  style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: _statusColor(status)),
                                                ),
                                              ),
                                            ],
                                          ),
                                          if (o['total_amount'] != null) ...[
                                            const SizedBox(height: 8),
                                            Text(
                                              '₹${o['total_amount']} · ${(o['payment_method'] ?? '').toString().toUpperCase()}',
                                              style: const TextStyle(fontSize: 13, color: Colors.black54),
                                            ),
                                          ],
                                          if (isAssignedPending) ...[
                                            const SizedBox(height: 10),
                                            Row(
                                              children: [
                                                Expanded(
                                                  child: OutlinedButton(
                                                    onPressed: isActing ? null : () => _rejectOrder(orderId),
                                                    style: OutlinedButton.styleFrom(foregroundColor: Colors.red),
                                                    child: const Text('Reject'),
                                                  ),
                                                ),
                                                const SizedBox(width: 8),
                                                Expanded(
                                                  child: ElevatedButton(
                                                    onPressed: isActing ? null : () => _acceptOrder(orderId),
                                                    style: ElevatedButton.styleFrom(
                                                      backgroundColor: primaryPurple,
                                                      foregroundColor: Colors.white,
                                                    ),
                                                    child: isActing
                                                        ? const SizedBox(
                                                            height: 16,
                                                            width: 16,
                                                            child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                                                          )
                                                        : const Text('Accept'),
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ],
                                        ],
                                      ),
                                    ),
                                  );
                                },
                              ),
                            ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _filterChip(String label, String value) {
    final selected = _filter == value;
    return GestureDetector(
      onTap: () => setState(() => _filter = value),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: selected ? primaryPurple : Colors.white,
          borderRadius: BorderRadius.circular(20),
        ),
        child: Text(
          label,
          style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: selected ? Colors.white : Colors.black54),
        ),
      ),
    );
  }
}
