// Next.js Edge Middleware — runs on every request before the page renders.
//
// Responsibilities:
//   1. Protect authenticated routes (/dashboard, /communities, /campaigns)
//   2. Redirect logged-in users away from /auth/* pages
//   3. Allow public routes through with no token check

import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const PROTECTED_PREFIXES = ["/dashboard", "/communities", "/campaigns", "/profile"];
const AUTH_PREFIXES      = ["/auth"];

export function middleware(request: NextRequest): NextResponse {
  const { pathname } = request.nextUrl;

  const accessToken  = request.cookies.get("cp_access_token")?.value;
  const refreshToken = request.cookies.get("cp_refresh_token")?.value;
  const hasSession   = !!(accessToken || refreshToken);

  // Redirect authenticated users away from login/OTP pages
  if (AUTH_PREFIXES.some((p) => pathname.startsWith(p)) && hasSession) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  // Redirect unauthenticated users away from protected pages
  if (PROTECTED_PREFIXES.some((p) => pathname.startsWith(p)) && !hasSession) {
    const loginUrl = new URL("/auth/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  // Run on all routes except static files, API routes, and Next internals
  matcher: ["/((?!_next/static|_next/image|favicon.ico|api/).*)"],
};
