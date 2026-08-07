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
      final data = await ApiService.markShipped(_order['id']);
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

  Future<void> _confirmDelivery() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await ApiService.confirmDelivery(_order['id']);
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
    final address = _order['address'] ?? {};
    final items = (_order['items'] ?? []) as List<dynamic>;
    final status = _order['status'] ?? '';

    return Scaffold(
      appBar: AppBar(title: Text('Order #${_order['id']}')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    address['full_name'] ?? '',
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                  const SizedBox(height: 4),
                  Text(address['phone'] ?? ''),
                  const SizedBox(height: 8),
                  Text('${address['line1'] ?? ''}, ${address['line2'] ?? ''}'),
                  Text('${address['city'] ?? ''}, ${address['state'] ?? ''} - ${address['pincode'] ?? ''}'),
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
                  ...items.map((item) {
                    final product = item['product'] ?? {};
                    return Padding(
                      padding: const EdgeInsets.symmetric(vertical: 4),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text('${product['name'] ?? ''} x${item['quantity']}'),
                          Text('₹${item['price']}'),
                        ],
                      ),
                    );
                  }),
                  const Divider(),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text('Total', style: TextStyle(fontWeight: FontWeight.bold)),
                      Text(
                        '₹${_order['total_amount']}',
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
          if (status == 'confirmed')
            ElevatedButton(
              onPressed: _loading ? null : _markShipped,
              style: ElevatedButton.styleFrom(padding: const EdgeInsets.all(16)),
              child: _loading
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('Mark as Shipped (Picked Up)'),
            )
          else if (status == 'shipped')
            ElevatedButton(
              onPressed: _loading ? null : _confirmDelivery,
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.all(16),
                backgroundColor: Colors.green,
              ),
              child: _loading
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('Confirm Delivery'),
            )
          else if (status == 'delivered')
            const Center(
              child: Text('✅ Delivered', style: TextStyle(color: Colors.green, fontSize: 18)),
            ),
        ],
      ),
    );
  }
}
