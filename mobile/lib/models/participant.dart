/// One joined member of a campaign — powers the avatar-stack UI on the
/// campaign detail screen ("18 people joined").
class Participant {
  const Participant({
    required this.userId,
    required this.name,
    this.avatarUrl,
    required this.joinedAt,
  });

  final String userId;
  final String name;
  final String? avatarUrl;
  final String joinedAt;

  String get initials {
    final trimmed = name.trim();
    if (trimmed.isEmpty) return '?';
    final parts = trimmed.split(RegExp(r'\s+'));
    if (parts.length == 1) return parts.first[0].toUpperCase();
    return (parts.first[0] + parts.last[0]).toUpperCase();
  }

  factory Participant.fromJson(Map<String, dynamic> json) => Participant(
        userId: json['user_id'] as String? ?? '',
        name: json['name'] as String? ?? 'Member',
        avatarUrl: json['avatar_url'] as String?,
        joinedAt: json['joined_at'] as String? ?? '',
      );
}
