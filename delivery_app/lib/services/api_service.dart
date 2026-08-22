import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

class ApiService {
  static const String baseUrl = 'https://ecommerce-backend-dd4u.onrender.com/api/v1';

  static Future<String?> getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString('token');
  }

  static Future<void> saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('token', token);
  }

  static Future<void> clearToken() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('token');
  }

  static Future<Map<String, String>> _headers({bool auth = true}) async {
    final headers = {'Content-Type': 'application/json'};
    if (auth) {
      final token = await getToken();
      if (token != null) headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  static Future<Map<String, dynamic>> sendOtp(String phone) async {
    final res = await http.post(
      Uri.parse('$baseUrl/delivery/send-otp'),
      headers: await _headers(auth: false),
      body: jsonEncode({'phone': phone}),
    );
    return jsonDecode(res.body);
  }

  static Future<Map<String, dynamic>> verifyOtp(String phone, String otp) async {
    final res = await http.post(
      Uri.parse('$baseUrl/delivery/verify-otp'),
      headers: await _headers(auth: false),
      body: jsonEncode({'phone': phone, 'otp': otp}),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode == 200 && data['token'] != null) {
      await saveToken(data['token']);
    }
    return data;
  }

  static Future<List<dynamic>> getMyDeliveries({String? status}) async {
    var url = '$baseUrl/delivery/orders';
    if (status != null) url += '?status=$status';
    final res = await http.get(Uri.parse(url), headers: await _headers());
    final data = jsonDecode(res.body);
    if (res.statusCode == 200) return data['orders'] ?? [];
    throw Exception(data['error'] ?? 'Failed to load orders');
  }

  static Future<Map<String, dynamic>> markShipped(int orderId) async {
    final res = await http.put(
      Uri.parse('$baseUrl/delivery/orders/$orderId/status'),
      headers: await _headers(),
      body: jsonEncode({'status': 'shipped'}),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode != 200) throw Exception(data['error'] ?? 'Failed to update status');
    return data;
  }

  static Future<Map<String, dynamic>> updateDeliveryStatus(
    int orderId,
    String status, {
    String? otp,
  }) async {
    final body = <String, dynamic>{'status': status};
    if (otp != null) body['otp'] = otp;
    final res = await http.put(
      Uri.parse('$baseUrl/delivery/orders/$orderId/delivery-status'),
      headers: await _headers(),
      body: jsonEncode(body),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode != 200) {
      throw Exception(data['error'] ?? 'Failed to update delivery status');
    }
    return data;
  }

  static Future<Map<String, dynamic>> confirmDelivery(int orderId) async {
    final res = await http.put(
      Uri.parse('$baseUrl/delivery/orders/$orderId/deliver'),
      headers: await _headers(),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode != 200) throw Exception(data['error'] ?? 'Failed to confirm delivery');
    return data;
  }

  static Future<Map<String, dynamic>> acceptAssignment(int orderId) async {
    final res = await http.put(
      Uri.parse('$baseUrl/delivery/orders/$orderId/accept'),
      headers: await _headers(),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode != 200) throw Exception(data['error'] ?? 'Failed to accept delivery');
    return data;
  }

  static Future<Map<String, dynamic>> rejectAssignment(int orderId, {String? reason}) async {
    final res = await http.put(
      Uri.parse('$baseUrl/delivery/orders/$orderId/reject'),
      headers: await _headers(),
      body: jsonEncode(
        (reason != null && reason.trim().isNotEmpty) ? {'reason': reason.trim()} : {},
      ),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode != 200) throw Exception(data['error'] ?? 'Failed to reject delivery');
    return data;
  }

  static Future<void> updateLocation(double lat, double lng) async {
    await http.put(
      Uri.parse('$baseUrl/delivery/location'),
      headers: await _headers(),
      body: jsonEncode({'lat': lat, 'lng': lng}),
    );
  }

  static Future<Map<String, dynamic>> getAvailability() async {
    final res = await http.get(Uri.parse('$baseUrl/delivery/availability'), headers: await _headers());
    final data = jsonDecode(res.body);
    if (res.statusCode != 200) throw Exception(data['error'] ?? 'Failed to load availability');
    return data;
  }

  static Future<Map<String, dynamic>> updateAvailability(bool isOnline) async {
    final res = await http.put(
      Uri.parse('$baseUrl/delivery/availability'),
      headers: await _headers(),
      body: jsonEncode({'status': isOnline ? 'online' : 'offline'}),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode != 200) throw Exception(data['error'] ?? 'Failed to update availability');
    return data;
  }

  static Future<Map<String, dynamic>> getEarnings() async {
    final res = await http.get(Uri.parse('$baseUrl/delivery/earnings'), headers: await _headers());
    final data = jsonDecode(res.body);
    if (res.statusCode != 200) throw Exception(data['error'] ?? 'Failed to load earnings');
    return data;
  }
}
