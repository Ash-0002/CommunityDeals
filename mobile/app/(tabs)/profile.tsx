// Profile tab — basic user info + settings

import {
  View,
  Text,
  ScrollView,
  StyleSheet,
  TouchableOpacity,
  useColorScheme,
} from "react-native";
import { useRouter } from "expo-router";
import { Colors } from "@/constants/Colors";

export default function ProfileScreen() {
  const scheme = useColorScheme() ?? "light";
  const C = Colors[scheme];
  const router = useRouter();

  const menuItems = [
    { icon: "📱", label: "Notifications" },
    { icon: "🏘️", label: "My Communities" },
    { icon: "📜", label: "Booking History" },
    { icon: "🔗", label: "Invite Friends" },
    { icon: "❓", label: "Help & Support" },
  ];

  return (
    <View style={[styles.root, { backgroundColor: C.bg }]}>
      <View style={[styles.header, { backgroundColor: C.card, borderBottomColor: C.border }]}>
        <Text style={[styles.headerTitle, { color: C.text }]}>Profile</Text>
      </View>

      <ScrollView contentContainerStyle={styles.scroll}>
        {/* Avatar */}
        <View style={[styles.avatarCard, { backgroundColor: C.card }]}>
          <View style={[styles.avatar, { backgroundColor: C.primary }]}>
            <Text style={styles.avatarChar}>R</Text>
          </View>
          <Text style={[styles.name, { color: C.text }]}>Rahul Mehta</Text>
          <Text style={[styles.phone, { color: C.muted }]}>+91 98765 43210</Text>
          <Text style={[styles.community, { color: C.primary }]}>Green Valley Society</Text>
        </View>

        {/* Stats */}
        <View style={styles.statRow}>
          {[
            { val: "2", label: "Active Deals" },
            { val: "₹600", label: "Saved" },
            { val: "5", label: "Completed" },
          ].map((s) => (
            <View key={s.label} style={[styles.statCard, { backgroundColor: C.card }]}>
              <Text style={[styles.statVal, { color: C.text }]}>{s.val}</Text>
              <Text style={[styles.statLabel, { color: C.muted }]}>{s.label}</Text>
            </View>
          ))}
        </View>

        {/* Menu */}
        <View style={[styles.menu, { backgroundColor: C.card }]}>
          {menuItems.map((item, i) => (
            <TouchableOpacity
              key={item.label}
              style={[
                styles.menuItem,
                i > 0 && { borderTopWidth: 1, borderTopColor: C.border },
              ]}
            >
              <Text style={{ fontSize: 18 }}>{item.icon}</Text>
              <Text style={[styles.menuLabel, { color: C.text }]}>{item.label}</Text>
              <Text style={[styles.menuArrow, { color: C.muted }]}>›</Text>
            </TouchableOpacity>
          ))}
        </View>

        <TouchableOpacity
          style={[styles.logoutBtn, { borderColor: "#EF4444" }]}
          onPress={() => router.replace("/auth")}
        >
          <Text style={styles.logoutText}>Log out</Text>
        </TouchableOpacity>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  root:   { flex: 1 },
  header: {
    paddingHorizontal: 20, paddingTop: 56, paddingBottom: 14,
    borderBottomWidth: 1,
  },
  headerTitle: { fontSize: 17, fontWeight: "800" },

  scroll:     { paddingHorizontal: 16, paddingTop: 20, paddingBottom: 40 },
  avatarCard: { borderRadius: 20, padding: 24, alignItems: "center", marginBottom: 14 },
  avatar: {
    width: 72, height: 72, borderRadius: 36,
    alignItems: "center", justifyContent: "center", marginBottom: 12,
  },
  avatarChar: { fontSize: 30, fontWeight: "900", color: "white" },
  name:       { fontSize: 20, fontWeight: "800", marginBottom: 2 },
  phone:      { fontSize: 13, fontWeight: "500", marginBottom: 6 },
  community:  { fontSize: 13, fontWeight: "700" },

  statRow:  { flexDirection: "row", gap: 10, marginBottom: 14 },
  statCard: { flex: 1, borderRadius: 14, padding: 14, alignItems: "center" },
  statVal:  { fontSize: 20, fontWeight: "800", marginBottom: 2 },
  statLabel:{ fontSize: 10, fontWeight: "600" },

  menu:     { borderRadius: 16, overflow: "hidden", marginBottom: 16 },
  menuItem: { flexDirection: "row", alignItems: "center", gap: 14, padding: 16 },
  menuLabel:{ flex: 1, fontSize: 14, fontWeight: "600" },
  menuArrow:{ fontSize: 20 },

  logoutBtn:  { borderWidth: 1.5, borderRadius: 14, paddingVertical: 14, alignItems: "center" },
  logoutText: { fontSize: 14, fontWeight: "700", color: "#EF4444" },
});
