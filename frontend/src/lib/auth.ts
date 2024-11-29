import { jwtDecode } from "jwt-decode";

interface JWTClaims {
  github_auth?: boolean;
  exp: number;
}

export function isAuthenticatedToGitHub(): boolean {
  const token = localStorage.getItem("token");
  if (!token) return false;

  try {
    const claims = jwtDecode<JWTClaims>(token);
    return claims.github_auth === true;
  } catch (err) {
    return false;
  }
}
