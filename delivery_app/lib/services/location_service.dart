import 'dart:async';
import 'package:geolocator/geolocator.dart';
import 'api_service.dart';

class LocationService {
  static Timer? _timer;

  static Future<bool> requestPermission() async {
    bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) return false;

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) return false;
    }
    if (permission == LocationPermission.deniedForever) return false;
    return true;
  }

  static void startTracking() {
    _timer?.cancel();
    _sendCurrentLocation();
    _timer = Timer.periodic(const Duration(seconds: 20), (_) {
      _sendCurrentLocation();
    });
  }

  static void stopTracking() {
    _timer?.cancel();
    _timer = null;
  }

  static Future<void> _sendCurrentLocation() async {
    try {
      final hasPermission = await requestPermission();
      if (!hasPermission) return;
      final pos = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
      );
      await ApiService.updateLocation(pos.latitude, pos.longitude);
    } catch (_) {
      // Silently skip a failed location update; next timer tick will retry.
    }
  }
}
