import { InvalidTokenError, jwtDecode } from "jwt-decode";

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
    if (err instanceof InvalidTokenError) {
      console.error(err);
      return false;
    }
    throw err;
  }
}
