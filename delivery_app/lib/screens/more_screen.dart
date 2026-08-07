import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../services/location_service.dart';
import 'login_screen.dart';

class MoreScreen extends StatelessWidget {
  const MoreScreen({super.key});

  static const Color pageBg = Color(0xFFF7F1FB);
  static const Color primaryPurple = Color(0xFF5B2A9E);

  Future<void> _logout(BuildContext context) async {
    await ApiService.clearToken();
    LocationService.stopTracking();
    if (!context.mounted) return;
    Navigator.pushAndRemoveUntil(
      context,
      MaterialPageRoute(builder: (_) => const LoginScreen()),
      (route) => false,
    );
  }

  Widget _tile(BuildContext context, IconData icon, String title, VoidCallback onTap) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(14), border: Border.all(color: const Color(0xFFEDE6F7))),
      child: ListTile(
        leading: Icon(icon, color: primaryPurple),
        title: Text(title),
        trailing: const Icon(Icons.chevron_right, size: 18),
        onTap: onTap,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: pageBg,
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            const Text(
              'More',
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.black87),
            ),
            const SizedBox(height: 20),
            _tile(context, Icons.help_outline, 'Help & Support', () {
              ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Coming soon')));
            }),
            _tile(context, Icons.info_outline, 'About App', () {
              ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Coming soon')));
            }),
            _tile(context, Icons.settings_outlined, 'Settings', () {
              ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Coming soon')));
            }),
            _tile(context, Icons.logout, 'Logout', () => _logout(context)),
          ],
        ),
      ),
    );
  }
}
