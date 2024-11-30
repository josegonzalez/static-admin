import { useRouter } from "next/router";
import { useEffect } from "react";

export default function LogoutPage() {
  const router = useRouter();

  useEffect(() => {
    // Clear localStorage
    localStorage.clear();

    // Clear all cookies
    document.cookie.split(";").forEach((cookie) => {
      const eqPos = cookie.indexOf("=");
      const name = eqPos > -1 ? cookie.slice(0, eqPos).trim() : cookie.trim();
      document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`;
    });

    // Redirect to home page
    router.push("/");
  }, [router]);

  // Return null since this is just a utility page
  return null;
}

// Disable layout for this page
LogoutPage.getLayout = function getLayout(page: React.ReactNode) {
  return page;
};
