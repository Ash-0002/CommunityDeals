class Community {
  const Community({
    required this.id,
    required this.name,
    required this.description,
    required this.type,
    required this.status,
    this.city,
    this.state,
    this.pinCode,
    this.address,
    required this.requiresApproval,
    this.inviteCode,
    this.logoUrl,
    required this.memberCount,
    required this.createdById,
  });

  final String id;
  final String name;
  final String description;
  final String type;
  final String status;
  final String? city;
  final String? state;
  final String? pinCode;
  final String? address;
  final bool requiresApproval;
  final String? inviteCode;
  final String? logoUrl;
  final int memberCount;
  final String createdById;

  factory Community.fromJson(Map<String, dynamic> json) => Community(
        id: json['id'] as String? ?? '',
        name: json['name'] as String? ?? '',
        description: json['description'] as String? ?? '',
        type: json['type'] as String? ?? '',
        status: json['status'] as String? ?? '',
        city: json['city'] as String?,
        state: json['state'] as String?,
        pinCode: json['pin_code'] as String?,
        address: json['address'] as String?,
        requiresApproval: json['requires_approval'] as bool? ?? false,
        inviteCode: json['invite_code'] as String?,
        logoUrl: json['logo_url'] as String?,
        memberCount: (json['member_count'] as num?)?.toInt() ?? 0,
        createdById: json['created_by_id'] as String? ?? '',
      );
}
