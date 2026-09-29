class User {
  const User({
    required this.id,
    required this.phone,
    required this.name,
    this.email,
    this.avatarUrl,
    required this.role,
    required this.isVerified,
  });

  final String id;
  final String phone;
  final String name;
  final String? email;
  final String? avatarUrl;
  final String role;
  final bool isVerified;

  factory User.fromJson(Map<String, dynamic> json) => User(
        id: json['id'] as String? ?? '',
        phone: json['phone'] as String? ?? '',
        name: json['name'] as String? ?? '',
        email: json['email'] as String?,
        avatarUrl: json['avatar_url'] as String?,
        role: json['role'] as String? ?? 'member',
        isVerified: json['is_verified'] as bool? ?? false,
      );

  String get initials {
    final trimmed = name.trim();
    if (trimmed.isEmpty) return '?';
    final parts = trimmed.split(RegExp(r'\s+'));
    if (parts.length == 1) return parts.first[0].toUpperCase();
    return (parts.first[0] + parts.last[0]).toUpperCase();
  }
}
