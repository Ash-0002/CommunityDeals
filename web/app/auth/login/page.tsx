"use client";

// OTP Login flow:
//   Step 1 → enter phone → POST /auth/send-otp
//   Step 2 → enter 6-digit OTP → POST /auth/verify-otp → save tokens → redirect

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api, APIError } from "@/lib/api";
import { saveTokens, saveUser } from "@/lib/auth";

type Step = "phone" | "otp";

// useSearchParams() requires a Suspense boundary for static prerendering —
// see https://nextjs.org/docs/messages/missing-suspense-with-csr-bailout
export default function LoginPage() {
  return (
    <Suspense fallback={null}>
      <LoginForm />
    </Suspense>
  );
}

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirect = searchParams.get("redirect") ?? "/dashboard";

  const [step, setStep] = useState<Step>("phone");
  const [phone, setPhone] = useState("");
  const [otp, setOtp] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Normalise phone: strip spaces/dashes, add +91 if missing country code
  function normalisePhone(raw: string): string {
    const digits = raw.replace(/\D/g, "");
    if (digits.startsWith("91") && digits.length === 12) return `+${digits}`;
    if (digits.length === 10) return `+91${digits}`;
    return `+${digits}`;
  }

  async function handleSendOTP(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!phone.trim()) {
      setError("Please enter your phone number");
      return;
    }
    setLoading(true);
    try {
      await api.auth.sendOTP(normalisePhone(phone));
      setStep("otp");
    } catch (err) {
      setError(err instanceof APIError ? err.message : "Failed to send OTP. Try again.");
    } finally {
      setLoading(false);
    }
  }

  async function handleVerifyOTP(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (otp.length !== 6) {
      setError("Enter the 6-digit OTP");
      return;
    }
    setLoading(true);
    try {
      const data = await api.auth.verifyOTP(normalisePhone(phone), otp);
      saveTokens(data.access_token, data.refresh_token);
      saveUser(data.user);
      router.replace(redirect);
    } catch (err) {
      setError(err instanceof APIError ? err.message : "Incorrect OTP. Try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-surface px-4">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-lg bg-primary-600 text-base font-bold text-white">
            C
          </div>
          <h1 className="text-xl font-bold text-ink">CommunityDeals</h1>
          <p className="mt-1 text-sm text-muted">
            {step === "phone" ? "Enter your mobile number to continue" : `We sent a 6-digit code to ${phone}`}
          </p>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          {step === "phone" ? (
            <form onSubmit={handleSendOTP} className="space-y-4">
              <Input
                label="Mobile number"
                type="tel"
                placeholder="98765 43210"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                autoFocus
                inputMode="numeric"
                error={error}
              />
              <Button type="submit" loading={loading} fullWidth size="lg">
                Send OTP
              </Button>
            </form>
          ) : (
            <form onSubmit={handleVerifyOTP} className="space-y-4">
              <Input
                label="OTP"
                type="text"
                placeholder="• • • • • •"
                value={otp}
                onChange={(e) => setOtp(e.target.value.replace(/\D/g, "").slice(0, 6))}
                autoFocus
                inputMode="numeric"
                maxLength={6}
                className="text-center text-2xl tracking-widest"
                error={error}
                hint="Check your SMS or WhatsApp"
              />
              <Button type="submit" loading={loading} fullWidth size="lg">
                Verify &amp; continue
              </Button>
              <button
                type="button"
                onClick={() => {
                  setStep("phone");
                  setOtp("");
                  setError("");
                }}
                className="w-full text-center text-sm text-muted hover:text-ink"
              >
                ← Change number
              </button>
            </form>
          )}
        </div>

        {process.env.NODE_ENV === "development" && (
          <p className="mt-4 text-center text-xs text-muted">
            Dev mode: use OTP <strong>111111</strong>
          </p>
        )}
      </div>
    </main>
  );
}
