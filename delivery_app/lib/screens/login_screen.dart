import 'package:flutter/material.dart';
import '../services/api_service.dart';
import 'otp_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _phoneController = TextEditingController();
  bool _loading = false;
  String? _error;

  static const Color primaryPurple = Color(0xFF5B2A9E);
  static const Color lightPurpleBg = Color(0xFFF3EEFB);

  Future<void> _sendOtp() async {
    final phone = _phoneController.text.trim();
    if (phone.length < 10) {
      setState(() => _error = 'Enter a valid phone number');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await ApiService.sendOtp(phone);
      setState(() => _loading = false);
      if (data['message'] == 'OTP sent successfully') {
        if (!mounted) return;
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => OtpScreen(phone: phone, testOtp: data['otp']?.toString()),
          ),
        );
      } else {
        setState(() => _error = data['error'] ?? 'Failed to send OTP');
      }
    } catch (e) {
      setState(() {
        _loading = false;
        _error = 'Network error. Please try again.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(
              width: double.infinity,
              color: lightPurpleBg,
              child: LayoutBuilder(
                builder: (context, constraints) {
                  final w = constraints.maxWidth;
                  final headerHeight = w * 1.05;
                  final imgWidth = w * 1.05;
                  return SizedBox(
                    height: headerHeight,
                    child: Stack(
                      clipBehavior: Clip.none,
                      children: [
                        Positioned(
                          bottom: 30,
                          right: -20,
                          width: imgWidth,
                          child: Image.asset(
                            'assets/images/delivery_scooter.png',
                            fit: BoxFit.contain,
                            alignment: Alignment.bottomRight,
                          ),
                        ),
                        const Positioned(
                          top: 24,
                          right: 24,
                          child: Icon(Icons.location_on, color: primaryPurple, size: 26),
                        ),
                        Positioned(
                          top: 60,
                          left: 24,
                          right: 24,
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              RichText(
                                text: const TextSpan(
                                  style: TextStyle(
                                    fontSize: 24,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.black87,
                                    height: 1.3,
                                  ),
                                  children: [
                                    TextSpan(text: 'Welcome Back,\n'),
                                    TextSpan(
                                      text: 'Delivery Partner!',
                                      style: TextStyle(color: primaryPurple),
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(height: 8),
                              const Text(
                                'Login to continue delivering\nhappiness',
                                style: TextStyle(fontSize: 14, color: Colors.black54),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
            Transform.translate(
              offset: const Offset(0, -28),
              child: Container(
                padding: const EdgeInsets.fromLTRB(24, 48, 24, 28),
                decoration: const BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.only(
                    topLeft: Radius.circular(28),
                    topRight: Radius.circular(28),
                  ),
                  boxShadow: [
                    BoxShadow(color: Color(0x14000000), blurRadius: 20, offset: Offset(0, -4)),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Center(
                      child: Container(
                        width: 64,
                        height: 64,
                        decoration: const BoxDecoration(
                          color: lightPurpleBg,
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(Icons.local_shipping, color: primaryPurple, size: 30),
                      ),
                    ),
                    const SizedBox(height: 16),
                    const Text(
                      'Delivery Partner Login',
                      textAlign: TextAlign.center,
                      style: TextStyle(fontSize: 19, fontWeight: FontWeight.bold, color: Colors.black87),
                    ),
                    const SizedBox(height: 6),
                    const Text(
                      'Enter your registered phone number',
                      textAlign: TextAlign.center,
                      style: TextStyle(fontSize: 13, color: Colors.black54),
                    ),
                    const SizedBox(height: 24),
                    Container(
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(color: primaryPurple.withOpacity(0.4)),
                      ),
                      child: Row(
                        children: [
                          const Padding(
                            padding: EdgeInsets.symmetric(horizontal: 14),
                            child: Text('+91', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
                          ),
                          const Icon(Icons.keyboard_arrow_down, color: Colors.black45, size: 18),
                          Container(width: 1, height: 28, color: Colors.black12, margin: const EdgeInsets.symmetric(horizontal: 10)),
                          const Icon(Icons.phone_outlined, color: primaryPurple, size: 18),
                          const SizedBox(width: 8),
                          Expanded(
                            child: TextField(
                              controller: _phoneController,
                              keyboardType: TextInputType.phone,
                              maxLength: 10,
                              decoration: const InputDecoration(
                                hintText: 'Phone number',
                                border: InputBorder.none,
                                counterText: '',
                                contentPadding: EdgeInsets.symmetric(vertical: 14),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    if (_error != null) ...[
                      const SizedBox(height: 10),
                      Text(_error!, style: const TextStyle(color: Colors.red, fontSize: 13)),
                    ],
                    const SizedBox(height: 20),
                    SizedBox(
                      height: 52,
                      child: ElevatedButton(
                        onPressed: _loading ? null : _sendOtp,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: primaryPurple,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(26)),
                          elevation: 0,
                        ),
                        child: _loading
                            ? const SizedBox(
                                height: 20,
                                width: 20,
                                child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                              )
                            : const Text(
                                'Send OTP',
                                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: Colors.white),
                              ),
                      ),
                    ),
                    const SizedBox(height: 24),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: const [
                        _TrustBadge(icon: Icons.verified_user_outlined, label: 'Safe & Secure\n100% Protected'),
                        _TrustBadge(icon: Icons.access_time_rounded, label: 'Quick Access\nLogin in seconds'),
                        _TrustBadge(icon: Icons.workspace_premium_outlined, label: 'Trusted Platform\nFor Delivery Partners'),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _TrustBadge extends StatelessWidget {
  final IconData icon;
  final String label;
  const _TrustBadge({required this.icon, required this.label});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Column(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: _LoginScreenState.lightPurpleBg,
              shape: BoxShape.circle,
            ),
            child: Icon(icon, color: _LoginScreenState.primaryPurple, size: 18),
          ),
          const SizedBox(height: 6),
          Text(
            label,
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 10, color: Colors.black54, height: 1.3),
          ),
        ],
      ),
    );
  }
}
