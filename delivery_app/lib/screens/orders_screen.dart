import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../services/location_service.dart';
import 'login_screen.dart';
import 'order_detail_screen.dart';

class OrdersScreen extends StatefulWidget {
  final void Function(int)? onSwitchTab;
  const OrdersScreen({super.key, this.onSwitchTab});

  @override
  State<OrdersScreen> createState() => _OrdersScreenState();
}

class _OrdersScreenState extends State<OrdersScreen> {
  List<dynamic> _orders = [];
  Map<String, dynamic>? _earnings;
  bool _loading = true;
  String? _error;
  String _filter = 'all';

  static const Color primaryPurple = Color(0xFF5B2A9E);
  static const Color pageBg = Color(0xFFF7F1FB);
  static const Color bannerBg = Color(0xFFEDE6F7);
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
      Map<String, dynamic>? earnings;
      try {
        earnings = await ApiService.getEarnings();
      } catch (_) {
        earnings = null;
      }
      setState(() {
        _orders = orders;
        _earnings = earnings;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Failed to load orders';
        _loading = false;
      });
    }
  }

  Future<void> _confirmLogout() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Logout'),
        content: const Text('Are you sure you want to logout?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('No'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Yes', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ApiService.clearToken();
      LocationService.stopTracking();
      if (!mounted) return;
      Navigator.pushAndRemoveUntil(
        context,
        MaterialPageRoute(builder: (_) => const LoginScreen()),
        (route) => false,
      );
    }
  }

  List<dynamic> get _filteredOrders {
    if (_filter == 'active') {
      return _orders.where((o) => (o['status'] ?? '') != 'delivered').toList();
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

  Color _iconBoxColor(String status) {
    return status == 'delivered' ? const Color(0xFFE1F5E6) : const Color(0xFFE7DEF6);
  }

  String _greeting() {
    final h = DateTime.now().hour;
    if (h < 12) return 'Good Morning, Partner!';
    if (h < 17) return 'Good Afternoon, Partner!';
    return 'Good Evening, Partner!';
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: pageBg,
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
                ? Center(child: Text(_error!))
                : RefreshIndicator(
                    onRefresh: _loadAll,
                    child: ListView(
                      padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
                      children: [
                        Row(
                          children: [
                            IconButton(
                              icon: const Icon(Icons.menu, color: Colors.black87),
                              onPressed: _confirmLogout,
                              tooltip: 'Logout',
                            ),
                            const Text(
                              'My Deliveries',
                              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.black87),
                            ),
                            const Spacer(),
                            IconButton(
                              icon: const Icon(Icons.notifications_none, color: Colors.black87),
                              onPressed: () {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  const SnackBar(content: Text('No new notifications')),
                                );
                              },
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(18),
                          decoration: BoxDecoration(
                            color: bannerBg,
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                _greeting(),
                                style: const TextStyle(
                                  fontSize: 17,
                                  fontWeight: FontWeight.bold,
                                  color: primaryPurple,
                                ),
                              ),
                              const SizedBox(height: 6),
                              const Text(
                                'Stay safe, deliver smiles.',
                                style: TextStyle(fontSize: 13, color: Colors.black54),
                              ),
                              const SizedBox(height: 14),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                decoration: BoxDecoration(
                                  color: Colors.white,
                                  borderRadius: BorderRadius.circular(20),
                                ),
                                child: const Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Icon(Icons.circle, color: Color(0xFF22C55E), size: 10),
                                    SizedBox(width: 6),
                                    Text('Online', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600)),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 16),
                        Row(
                          children: [
                            _FilterChip(
                              label: 'All Orders',
                              selected: _filter == 'all',
                              onTap: () => setState(() => _filter = 'all'),
                            ),
                            const SizedBox(width: 8),
                            _FilterChip(
                              label: 'Active',
                              selected: _filter == 'active',
                              onTap: () => setState(() => _filter = 'active'),
                            ),
                            const SizedBox(width: 8),
                            _FilterChip(
                              label: 'Completed',
                              selected: _filter == 'completed',
                              onTap: () => setState(() => _filter = 'completed'),
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        if (_filteredOrders.isEmpty)
                          const Padding(
                            padding: EdgeInsets.symmetric(vertical: 40),
                            child: Center(child: Text('No orders here')),
                          )
                        else
                          ...List.generate(_filteredOrders.length, (index) {
                            final order = _filteredOrders[index];
                            final status = (order['status'] ?? '').toString();
                            final city = order['address']?['city'];
                            final state = order['address']?['state'];
                            final location = [city, state].where((e) => e != null && e.toString().isNotEmpty).join(', ');
                            final time = _formatTime(order['created_at'] ?? order['updated_at']);
                            return Container(
                              margin: const EdgeInsets.only(bottom: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(color: cardBg, width: 1.2),
                              ),
                              child: Material(
                                color: Colors.transparent,
                                child: InkWell(
                                  borderRadius: BorderRadius.circular(16),
                                  onTap: () async {
                                    await Navigator.push(
                                      context,
                                      MaterialPageRoute(builder: (_) => OrderDetailScreen(order: order)),
                                    );
                                    _loadAll();
                                  },
                                  child: Padding(
                                    padding: const EdgeInsets.all(14),
                                    child: Row(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Container(
                                          width: 44,
                                          height: 44,
                                          decoration: BoxDecoration(
                                            color: _iconBoxColor(status),
                                            borderRadius: BorderRadius.circular(12),
                                          ),
                                          child: Icon(_statusIcon(status), color: primaryPurple, size: 22),
                                        ),
                                        const SizedBox(width: 12),
                                        Expanded(
                                          child: Column(
                                            crossAxisAlignment: CrossAxisAlignment.start,
                                            children: [
                                              Row(
                                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                                children: [
                                                  Text(
                                                    'Order #${order['id']}',
                                                    style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w700),
                                                  ),
                                                  Container(
                                                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                                    decoration: BoxDecoration(
                                                      color: _statusColor(status),
                                                      borderRadius: BorderRadius.circular(20),
                                                    ),
                                                    child: Text(
                                                      status.isEmpty ? '' : status[0].toUpperCase() + status.substring(1),
                                                      style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.w600),
                                                    ),
                                                  ),
                                                ],
                                              ),
                                              const SizedBox(height: 4),
                                              if (location.isNotEmpty)
                                                Text(location, style: const TextStyle(fontSize: 13, color: Colors.black54)),
                                              const SizedBox(height: 2),
                                              Row(
                                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                                children: [
                                                  Text(
                                                    '\u20B9${order['total_amount']} \u2022 ${order['payment_method']?.toString().toUpperCase() ?? ''}',
                                                    style: const TextStyle(fontSize: 13, color: Colors.black54),
                                                  ),
                                                  Row(
                                                    children: [
                                                      if (time != null)
                                                        Text(time, style: const TextStyle(fontSize: 12, color: Colors.black45)),
                                                      const SizedBox(width: 4),
                                                      const Icon(Icons.chevron_right, size: 18, color: Colors.black38),
                                                    ],
                                                  ),
                                                ],
                                              ),
                                            ],
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ),
                            );
                          }),
                        const SizedBox(height: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 18),
                          decoration: BoxDecoration(
                            color: bannerBg,
                            borderRadius: BorderRadius.circular(18),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  const Text("Today's Deliveries", style: TextStyle(fontSize: 12, color: primaryPurple, fontWeight: FontWeight.w600)),
                                  const SizedBox(height: 4),
                                  Text(
                                    '${_earnings?['today_deliveries'] ?? 0}',
                                    style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.black87),
                                  ),
                                ],
                              ),
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  const Text("Today's Earnings", style: TextStyle(fontSize: 12, color: primaryPurple, fontWeight: FontWeight.w600)),
                                  const SizedBox(height: 4),
                                  Text(
                                    '\u20B9${_earnings?['today_earnings'] ?? 0}',
                                    style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.black87),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
      ),
    );
  }
}

class _FilterChip extends StatelessWidget {
  final String label;
  final bool selected;
  final VoidCallback onTap;

  const _FilterChip({required this.label, required this.selected, required this.onTap});

  static const Color primaryPurple = Color(0xFF5B2A9E);
  static const Color chipUnselectedBg = Color(0xFFE9E1F5);

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        decoration: BoxDecoration(
          color: selected ? primaryPurple : chipUnselectedBg,
          borderRadius: BorderRadius.circular(20),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: selected ? Colors.white : Colors.black87,
            fontSize: 13,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }
}