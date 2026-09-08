import 'dart:async';
import 'package:flutter/material.dart';
import '../services/api_service.dart';

class HomeScreen extends StatefulWidget {
  final void Function(int)? onSwitchTab;
  const HomeScreen({super.key, this.onSwitchTab});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  static const Color primaryPurple = Color(0xFF5B2A9E);
  static const Color pageBg = Color(0xFFF7F1FB);
  static const Color bannerBg = Color(0xFFEDE6F7);
  static const Color cardBg = Color(0xFFF3EDFA);

  bool? _isOnline;
  bool _togglingOnline = false;

  bool _loading = true;
  String? _error;

  Map<String, dynamic>? _earnings;
  Map<String, dynamic>? _codSummary;
  List<dynamic> _orders = [];

  @override
  void initState() {
    super.initState();
    _loadAll();
    _loadAvailability();
  }

  Future<void> _loadAll() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final orders = await ApiService.getMyDeliveries();
      Map<String, dynamic>? earnings;
      Map<String, dynamic>? codSummary;
      try {
        earnings = await ApiService.getEarnings();
      } catch (_) {
        earnings = null;
      }
      try {
        codSummary = await ApiService.getCODSummary();
      } catch (_) {
        codSummary = null;
      }
      setState(() {
        _orders = orders;
        _earnings = earnings;
        _codSummary = codSummary;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Failed to load dashboard';
        _loading = false;
      });
    }
  }

  Future<void> _loadAvailability() async {
    try {
      final data = await ApiService.getAvailability();
      if (mounted) setState(() => _isOnline = data['is_online'] == true);
    } catch (_) {
      // Badge just won't show a definite state yet.
    }
  }

  Future<void> _toggleOnline() async {
    final next = !(_isOnline ?? false);
    setState(() => _togglingOnline = true);
    try {
      final data = await ApiService.updateAvailability(next);
      if (mounted) setState(() => _isOnline = data['is_online'] == true);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString().replaceFirst('Exception: ', ''))),
        );
      }
    } finally {
      if (mounted) setState(() => _togglingOnline = false);
    }
  }

  String _greeting() {
    final h = DateTime.now().hour;
    if (h < 12) return 'Good Morning, Partner!';
    if (h < 17) return 'Good Afternoon, Partner!';
    return 'Good Evening, Partner!';
  }

  // The active order is whatever isn't yet delivered - there should
  // realistically be at most one at a time given the assignment flow,
  // so the first non-delivered order is shown.
  Map<String, dynamic>? get _activeOrder {
    for (final o in _orders) {
      final status = (o['status'] ?? '').toString();
      if (status != 'delivered' && status != 'cancelled') return o;
    }
    return null;
  }

  int get _todayCompletedCount {
    final now = DateTime.now();
    return _orders.where((o) {
      if ((o['status'] ?? '') != 'delivered') return false;
      final raw = o['delivered_at'] ?? o['updated_at'];
      if (raw == null) return false;
      try {
        final dt = DateTime.parse(raw.toString()).toLocal();
        return dt.year == now.year && dt.month == now.month && dt.day == now.day;
      } catch (_) {
        return false;
      }
    }).length;
  }

  @override
  Widget build(BuildContext context) {
    final todayEarnings = _earnings?['today_earnings'];
    final todayDeliveries = _earnings?['today_deliveries'] ?? _todayCompletedCount;
    final codCollectedToday = _codSummary?['today_collected'];
    final pendingSettlement = _codSummary?['pending_settlement'];

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
                            Expanded(
                              child: Text(
                                _greeting(),
                                style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.black87),
                              ),
                            ),
                            GestureDetector(
                              onTap: _togglingOnline ? null : _toggleOnline,
                              child: Container(
                                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                decoration: BoxDecoration(
                                  color: _isOnline == true ? const Color(0xFFE1F5E6) : const Color(0xFFFDEAEA),
                                  borderRadius: BorderRadius.circular(20),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    if (_togglingOnline)
                                      const SizedBox(
                                        width: 10,
                                        height: 10,
                                        child: CircularProgressIndicator(strokeWidth: 2),
                                      )
                                    else
                                      Container(
                                        width: 8,
                                        height: 8,
                                        decoration: BoxDecoration(
                                          shape: BoxShape.circle,
                                          color: _isOnline == true ? Colors.green : Colors.red,
                                        ),
                                      ),
                                    const SizedBox(width: 6),
                                    Text(
                                      _isOnline == true ? 'ONLINE' : 'OFFLINE',
                                      style: TextStyle(
                                        fontSize: 12,
                                        fontWeight: FontWeight.bold,
                                        color: _isOnline == true ? Colors.green[800] : Colors.red[800],
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),

                        // Today's summary
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(18),
                          decoration: BoxDecoration(color: bannerBg, borderRadius: BorderRadius.circular(20)),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                "Today's Summary",
                                style: TextStyle(fontWeight: FontWeight.w600, color: Colors.black54),
                              ),
                              const SizedBox(height: 14),
                              Row(
                                children: [
                                  Expanded(
                                    child: _statTile("Today's Deliveries", '$todayDeliveries'),
                                  ),
                                  Expanded(
                                    child: _statTile(
                                      "Today's Earnings",
                                      todayEarnings != null ? '₹${todayEarnings.toStringAsFixed(0)}' : '—',
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 10),
                              Row(
                                children: [
                                  Expanded(
                                    child: _statTile(
                                      'COD Collected',
                                      codCollectedToday != null ? '₹${codCollectedToday.toStringAsFixed(0)}' : '—',
                                    ),
                                  ),
                                  Expanded(
                                    child: _statTile(
                                      'Pending Settlement',
                                      pendingSettlement != null ? '₹${pendingSettlement.toStringAsFixed(0)}' : '—',
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 20),

                        // Active delivery
                        if (_activeOrder != null) ...[
                          const Text(
                            'ACTIVE DELIVERY',
                            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Colors.black54, letterSpacing: 0.5),
                          ),
                          const SizedBox(height: 8),
                          Container(
                            width: double.infinity,
                            padding: const EdgeInsets.all(16),
                            decoration: BoxDecoration(
                              color: cardBg,
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(color: primaryPurple.withValues(alpha: 0.15)),
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  children: [
                                    Text(
                                      'Order #${_activeOrder!['order_id'] ?? _activeOrder!['id']}',
                                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                                    ),
                                    const Spacer(),
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                      decoration: BoxDecoration(
                                        color: primaryPurple.withValues(alpha: 0.12),
                                        borderRadius: BorderRadius.circular(10),
                                      ),
                                      child: Text(
                                        (_activeOrder!['delivery_status'] ?? _activeOrder!['status'] ?? '')
                                            .toString()
                                            .replaceAll('_', ' ')
                                            .toUpperCase(),
                                        style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: primaryPurple),
                                      ),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 10),
                                if (_activeOrder!['payment_method'] == 'cod')
                                  Text(
                                    'COD: ₹${_activeOrder!['total_amount'] ?? '-'}',
                                    style: const TextStyle(color: Colors.black54, fontSize: 13),
                                  ),
                                const SizedBox(height: 12),
                                SizedBox(
                                  width: double.infinity,
                                  child: ElevatedButton(
                                    onPressed: () => widget.onSwitchTab?.call(1),
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: primaryPurple,
                                      foregroundColor: Colors.white,
                                      padding: const EdgeInsets.symmetric(vertical: 12),
                                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                                    ),
                                    child: const Text('View Delivery'),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ] else
                          Container(
                            width: double.infinity,
                            padding: const EdgeInsets.all(20),
                            decoration: BoxDecoration(color: cardBg, borderRadius: BorderRadius.circular(16)),
                            child: const Center(
                              child: Text('No active delivery right now', style: TextStyle(color: Colors.black45)),
                            ),
                          ),
                      ],
                    ),
                  ),
      ),
    );
  }

  Widget _statTile(String label, String value) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 4),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(14)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontSize: 11, color: Colors.black45)),
          const SizedBox(height: 4),
          Text(value, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.black87)),
        ],
      ),
    );
  }
}
